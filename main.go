package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

const (
	primaryToken   = "your_primary_github_token"   // Replace with your primary account token
	secondaryToken = "your_secondary_github_token" // Replace with your secondary account token
	primaryUser    = "your_primary_username"       // Replace with your primary username
	secondaryUser  = "your_secondary_username"     // Replace with your secondary username
	backupDir      = "./github_backups"            // Directory to store temporary clones
)

func main() {
	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Fatalf("Failed to create backup directory: %v", err)
	}

	// Set up GitHub clients
	primaryClient := createGitHubClient(primaryToken)
	secondaryClient := createGitHubClient(secondaryToken)

	// Get all repositories from primary account
	repos, err := getAllRepos(primaryClient, primaryUser)
	if err != nil {
		log.Fatalf("Failed to get repositories: %v", err)
	}

	fmt.Printf("Found %d repositories to backup\n", len(repos))

	// Process each repository
	for _, repo := range repos {
		if err := backupRepo(repo, secondaryClient); err != nil {
			log.Printf("Failed to backup %s: %v", repo.GetName(), err)
			continue
		}
	}

	fmt.Println("Backup completed!")
}

func createGitHubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func getAllRepos(client *github.Client, username string) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opt := &github.RepositoryListOptions{
		Type: "owner",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := client.Repositories.List(context.Background(), username, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func backupRepo(repo *github.Repository, secondaryClient *github.Client) error {
	repoName := repo.GetName()
	fmt.Printf("Processing repository: %s\n", repoName)

	// Check if repo exists in secondary account
	_, _, err := secondaryClient.Repositories.Get(context.Background(), secondaryUser, repoName)
	if err == nil {
		fmt.Printf("Repository %s already exists in secondary account, skipping...\n", repoName)
		return nil
	}

	// Create repo in secondary account
	newRepo := &github.Repository{
		Name:        github.String(repoName),
		Description: repo.Description,
		Private:     repo.Private,
	}

	_, _, err = secondaryClient.Repositories.Create(context.Background(), "", newRepo)
	if err != nil {
		return fmt.Errorf("failed to create repository in secondary account: %v", err)
	}

	// Clone from primary
	cloneURL := strings.Replace(repo.GetCloneURL(), "https://", fmt.Sprintf("https://%s@", primaryUser), 1)
	repoPath := filepath.Join(backupDir, repoName)

	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to clean up existing directory: %v", err)
	}

	if err := runCommand("git", "clone", "--mirror", cloneURL, repoPath); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	// Push to secondary
	pushURL := strings.Replace(
		fmt.Sprintf("https://github.com/%s/%s.git", secondaryUser, repoName),
		"https://",
		fmt.Sprintf("https://%s@", secondaryUser),
		1,
	)

	if err := runCommandInDir(repoPath, "git", "remote", "set-url", "--push", "origin", pushURL); err != nil {
		return fmt.Errorf("failed to set push URL: %v", err)
	}

	if err := runCommandInDir(repoPath, "git", "push", "--mirror"); err != nil {
		return fmt.Errorf("failed to push to secondary: %v", err)
	}

	// Clean up
	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to clean up repository: %v", err)
	}

	fmt.Printf("Successfully backed up %s\n", repoName)
	return nil
}

func runCommand(name string, arg ...string) error {
	return runCommandInDir("", name, arg...)
}

func runCommandInDir(dir, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}