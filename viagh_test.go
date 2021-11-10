package viagh

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v39/github"
)

func TestRequest(t *testing.T) {
	ctx := context.Background()
	client, err := NewHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	gh := github.NewClient(client)

	{
		// GET users/k1LoW/
		u, _, err := gh.Users.Get(ctx, "k1LoW")
		if err != nil {
			t.Fatal(err)
		}
		got := u.GetLogin()
		if want := "k1LoW"; got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}

	{
		// GET search/repositories?page=1&per_page=10&q=specinfra&sort=created
		res, _, err := gh.Search.Repositories(ctx, "specinfra", &github.SearchOptions{
			Sort: "created",
			ListOptions: github.ListOptions{
				Page:    1,
				PerPage: 10,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		got := len(res.Repositories)
		if want := 10; got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}

	{
		c, _, err := gh.Issues.CreateComment(ctx, "k1LoW", "viagh", 1, &github.IssueComment{
			Body: github.String("Add comment via `gh`"),
		})
		if err != nil {
			t.Fatal(err)
		}

		got := c.GetBody()
		if want := "Add comment via `gh`"; got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}

		if _, err := gh.Issues.DeleteComment(ctx, "k1LoW", "viagh", c.GetID()); err != nil {
			t.Fatal(err)
		}

		if _, _, err = gh.Issues.AddLabelsToIssue(ctx, "k1LoW", "viagh", 1, []string{"duplicate", "enhancement"}); err != nil {
			t.Fatal(err)
		}

		if _, err := gh.Issues.DeleteLabel(ctx, "k1LoW", "viagh", "duplicate"); err != nil {
			t.Fatal(err)
		}

	}

	{
		i, _, err := gh.Issues.Get(ctx, "k1LoW", "viagh", 1)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"enhancement", "wontfix"}
		got := []string{}
		for _, l := range i.Labels {
			got = append(got, l.GetName())
		}
		sort.Slice(got, func(i, j int) bool {
			return got[i] < got[j]
		})
		if diff := cmp.Diff(got, want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
		if _, err := gh.Issues.DeleteLabel(ctx, "k1LoW", "viagh", "enhancement"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRequestError(t *testing.T) {
	client, err := NewHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	gh := github.NewClient(client)

	{
		// GET users/konohitohaimasen/
		_, res, err := gh.Users.Get(ctx, "konohitohaimasen")
		if err == nil {
			t.Error("should be not found")
		}
		got := res.StatusCode
		if want := 404; got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}
}
