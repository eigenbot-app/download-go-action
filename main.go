package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v50/github"
)

func main() {
	var (
		owner, repo, name, token string
	)
	flag.StringVar(&owner, "owner", "", "")
	flag.StringVar(&repo, "repo", "", "")
	flag.StringVar(&name, "name", "", "")
	flag.StringVar(&token, "token", "", "")
	flag.Parse()

	if defaultToken := os.Getenv("GH_TOKEN"); token == "" && defaultToken != "" {
		token = defaultToken
	}

	ctx := context.Background()
	gh := github.NewTokenClient(ctx, token)
	opts := &github.ListOptions{PerPage: 100}

	var candidates []*github.RepositoryRelease
	for {
		rels, resp, err := gh.Repositories.ListReleases(ctx, owner, repo, opts)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "list releases: %s\n", err)
			os.Exit(1)
		}

		for _, rel := range rels {
			if strings.HasPrefix(rel.GetTagName(), name) && !rel.GetDraft() {
				candidates = append(candidates, rel)
			}
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	var mostRecent *github.RepositoryRelease
	for _, c := range candidates {
		if mostRecent == nil {
			mostRecent = c
			continue
		}
		if c.GetPublishedAt().After(c.GetPublishedAt().Time) {
			mostRecent = c
		}
	}

	if mostRecent == nil {
		_, _ = fmt.Fprintf(os.Stderr, "no release matching tag prefix `%s` was found\n", name)
		os.Exit(1)
	}

	fmt.Printf("Downloading most recent release of `%s` at tag `%s`...\n", name, mostRecent.GetTagName())

	cmd := exec.Command(
		"gh", "release", "download",
		mostRecent.GetTagName(),
		"--repo", owner+"/"+repo,
		"--pattern", name,
		"--clobber",
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = []string{
		fmt.Sprintf("GH_TOKEN=%s", token),
	}
	if err := cmd.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "download release: %s\n", err)
		os.Exit(1)
	}

	if err := os.Chmod(name, 0755); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "chmod: %s\n", err)
		os.Exit(1)
	}
}
