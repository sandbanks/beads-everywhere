package fleet

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"beads-fleet/pkg/config"
	"beads-fleet/pkg/discovery"
	"beads-fleet/pkg/models"
)

type IssuesResponse struct {
	Issues []models.Issue `json:"issues"`
	Total  int            `json:"total"`
}

type Service struct {
	cfg        *config.Config
	discoverer *discovery.Discoverer
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		cfg:        cfg,
		discoverer: discovery.NewDiscoverer(cfg),
	}
}

func (s *Service) Config() *config.Config {
	return s.cfg
}

func (s *Service) GetProjectsWithStats() ([]models.Project, error) {
	repos, err := s.discoverer.FindRepositories()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	for i := range repos {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			issues := s.readRepoIssues(repos[idx].Path)
			repos[idx].TotalIssues = len(issues)
			for _, iss := range issues {
				switch iss.Status {
				case "open":
					repos[idx].OpenIssues++
				case "in_progress":
					repos[idx].InProgIssues++
				case "closed":
					repos[idx].ClosedIssues++
				}
			}
		}(i)
	}
	wg.Wait()

	return repos, nil
}

func (s *Service) findRepoPath(repoName string) (string, error) {
	repos, err := s.discoverer.FindRepositories()
	if err != nil {
		return "", err
	}
	for _, r := range repos {
		if strings.EqualFold(r.Name, repoName) || r.Path == repoName {
			return r.Path, nil
		}
	}
	return "", fmt.Errorf("project %q not found in fleet", repoName)
}

func (s *Service) readRepoIssues(repoPath string) []models.Issue {
	repoName := filepath.Base(repoPath)

	// Try br list --json first
	cmd := exec.Command("br", "list", "--json")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	var list []models.Issue

	if err == nil && len(output) > 0 {
		var resp IssuesResponse
		if err := json.Unmarshal(output, &resp); err == nil {
			list = resp.Issues
		}
	}

	// Fallback to direct JSONL read
	if len(list) == 0 {
		jsonlPath := filepath.Join(repoPath, ".beads", "issues.jsonl")
		content, err := os.ReadFile(jsonlPath)
		if err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				var iss models.Issue
				if err := json.Unmarshal([]byte(line), &iss); err == nil {
					list = append(list, iss)
				}
			}
		}
	}

	for i := range list {
		list[i].Project = repoName
		list[i].ProjectPath = repoPath
	}

	return list
}

func (s *Service) ListFleetIssues(repoFilter, statusFilter, search string) ([]models.Issue, error) {
	repos, err := s.discoverer.FindRepositories()
	if err != nil {
		return nil, err
	}

	var allIssues []models.Issue
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, repo := range repos {
		if repoFilter != "" && repoFilter != "all" && !strings.EqualFold(repo.Name, repoFilter) && repo.Path != repoFilter {
			continue
		}

		wg.Add(1)
		go func(r models.Project) {
			defer wg.Done()
			issues := s.readRepoIssues(r.Path)
			mu.Lock()
			allIssues = append(allIssues, issues...)
			mu.Unlock()
		}(repo)
	}

	wg.Wait()

	// Filter
	var filtered []models.Issue
	query := strings.ToLower(search)

	for _, iss := range allIssues {
		if statusFilter != "" && statusFilter != "all" {
			if statusFilter == "ready" || statusFilter == "open" {
				if iss.Status != "open" {
					continue
				}
			} else if iss.Status != statusFilter {
				continue
			}
		}
		if search != "" {
			if !strings.Contains(strings.ToLower(iss.Title), query) &&
				!strings.Contains(strings.ToLower(iss.Description), query) &&
				!strings.Contains(strings.ToLower(iss.ID), query) &&
				!strings.Contains(strings.ToLower(iss.Project), query) {
				continue
			}
		}
		filtered = append(filtered, iss)
	}

	// Sort by priority (asc 0->4) then created_at (desc)
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Priority != filtered[j].Priority {
			return filtered[i].Priority < filtered[j].Priority
		}
		return filtered[i].CreatedAt > filtered[j].CreatedAt
	})

	return filtered, nil
}

func (s *Service) GetIssue(id string) (*models.Issue, error) {
	issues, err := s.ListFleetIssues("all", "all", id)
	if err != nil {
		return nil, err
	}
	for _, iss := range issues {
		if iss.ID == id {
			return &iss, nil
		}
	}
	return nil, fmt.Errorf("issue %q not found in fleet", id)
}

func (s *Service) CreateIssue(repoName, title, desc, issueType string, priority int) (*models.Issue, error) {
	repoPath, err := s.findRepoPath(repoName)
	if err != nil {
		return nil, err
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}
	if issueType == "" {
		issueType = "task"
	}
	if desc == "" {
		desc = title
	}

	cmd := exec.Command("br", "create",
		"--title="+title,
		"--description="+desc,
		"--type="+issueType,
		"--priority="+strconv.Itoa(priority),
	)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("br create failed in %s: %s (%w)", repoName, string(out), err)
	}

	// Extract created issue ID
	createdID := ""
	outStr := string(out)
	prefix := filepath.Base(repoPath) + "-"
	if idx := strings.Index(outStr, prefix); idx != -1 {
		endIdx := strings.IndexAny(outStr[idx:], ": \n\t")
		if endIdx != -1 {
			createdID = outStr[idx : idx+endIdx]
		}
	}

	go s.syncGitRepo(repoPath, "bead: "+title)

	if createdID != "" {
		if iss, err := s.GetIssue(createdID); err == nil {
			return iss, nil
		}
	}

	return &models.Issue{
		ID:          createdID,
		Project:     filepath.Base(repoPath),
		ProjectPath: repoPath,
		Title:       title,
		Description: desc,
		Status:      "open",
		Priority:    priority,
		IssueType:   issueType,
	}, nil
}

func (s *Service) UpdateIssueStatus(id, newStatus string) (*models.Issue, error) {
	iss, err := s.GetIssue(id)
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	if newStatus == "closed" {
		cmd = exec.Command("br", "close", id, "--reason=Completed")
	} else {
		cmd = exec.Command("br", "update", id, "--status="+newStatus)
	}
	cmd.Dir = iss.ProjectPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("br update failed: %s (%w)", string(out), err)
	}

	go s.syncGitRepo(iss.ProjectPath, "update "+id+" -> "+newStatus)

	return s.GetIssue(id)
}

func (s *Service) DeleteIssue(id string) error {
	iss, err := s.GetIssue(id)
	if err != nil {
		return err
	}

	cmd := exec.Command("br", "delete", id)
	cmd.Dir = iss.ProjectPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("br delete failed: %s (%w)", string(out), err)
	}

	go s.syncGitRepo(iss.ProjectPath, "delete issue "+id)
	return nil
}

func (s *Service) SyncAll() error {
	repos, err := s.discoverer.FindRepositories()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, r := range repos {
		if !r.HasGit {
			continue
		}
		wg.Add(1)
		go func(repoPath string) {
			defer wg.Done()
			s.syncGitRepo(repoPath, "")
		}(r.Path)
	}
	wg.Wait()
	return nil
}

func (s *Service) syncGitRepo(repoPath, commitMsg string) {
	flushCmd := exec.Command("br", "sync", "--flush-only")
	flushCmd.Dir = repoPath
	_ = flushCmd.Run()

	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		return
	}

	pullCmd := exec.Command("git", "pull", "--rebase")
	pullCmd.Dir = repoPath
	_ = pullCmd.Run()

	if commitMsg != "" {
		addCmd := exec.Command("git", "add", ".beads/issues.jsonl")
		addCmd.Dir = repoPath
		_ = addCmd.Run()

		commitCmd := exec.Command("git", "commit", "-m", commitMsg)
		commitCmd.Dir = repoPath
		_ = commitCmd.Run()

		time.Sleep(200 * time.Millisecond)
		pushCmd := exec.Command("git", "push")
		pushCmd.Dir = repoPath
		_ = pushCmd.Run()
	}
}
