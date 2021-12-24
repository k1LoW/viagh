package viagh

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v39/github"
)

const (
	testOwner = "k1LoW"
	testRepo  = "viagh"
)

func Example() {
	ctx := context.Background()
	client, _ := NewHTTPClient()
	gh := github.NewClient(client)

	u, _, _ := gh.Users.Get(ctx, testOwner)
	fmt.Println(u.GetLogin())
	// Unordered output:
	// k1LoW
}

func TestRequest(t *testing.T) {
	ctx := context.Background()
	client, err := NewHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	gh := github.NewClient(client)

	{
		// GET users/k1LoW/
		u, _, err := gh.Users.Get(ctx, testOwner)
		if err != nil {
			t.Fatal(err)
		}
		got := u.GetLogin()
		if want := testOwner; got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}

	{
		// GET search/repositories?page=1&per_page=5&q=specinfra&sort=created
		res, _, err := gh.Search.Repositories(ctx, "specinfra", &github.SearchOptions{
			Sort: "created",
			ListOptions: github.ListOptions{
				Page:    1,
				PerPage: 5,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		got := len(res.Repositories)
		if want := 5; got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}

	{
		b, _, err := gh.Repositories.GetBranch(ctx, testOwner, testRepo, "main", false)
		if err != nil {
			t.Fatal(err)
		}
		tree, _, err := gh.Git.GetTree(ctx, testOwner, testRepo, b.GetCommit().GetSHA(), true)
		if err != nil {
			t.Fatal(err)
		}
		if len(tree.Entries) == 0 {
			t.Error("invalid")
		}
	}

	{
		c, _, err := gh.Issues.CreateComment(ctx, testOwner, testRepo, 1, &github.IssueComment{
			Body: github.String("Add comment via `gh`"),
		})
		if err != nil {
			t.Fatal(err)
		}

		got := c.GetBody()
		if want := "Add comment via `gh`"; got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}

		if _, err := gh.Issues.DeleteComment(ctx, testOwner, testRepo, c.GetID()); err != nil {
			t.Fatal(err)
		}

		if _, _, err = gh.Issues.AddLabelsToIssue(ctx, testOwner, testRepo, 1, []string{"duplicate", "enhancement"}); err != nil {
			t.Fatal(err)
		}

		if _, err := gh.Issues.DeleteLabel(ctx, testOwner, testRepo, "duplicate"); err != nil {
			t.Fatal(err)
		}

	}

	{
		i, _, err := gh.Issues.Get(ctx, testOwner, testRepo, 1)
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
		if _, err := gh.Issues.DeleteLabel(ctx, testOwner, testRepo, "enhancement"); err != nil {
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

func TestRequestWithPagination(t *testing.T) {
	ctx := context.Background()
	client, err := NewHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	gh := github.NewClient(client)

	{
		var got string
		page := 1
	L:
		for {
			tags, res, err := gh.Repositories.ListTags(ctx, "golang", "go", &github.ListOptions{
				Page:    page,
				PerPage: 100,
			})
			if err != nil {
				t.Fatal(err)
			}
			for _, t := range tags {
				if t.GetName() == "go1" {
					got = t.GetCommit().GetSHA()
					break L
				}
			}
			if res.NextPage == 0 {
				break
			}
			page += 1
		}

		if want := "6174b5e21e73714c63061e66efdbe180e1c5491d"; got != want {
			t.Errorf("got %v\nwant %v", got, want)
		}
	}
}
