package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Minimum number of stargazers (stargazers_count) for a repository
var minimumStargazes = 150

func getRepositories(ctx context.Context, client *github.Client) []github.Repository {
	var repos []github.Repository

	for true {
		result, response, err := client.Search.Repositories(ctx, fmt.Sprintf("stars:>=%d", minimumStargazes), &github.SearchOptions{})

		if _, ok := err.(*github.RateLimitError); ok {
			log.Println("Hit rate limit")
			break
		}

		if err != nil {
			log.Fatal("Failed to load repos: ", err)
		}

		repos = append(repos, result.Repositories...)

		log.Println(fmt.Sprintf("Retrieved %d repositories, %d requests left", len(result.Repositories), response.Remaining))
		break
	}

	return repos
}

// Project is repo with readme
type Project struct {
	Payload github.Repository
	Readme  string
}

func getReadmes(ctx context.Context, client *github.Client, repos []github.Repository) []Project {
	var projects []Project
	for i := range repos {
		readme, _, err := client.Repositories.GetReadme(ctx, *(*repos[i].Owner).Login, *repos[i].Name, &github.RepositoryContentGetOptions{})

		if err != nil {
			log.Fatal("Failed to load readme: ", err)
		}

		projects = append(projects, Project{
			Payload: repos[i],
			Readme:  *readme.Content,
		})

		log.Println(fmt.Sprintf("Retrieved README for %s", *repos[i].Name))
	}

	return projects
}

func main() {
	accessToken := os.Getenv("QoDA_GITHUB_PAT")
	if accessToken == "" {
		log.Fatal("GitHub Personal Access token was not defined")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// Get repositories
	repos := getRepositories(ctx, client)

	// Get all READMEs from repositories
	readmes := getReadmes(ctx, client, repos)

	// Generate json
	json, err := json.Marshal(readmes)
	if err != nil {
		log.Fatal("Failed to generate json: ", err)
	}

	// Save to file
	err = ioutil.WriteFile("output.json", json, 0644)
}
