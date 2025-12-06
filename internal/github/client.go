// Package github provides GitHub API client functionality for PR status and checks.
package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/ryantking/agentctl/internal/git"
)

// Client wraps the GitHub API client using gh CLI authentication.
type Client struct {
	apiClient *api.RESTClient
	owner     string
	repo      string
}

// NewClient creates a new GitHub API client using gh CLI authentication.
// Uses gh's authentication flow (gh auth login) automatically.
func NewClient(repoRoot string) (*Client, error) {
	// Get repository owner and name
	owner, repoName, err := getRepoInfo(repoRoot)
	if err != nil {
		return nil, err
	}

	// Create API client using gh's authentication
	apiClient, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub API client: %w", err)
	}

	return &Client{
		apiClient: apiClient,
		owner:     owner,
		repo:      repoName,
	}, nil
}

// getRepoInfo extracts owner and repo name from git remote.
func getRepoInfo(repoRoot string) (string, string, error) {
	gitRepo, err := git.OpenRepo(repoRoot)
	if err != nil {
		return "", "", fmt.Errorf("failed to open repository: %w", err)
	}

	remotes, err := gitRepo.Remotes()
	if err != nil {
		return "", "", fmt.Errorf("failed to get remotes: %w", err)
	}

	// Look for origin remote
	var originURL string
	for _, remote := range remotes {
		if remote.Config().Name == "origin" {
			urls := remote.Config().URLs
			if len(urls) > 0 {
				originURL = urls[0]
				break
			}
		}
	}

	if originURL == "" {
		return "", "", fmt.Errorf("no origin remote found")
	}

	// Parse GitHub URL (supports both https and ssh formats)
	owner, repoName := parseGitHubURL(originURL)
	if owner == "" || repoName == "" {
		return "", "", fmt.Errorf("failed to parse GitHub URL: %s", originURL)
	}

	return owner, repoName, nil
}

// parseGitHubURL extracts owner and repo from GitHub URL.
// Supports:
// - https://github.com/owner/repo.git
// - https://github.com/owner/repo
// - git@github.com:owner/repo.git
// - git@github.com:owner/repo
func parseGitHubURL(url string) (string, string) {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH format: git@github.com:owner/repo
	if strings.HasPrefix(url, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(url, "git@github.com:"), "/")
		if len(parts) >= 2 {
			return parts[0], parts[1]
		}
	}

	// Handle HTTPS format: https://github.com/owner/repo
	if strings.Contains(url, "github.com/") {
		parts := strings.Split(url, "github.com/")
		if len(parts) >= 2 {
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) >= 2 {
				return pathParts[0], pathParts[1]
			}
		}
	}

	return "", ""
}

// PRStatus represents PR status information.
type PRStatus struct {
	Number int
	Title  string
	URL    string
	Review string
	Checks string
}

// GetPRForBranch finds the PR associated with a branch.
func (c *Client) GetPRForBranch(_ context.Context, branch string) (*PRStatus, error) {
	// Query REST API for PRs with this head branch
	path := fmt.Sprintf("repos/%s/%s/pulls?head=%s:%s&state=open", c.owner, c.repo, c.owner, branch)
	
	var prs []struct {
		Number  int    `json:"number"`
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		State   string `json:"state"`
		Head    struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}

	err := c.apiClient.Get(path, &prs)
	if err != nil {
		return nil, fmt.Errorf("failed to query GitHub API: %w", err)
	}

	if len(prs) == 0 {
		return nil, fmt.Errorf("no open PR found for branch %s", branch)
	}

	pr := prs[0]
	status := &PRStatus{
		Number: pr.Number,
		Title:  pr.Title,
		URL:    pr.HTMLURL,
	}

	// Get PR details including review decision and checks
	prDetail, err := c.getPRDetails(pr.Number, pr.Head.SHA)
	if err == nil {
		if prDetail.ReviewDecision != "" {
			status.Review = prDetail.ReviewDecision
		}
		status.Checks = prDetail.ChecksSummary
	}

	return status, nil
}

// prDetailResponse represents PR detail response.
type prDetailResponse struct {
	ReviewDecision string
	ChecksSummary  string
}

// getPRDetails gets detailed PR information including checks.
func (c *Client) getPRDetails(_ int, headSHA string) (*prDetailResponse, error) {
	detail := &prDetailResponse{}

	// Get check runs for the PR head SHA
	checkPath := fmt.Sprintf("repos/%s/%s/commits/%s/check-runs", c.owner, c.repo, headSHA)
	
	var checkRunsResponse struct {
		TotalCount int `json:"total_count"`
		CheckRuns []struct {
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		} `json:"check_runs"`
	}

	err := c.apiClient.Get(checkPath, &checkRunsResponse)
	if err == nil {
		var passed, failed, pending int
		for _, run := range checkRunsResponse.CheckRuns {
			switch run.Status {
			case "completed":
				switch run.Conclusion {
				case "success":
					passed++
				case "failure", "cancelled":
					failed++
				}
			case "in_progress", "queued":
				pending++
			}
		}

		switch {
		case failed > 0:
			detail.ChecksSummary = fmt.Sprintf("%d failing", failed)
		case pending > 0:
			detail.ChecksSummary = fmt.Sprintf("%d pending", pending)
		case passed > 0:
			detail.ChecksSummary = fmt.Sprintf("%d passed", passed)
		}
	}

	return detail, nil
}
