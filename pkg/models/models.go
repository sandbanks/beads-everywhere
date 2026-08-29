package models

type Project struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	TotalIssues  int    `json:"total_issues"`
	OpenIssues   int    `json:"open_issues"`
	InProgIssues int    `json:"in_prog_issues"`
	ClosedIssues int    `json:"closed_issues"`
	HasGit       bool   `json:"has_git"`
	GitBranch    string `json:"git_branch,omitempty"`
}

type Issue struct {
	ID              string   `json:"id"`
	Project         string   `json:"project"`
	ProjectPath     string   `json:"project_path"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Status          string   `json:"status"` // "open", "in_progress", "closed", "deferred"
	Priority        int      `json:"priority"` // 0=P0, 1=P1, 2=P2, 3=P3, 4=P4
	IssueType       string   `json:"issue_type"` // "feature", "task", "bug", "epic", "chore"
	CreatedAt       string   `json:"created_at"`
	CreatedBy       string   `json:"created_by,omitempty"`
	UpdatedAt       string   `json:"updated_at"`
	SourceRepo      string   `json:"source_repo,omitempty"`
	SourceRepoPath  string   `json:"source_repo_path,omitempty"`
	DependencyCount int      `json:"dependency_count,omitempty"`
	DependentCount  int      `json:"dependent_count,omitempty"`
	BlockedBy       []string `json:"blocked_by,omitempty"`
}

func (i *Issue) PriorityBadgeClass() string {
	switch i.Priority {
	case 0:
		return "bg-rose-500/20 text-rose-400 border-rose-500/40"
	case 1:
		return "bg-amber-500/20 text-amber-400 border-amber-500/40"
	case 2:
		return "bg-sky-500/20 text-sky-400 border-sky-500/40"
	case 3:
		return "bg-indigo-500/20 text-indigo-400 border-indigo-500/40"
	default:
		return "bg-zinc-700/20 text-zinc-400 border-zinc-700/40"
	}
}

func (i *Issue) PriorityLabel() string {
	switch i.Priority {
	case 0:
		return "P0 Critical 🔥"
	case 1:
		return "P1 High ⚡️"
	case 2:
		return "P2 Normal 📌"
	case 3:
		return "P3 Low 💤"
	default:
		return "P4 Backlog 📋"
	}
}

func (i *Issue) StatusBadgeClass() string {
	switch i.Status {
	case "in_progress":
		return "bg-emerald-500/20 text-emerald-400 border-emerald-500/40"
	case "closed":
		return "bg-zinc-800 text-zinc-500 border-zinc-700/50"
	case "deferred":
		return "bg-purple-500/20 text-purple-400 border-purple-500/40"
	default:
		return "bg-amber-500/20 text-amber-400 border-amber-500/40"
	}
}
