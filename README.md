# viagh

:octocat: `viagh.NewHTTPClient` returns a `*http.Client` that makes API requests via the `gh` command.

## Why viagh?

When writing a GitHub CLI extension, the extension needs to exec the gh command internally to execute API requests with credentials.

By using `http.Client` that executes API requests via the `gh` command, you can use existing useful packages such as [go-github](https://github.com/google/go-github) without modification.

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
