package viagh

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/cli/safeexec"
	"github.com/goccy/go-json"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

const defaultPerPage = 30

var statusRe = regexp.MustCompile(`\(HTTP (\d\d\d)\)`)

// NewHTTPClient returns a *http.Client that makes API requests via the `gh` command
func NewHTTPClient() (*http.Client, error) {
	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		return nil, err
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			args    []string
			stdin   string
			page    int
			perPage int
			u       *url.URL
		)
		switch r.Method {
		case http.MethodGet, http.MethodDelete:
			page, perPage, u, err = parseURLAndSeparatePageAndPerPage(r.RequestURI)
			ep := strings.TrimPrefix(u.RequestURI(), "/")
			args = []string{"api", "-X", r.Method, ep}
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if page > 0 {
				args = append(args, "--paginate")
			}
		case http.MethodPost, http.MethodPatch:
			ep := strings.TrimPrefix(r.RequestURI, "/")
			args = []string{"api", "-X", r.Method, ep}
			b := r.Body
			defer func() { _ = b.Close() }()
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, b); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var body interface{}
			if err := json.Unmarshal(buf.Bytes(), &body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			switch kv := body.(type) {
			case map[string]interface{}:
				for k, v := range kv {
					args = append(args, "-f", fmt.Sprintf("%s=%s", k, v.(string)))
				}
			case []interface{}:
				// ref: https://github.com/cli/cli/issues/1484
				stdin = buf.String()
				args = append(args, "--input", "-")
			default:
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			// Why does the server return 400?
			// Because this library is essentially a client
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cmd := exec.Command(ghBin, args...) //nolint:gosec
		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if stdin != "" {
			cmd.Stdin = strings.NewReader(stdin)
		}
		if err := cmd.Run(); err != nil {
			m := statusRe.FindSubmatch(stderr.Bytes())
			if len(m) != 2 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			s, err := strconv.Atoi(string(m[1]))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(s)
		}

		b := stdout.Bytes()
		if page > 0 {
			pages := bytes.Split(b, []byte("}][{"))
			if len(pages) > 1 {
				if page == 1 {
					b = append(pages[0], []byte("}]")...)
				} else if page == len(pages) {
					b = append([]byte("[{"), pages[page-1]...)
				} else {
					b = append([]byte("[{"), pages[page-1]...)
					b = append(b, []byte("}]")...)
				}
			}
			// TODO: cache response
			if page < len(pages) {
				q := u.Query()
				q.Del("page")
				q.Del("per_page")
				q.Add("page", strconv.Itoa(page+1))
				q.Add("per_page", strconv.Itoa(perPage))
				u.RawQuery = q.Encode()
				w.Header().Set("link", fmt.Sprintf("<%s>; rel=\"next\"", u.RequestURI()))
			}
		}

		_, _ = w.Write(b)
	})

	mockServer := httptest.NewServer(h)

	c := mockServer.Client()

	c.Transport = &mock.EnforceHostRoundTripper{
		Host:                 mockServer.URL,
		UpstreamRoundTripper: mockServer.Client().Transport,
	}

	return c, nil
}

func parseURLAndSeparatePageAndPerPage(reqURI string) (int, int, *url.URL, error) {
	var page, perPage int
	u, err := url.Parse(reqURI)
	if err != nil {
		return 0, 0, nil, err
	}
	q := u.Query()
	sp, ok := q["page"]
	if ok && len(sp) > 0 {
		q.Del("page")
		page, err = strconv.Atoi(sp[0])
		if err != nil {
			return 0, 0, nil, err
		}
		spp, ok := q["per_page"]
		if ok && len(spp) > 0 {
			perPage, err = strconv.Atoi(spp[0])
			if err != nil {
				return 0, 0, nil, err
			}
		}
	}
	u.RawQuery = q.Encode()
	if page > 0 && perPage == 0 {
		perPage = defaultPerPage
	}
	return page, perPage, u, nil
}
