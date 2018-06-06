package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/BluntSporks/readability"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Minimum number of stargazers (stargazers_count) for a repository
var minimumStargazes = 150

// Sample size
var sampleSize = 1000

/*
 * Supported documentation extensions
 * Markdown (.md): docpress, GitBook, Slate, ReadTheDocs, docsify, docpress, DocumentUp, docbox
 * AsciiDoc (.adoc): GitBook
 * reStructuredText (.rst): ReadTheDocs
 */
var docsExtensions = []string{".md", ".rst", ".adoc"}

// File to output data to
var outputFilename = "output.json"

func getRepositories(ctx context.Context, client *github.Client) []github.Repository {
	var repos []github.Repository
	opt := &github.SearchOptions{
		Sort: "updated",
		ListOptions: github.ListOptions{
			PerPage: sampleSize,
		},
	}

	for len(repos) < sampleSize {
		opt.ListOptions.PerPage = sampleSize - len(repos)
		result, response, err := client.Search.Repositories(ctx, fmt.Sprintf("stars:>=%d", minimumStargazes), opt)

		if _, ok := err.(*github.RateLimitError); ok {
			log.Println("Hit rate limit, retrying in 10 seconds")
			time.Sleep(10 * time.Second)
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

		log.Println(
			fmt.Sprintf("Retrieved %d repositories, %d requests left, %d repositories left",
				len(result.Repositories),
				response.Remaining,
				sampleSize-len(repos),
			),
		)
	}

	return repos
}

func filterToDocs(tree github.Tree) ([]github.TreeEntry, int) {
	var entries []github.TreeEntry
	totalBytes := 0

	for i := range tree.Entries {
		if tree.Entries[i].Size == nil {
			continue
		}

		for j := range docsExtensions {
			if strings.HasSuffix(strings.ToLower(*tree.Entries[i].Path), docsExtensions[j]) {
				totalBytes += *tree.Entries[i].Size
				entries = append(entries, tree.Entries[i])
				break
			}
		}

	}

	return entries, totalBytes
}

func getReadabilityScore(ctx context.Context, client *github.Client, owner string, repo string, files []github.TreeEntry) float64 {
	totalScore := float64(0)
	n := float64(0)
	for i := range files {
		if files[i].Content == nil {
			newContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, *files[i].Path, &github.RepositoryContentGetOptions{})
			for _, ok := err.(*github.RateLimitError); !ok; _, ok = err.(*github.RateLimitError) {
				log.Println("Hit rate limit, retrying in 10 seconds")
				time.Sleep(10 * time.Second)
			}

			if err != nil {
				log.Println("Failed to get file", err)
				continue
			}
			files[i].Content = newContent.Content
		}
		if files[i].Content == nil {
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(*files[i].Content)
		if err != nil {
			log.Fatal("Failed to decode base64 string")
		}
		score := read.Gfi(string(decoded))
		if !math.IsInf(score, 1) {
			totalScore += score
			n += 1.0
		}
	}

	return totalScore / n
}

// Project is repo with readme
type Project struct {
	Name            string
	Forks           int
	TotalBytes      int
	GunningFogIndex float64
}

func getFileData(ctx context.Context, client *github.Client, repos []github.Repository) []Project {
	var projects []Project

	for i := range repos {
		latestCommit, _, err := client.Repositories.GetCommit(ctx, *(*repos[i].Owner).Login, *repos[i].Name, *repos[i].DefaultBranch)

		if err != nil {
			log.Println("Failed to load latest commit:", err)
			continue
		}

		tree, _, err := client.Git.GetTree(ctx, *(*repos[i].Owner).Login, *repos[i].Name, *latestCommit.SHA, true)

		if err != nil {
			log.Println("Failed to load latest commit:", err)
			continue
		}

		files, totalBytes := filterToDocs(*tree)

		projects = append(projects, Project{
			Name:            *repos[i].FullName,
			Forks:           *repos[i].ForksCount,
			TotalBytes:      totalBytes,
			GunningFogIndex: getReadabilityScore(ctx, client, *(*repos[i].Owner).Login, *repos[i].Name, files),
		})

		log.Println(fmt.Sprintf("Retrieved factors for %s", *repos[i].Name))
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
