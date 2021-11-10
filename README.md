# viagh

:octocat: `viagh.NewHTTPClient` returns a `*http.Client` that makes API requests via the `gh` command.

## Usage

``` go
package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/v39/github"
	"github.com/k1LoW/viagh"
)

func main() {
	ctx := context.Background()
	client, _ := viagh.NewHTTPClient()
	gh := github.NewClient(client)

	u, _, _ := gh.Users.Get(ctx, "k1LoW")
	fmt.Println(u.GetLogin())
}
```
