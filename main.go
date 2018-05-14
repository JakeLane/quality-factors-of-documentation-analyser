package main

import (
	"context"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Minimum number of stargazers (stargazers_count) for a repository
var minimumStargazes = 150

func getSampleRepositories(ctx context.Context, client *github.Client) []*github.Repository {
	repos, response, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{})
	if err == nil {
		log.Fatal(err, response)
	}

	ret := []*github.Repository{}
	for i := range repos {
		if *repos[i].StargazersCount >= minimumStargazes {
			ret = append(repos, repos[i])
		}
	}

	return ret
}

func getReadmes(ctx context.Context, client *github.Client) {
	// readme, response, err := client.Repositories.GetReadme(ctx, "golang", "go", &github.RepositoryContentGetOptions{})
}

func main() {
	accessToken := os.Getenv("QoDA_GITHUB_PAT")
	if accessToken == "" {
		log.Fatal("GitHub Personal Access token was not defined")
		os.Exit(1)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// Get a sample of repositories
	repos := getSampleRepositories(ctx, client)
	println(repos)

	// Get all READMEs from repositories
	// getReadmes(ctx, client)
}
