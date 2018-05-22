package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Minimum number of stargazers (stargazers_count) for a repository
var minimumStargazes = 150

// Sample size
var sampleSize = 10000

// File to output data to
var outputFilename = "output.json"

func getRepositories(ctx context.Context, client *github.Client) []github.Repository {
	var repos []github.Repository
	opt := &github.SearchOptions{
		Sort: "updated",
	}

	for len(repos) < sampleSize {
		result, response, err := client.Search.Repositories(ctx, fmt.Sprintf("stars:>=%d", minimumStargazes), opt)

		if _, ok := err.(*github.RateLimitError); ok {
			log.Println("Hit rate limit, retrying in 10 seconds")
			time.Sleep(10)
			continue
		}

		if err != nil {
			log.Fatal("Failed to load repos:", err)
		}

		repos = append(repos, result.Repositories...)

		if response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage

		log.Println(fmt.Sprintf("Retrieved %d repositories, %d requests left", len(result.Repositories), response.Remaining))
	}

	return repos
}

func filterToMarkdown(tree github.Tree) ([]github.TreeEntry, int) {
	var entries []github.TreeEntry
	totalBytes := 0

	for i := range tree.Entries {
		if strings.HasSuffix(*tree.Entries[i].Path, ".md") {
			totalBytes += *tree.Entries[i].Size
			entries = append(entries, tree.Entries[i])
		}
	}

	return entries, totalBytes
}

// Project is repo with readme
type Project struct {
	RepoInfo    github.Repository
	Readme      github.RepositoryContent
	TreeEntries []github.TreeEntry
	TotalBytes  int
}

func getFileData(ctx context.Context, client *github.Client, repos []github.Repository) []Project {
	var projects []Project

	for i := range repos {
		readme, response, err := client.Repositories.GetReadme(ctx, *(*repos[i].Owner).Login, *repos[i].Name, &github.RepositoryContentGetOptions{})

		if response.StatusCode == 404 {
			log.Println("No readme for", *repos[i].Name, "skipping")
			continue
		} else if err != nil {
			log.Println("Failed to load readme:", err)
			continue
		}

		latestCommit, response, err := client.Repositories.GetCommit(ctx, *(*repos[i].Owner).Login, *repos[i].Name, *repos[i].DefaultBranch)

		if err != nil {
			log.Println("Failed to load latest commit:", err)
			continue
		}

		tree, response, err := client.Git.GetTree(ctx, *(*repos[i].Owner).Login, *repos[i].Name, *latestCommit.SHA, true)

		if err != nil {
			log.Println("Failed to load latest commit:", err)
			continue
		}

		entries, totalBytes := filterToMarkdown(*tree)

		projects = append(projects, Project{
			RepoInfo:    repos[i],
			Readme:      *readme,
			TreeEntries: entries,
			TotalBytes:  totalBytes,
		})

		log.Println(fmt.Sprintf("Retrieved language size for %s", *repos[i].Name))
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
	projects := getFileData(ctx, client, repos)

	// Generate json
	json, err := json.Marshal(projects)
	if err != nil {
		log.Fatal("Failed to generate json:", err)
	}
	log.Println("Saved json to", outputFilename)

	// Save to file
	err = ioutil.WriteFile(outputFilename, json, 0644)
}
