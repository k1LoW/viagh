package viagh

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/cli/safeexec"
	"github.com/goccy/go-json"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

var statusRe = regexp.MustCompile(`\(HTTP (\d\d\d)\)`)

// NewHTTPClient returns a *http.Client that makes API requests via the `gh` command
func NewHTTPClient() (*http.Client, error) {
	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		return nil, err
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var args []string
		var stdin string
		switch r.Method {
		case http.MethodGet, http.MethodDelete:
			ep := strings.TrimPrefix(r.RequestURI, "/")
			args = []string{"api", "-X", r.Method, ep}
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
		_, _ = w.Write(stdout.Bytes())
	})

	mockServer := httptest.NewServer(h)

	c := mockServer.Client()

	c.Transport = &mock.EnforceHostRoundTripper{
		Host:                 mockServer.URL,
		UpstreamRoundTripper: mockServer.Client().Transport,
	}

	return c, nil
}
