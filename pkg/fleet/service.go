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

	"github.com/sandbanks/beads-everywhere/pkg/config"
	"github.com/sandbanks/beads-everywhere/pkg/discovery"
	"github.com/sandbanks/beads-everywhere/pkg/models"
)

func getBeadsBin() string {
	if bin, err := exec.LookPath("br"); err == nil && bin != "" {
		return bin
	}
	if bin, err := exec.LookPath("bd"); err == nil && bin != "" {
		return bin
	}
	return "br"
}

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
	return s.loadRepoStats(repos), nil
}

func (s *Service) GetAllProjectsWithStats() ([]models.Project, error) {
	repos, err := s.discoverer.FindAllRepositories()
	if err != nil {
		return nil, err
	}
	return s.loadRepoStats(repos), nil
}

func (s *Service) loadRepoStats(repos []models.Project) []models.Project {
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

	return repos
}

func (s *Service) findRepoPath(repoName string) (string, error) {
	repos, err := s.discoverer.FindAllRepositories()
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
	var list []models.Issue

	// 1. Direct, ultra-fast JSONL read (< 0.1ms)
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

	// 2. Fallback to `br list --json` (or `bd list --json`) if JSONL was empty
	if len(list) == 0 {
		cmd := exec.Command(getBeadsBin(), "list", "--json")
		cmd.Dir = repoPath
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			var resp IssuesResponse
			if err := json.Unmarshal(output, &resp); err == nil {
				list = resp.Issues
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

	cmd := exec.Command(getBeadsBin(), "create",
		"--title="+title,
		"--description="+desc,
		"--type="+issueType,
		"--priority="+strconv.Itoa(priority),
	)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("beads create failed in %s: %s (%w)", repoName, string(out), err)
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
		cmd = exec.Command(getBeadsBin(), "close", id, "--reason=Completed")
	} else {
		cmd = exec.Command(getBeadsBin(), "update", id, "--status="+newStatus)
	}
	cmd.Dir = iss.ProjectPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("beads update failed: %s (%w)", string(out), err)
	}

	go s.syncGitRepo(iss.ProjectPath, "update "+id+" -> "+newStatus)

	return s.GetIssue(id)
}

func (s *Service) UpdateIssue(id string, title string, desc string, priority int, issueType string) (*models.Issue, error) {
	iss, err := s.GetIssue(id)
	if err != nil {
		return nil, err
	}

	args := []string{"update", id}
	if title != "" {
		args = append(args, "--title="+title)
	}
	if desc != "" {
		args = append(args, "--description="+desc)
	}
	if priority >= 0 && priority <= 4 {
		args = append(args, "--priority="+strconv.Itoa(priority))
	}
	if issueType != "" {
		args = append(args, "--type="+issueType)
	}

	cmd := exec.Command(getBeadsBin(), args...)
	cmd.Dir = iss.ProjectPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("beads update failed: %s (%w)", string(out), err)
	}

	go s.syncGitRepo(iss.ProjectPath, "edit issue: "+id)

	return s.GetIssue(id)
}

func (s *Service) DeleteIssue(id string) error {
	iss, err := s.GetIssue(id)
	if err != nil {
		return err
	}

	cmd := exec.Command(getBeadsBin(), "delete", id)
	cmd.Dir = iss.ProjectPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("beads delete failed: %s (%w)", string(out), err)
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
	flushCmd := exec.Command(getBeadsBin(), "sync", "--flush-only")
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

type schemaMigrationPlanOutput struct {
	Eligible     bool   `json:"eligible"`
	FromVersion  int    `json:"from_version"`
	ToVersion    int    `json:"to_version"`
	PlanToken    string `json:"plan_token"`
	ApplyCommand string `json:"apply_command"`
	Note         string `json:"note"`
}

type doctorOutput struct {
	WorkspaceHealth string `json:"workspace_health"`
}

func (s *Service) DoctorAndMigrate(repair, activeOnly bool) ([]models.DoctorRepoResult, error) {
	var repos []models.Project
	var err error
	if activeOnly {
		repos, err = s.discoverer.FindRepositories()
	} else {
		repos, err = s.discoverer.FindAllRepositories()
	}
	if err != nil {
		return nil, err
	}

	results := make([]models.DoctorRepoResult, len(repos))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)

	for i := range repos {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			results[idx] = s.doctorRepo(repos[idx], repair)
		}(i)
	}

	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
	})

	return results, nil
}

func (s *Service) doctorRepo(repo models.Project, repair bool) models.DoctorRepoResult {
	res := models.DoctorRepoResult{
		Name:     repo.Name,
		Path:     repo.Path,
		Health:   "unknown",
		Archived: repo.Archived,
	}

	bin := getBeadsBin()

	// 1. Schema migration check & auto-apply
	planCmd := exec.Command(bin, "doctor", "migrate-schema", "plan", "--json")
	planCmd.Dir = repo.Path
	planOut, _ := planCmd.CombinedOutput()

	var plan schemaMigrationPlanOutput
	if err := json.Unmarshal(planOut, &plan); err == nil {
		res.FromVersion = plan.FromVersion
		res.ToVersion = plan.ToVersion

		if plan.Eligible && plan.PlanToken != "" {
			applyCmd := exec.Command(bin, "doctor", "migrate-schema", "apply", "--plan-token", plan.PlanToken, "--json")
			applyCmd.Dir = repo.Path
			applyOut, applyErr := applyCmd.CombinedOutput()
			if applyErr != nil {
				res.Error = fmt.Sprintf("migration failed: %s", strings.TrimSpace(string(applyOut)))
			} else {
				res.Migrated = true
				res.FromVersion = plan.FromVersion
				res.ToVersion = plan.ToVersion
			}
		}
	} else {
		outStr := strings.TrimSpace(string(planOut))
		if strings.Contains(outStr, "Schema version mismatch") {
			res.Error = outStr
		}
	}

	// 2. Health check
	docCmd := exec.Command(bin, "doctor", "--json")
	docCmd.Dir = repo.Path
	docOut, _ := docCmd.CombinedOutput()

	var doc doctorOutput
	if err := json.Unmarshal(docOut, &doc); err == nil && doc.WorkspaceHealth != "" {
		res.Health = doc.WorkspaceHealth
	} else {
		for _, line := range strings.Split(string(docOut), "\n") {
			if strings.Contains(line, "HEALTH workspace:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					res.Health = strings.TrimSpace(parts[1])
				}
				break
			}
		}
		if res.Health == "unknown" && len(docOut) > 0 {
			res.Error = strings.TrimSpace(string(docOut))
		}
	}

	// 3. Optional repair if degraded/recoverable
	if repair && (res.Health == "recoverable" || res.Health == "degraded" || res.Health == "unsafe" || res.Health == "error") {
		repCmd := exec.Command(bin, "doctor", "--repair")
		repCmd.Dir = repo.Path
		_ = repCmd.Run()
		res.Repaired = true

		docCmd2 := exec.Command(bin, "doctor", "--json")
		docCmd2.Dir = repo.Path
		if docOut2, err := docCmd2.CombinedOutput(); err == nil {
			var doc2 doctorOutput
			if err := json.Unmarshal(docOut2, &doc2); err == nil && doc2.WorkspaceHealth != "" {
				res.Health = doc2.WorkspaceHealth
			}
		}
	}

	// 4. Issue counts
	issues := s.readRepoIssues(repo.Path)
	res.TotalIssues = len(issues)
	for _, iss := range issues {
		if iss.Status != "closed" {
			res.OpenIssues++
		}
	}

	return res
}

