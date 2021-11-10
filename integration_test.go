package viagh

import (
	"context"
	"flag"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v39/github"
	"github.com/k1LoW/go-github-client/v39/factory"
)

var integration = flag.Bool("integration", false, "run integration tests")

func TestGet(t *testing.T) {
	if !*integration {
		t.Skip()
	}
	ctx := context.Background()
	client, err := NewHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	gh := github.NewClient(client)
	g, err := factory.NewGithubClient()
	if err != nil {
		t.Fatal(err)
	}

	{
		// GET users/k1LoW/
		r1, _, err := gh.Users.Get(ctx, "k1LoW")
		if err != nil {
			t.Fatal(err)
		}
		r2, _, err := g.Users.Get(ctx, "k1LoW")
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(r1, r2, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}

	{
		// GET search/repositories?page=1&per_page=10&q=specinfra&sort=created
		r1, _, err := gh.Search.Repositories(ctx, "specinfra", &github.SearchOptions{
			Sort: "created",
			ListOptions: github.ListOptions{
				Page:    1,
				PerPage: 10,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		r2, _, err := g.Search.Repositories(ctx, "specinfra", &github.SearchOptions{
			Sort: "created",
			ListOptions: github.ListOptions{
				Page:    1,
				PerPage: 10,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(r1, r2, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func TestPost(t *testing.T) {
	if !*integration {
		t.Skip()
	}
	ctx := context.Background()
	client, err := NewHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	gh := github.NewClient(client)
	g, err := factory.NewGithubClient()
	if err != nil {
		t.Fatal(err)
	}
	r1, _, err := gh.Issues.Create(ctx, "k1LoW", "private-repo-for-test", &github.IssueRequest{
		Title: github.String("Hello via gh"),
		Body: github.String(`Hello=
via gh`),
	})
	if err != nil {
		t.Fatal(err)
	}
	r2, _, err := g.Issues.Create(ctx, "k1LoW", "private-repo-for-test", &github.IssueRequest{
		Title: github.String("Hello via go-github"),
		Body: github.String(`Hello=
via go-github`),
	})
	if err != nil {
		t.Fatal(err)
	}

	if r1.GetNumber()+1 != r2.GetNumber() {
		t.Errorf("got %v %v", r1.GetNumber(), r2.GetNumber())
	}

}

func TestPatch(t *testing.T) {
	if !*integration {
		t.Skip()
	}
	ctx := context.Background()
	client, err := NewHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	gh := github.NewClient(client)
	g, err := factory.NewGithubClient()
	if err != nil {
		t.Fatal(err)
	}
	r1, _, err := gh.Issues.Edit(ctx, "k1LoW", "private-repo-for-test", 1, &github.IssueRequest{
		Title: github.String("Hello via gh"),
		Body: github.String(`Hello=
via gh`),
	})
	if err != nil {
		t.Fatal(err)
	}
	r2, _, err := g.Issues.Edit(ctx, "k1LoW", "private-repo-for-test", 1, &github.IssueRequest{
		Title: github.String("Hello via go-github"),
		Body: github.String(`Hello=
via go-github`),
	})
	if err != nil {
		t.Fatal(err)
	}

	if r1.GetNumber() != r2.GetNumber() {
		t.Errorf("got %v %v", r1.GetNumber(), r2.GetNumber())
	}
}

func TestDelete(t *testing.T) {
	if !*integration {
		t.Skip()
	}
	ctx := context.Background()
	client, err := NewHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	gh := github.NewClient(client)
	g, err := factory.NewGithubClient()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = g.Issues.AddLabelsToIssue(ctx, "k1LoW", "private-repo-for-test", 1, []string{"duplicate", "enhancement"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err = gh.Issues.AddLabelsToIssue(ctx, "k1LoW", "private-repo-for-test", 1, []string{"bug", "documentation"}); err != nil {
		t.Fatal(err)
	}

	if _, err := gh.Issues.DeleteLabel(ctx, "k1LoW", "private-repo-for-test", "bug"); err != nil {
		t.Fatal(err)
	}

	if _, err := g.Issues.DeleteLabel(ctx, "k1LoW", "private-repo-for-test", "duplicate"); err != nil {
		t.Fatal(err)
	}

	i, _, err := gh.Issues.Get(ctx, "k1LoW", "private-repo-for-test", 1)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"documentation", "enhancement"}
	got := []string{}
	for _, l := range i.Labels {
		got = append(got, l.GetName())
	}
	if diff := cmp.Diff(got, want, nil); diff != "" {
		t.Errorf("%s", diff)
	}
}
