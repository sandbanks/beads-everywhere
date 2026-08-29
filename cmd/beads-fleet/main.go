package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"

	beadsfleet "beads-fleet"
	"beads-fleet/pkg/config"
	"beads-fleet/pkg/fleet"
	"beads-fleet/templates/components"
	"beads-fleet/templates/pages"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "beads-fleet",
	Short: "🌐 Multi-repo issue aggregator and web command center for Beads",
	Run: func(cmd *cobra.Command, args []string) {
		runReadyCmd("", "open", "")
	},
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	svc := fleet.NewService(cfg)

	// Subcommands
	rootCmd.AddCommand(newScanCmd(svc))
	rootCmd.AddCommand(newReposCmd(svc))
	rootCmd.AddCommand(newReadyCmd(svc))
	rootCmd.AddCommand(newSearchCmd(svc))
	rootCmd.AddCommand(newCreateCmd(svc))
	rootCmd.AddCommand(newWebCmd(svc))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newScanCmd(svc *fleet.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan filesystem roots for all repositories tracked with Beads",
		Run: func(cmd *cobra.Command, args []string) {
			repos, err := svc.GetProjectsWithStats()
			if err != nil {
				log.Fatalf("Discovery error: %v", err)
			}
			fmt.Printf("🔍 Discovered %d Beads repositories across fleet:\n\n", len(repos))
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "PROJECT\tOPEN\tIN_PROGRESS\tCLOSED\tTOTAL\tPATH")
			for _, r := range repos {
				fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%s\n", r.Name, r.OpenIssues, r.InProgIssues, r.ClosedIssues, r.TotalIssues, r.Path)
			}
			w.Flush()
		},
	}
}

func newReposCmd(svc *fleet.Service) *cobra.Command {
	return &cobra.Command{
		Use:     "repos",
		Aliases: []string{"list", "ls"},
		Short:   "List all tracked repositories and active issue counts",
		Run: func(cmd *cobra.Command, args []string) {
			repos, err := svc.GetProjectsWithStats()
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "PROJECT\tREADY\tIN_PROGRESS\tTOTAL\tPATH")
			for _, r := range repos {
				fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%s\n", r.Name, r.OpenIssues, r.InProgIssues, r.TotalIssues, r.Path)
			}
			w.Flush()
		},
	}
}

func newReadyCmd(svc *fleet.Service) *cobra.Command {
	var repoFilter string
	cmd := &cobra.Command{
		Use:   "ready",
		Short: "List all actionable, unblocked issues across all fleet projects",
		Run: func(cmd *cobra.Command, args []string) {
			issues, err := svc.ListFleetIssues(repoFilter, "open", "")
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			if len(issues) == 0 {
				fmt.Println("✨ All clear! No open issues across fleet.")
				return
			}
			fmt.Printf("⚡️ Ready Issues across Fleet (%d total):\n\n", len(issues))
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PRIORITY\tPROJECT\tID\tTITLE")
			for _, iss := range issues {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", iss.PriorityLabel(), iss.Project, iss.ID, iss.Title)
			}
			w.Flush()
		},
	}
	cmd.Flags().StringVarP(&repoFilter, "repo", "r", "", "Filter by project name")
	return cmd
}

func newSearchCmd(svc *fleet.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search issues across all repositories in the fleet",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := args[0]
			issues, err := svc.ListFleetIssues("all", "all", query)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			if len(issues) == 0 {
				fmt.Printf("No issues matching %q found.\n", query)
				return
			}
			fmt.Printf("🔍 Search results for %q (%d found):\n\n", query, len(issues))
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "STATUS\tPRIORITY\tPROJECT\tID\tTITLE")
			for _, iss := range issues {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", iss.Status, iss.PriorityLabel(), iss.Project, iss.ID, iss.Title)
			}
			w.Flush()
		},
	}
}

func newCreateCmd(svc *fleet.Service) *cobra.Command {
	var repo, title, desc, issueType string
	var priority int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue in any fleet repository from anywhere",
		Run: func(cmd *cobra.Command, args []string) {
			if repo == "" || title == "" {
				log.Fatal("Error: --repo and --title are required")
			}
			iss, err := svc.CreateIssue(repo, title, desc, issueType, priority)
			if err != nil {
				log.Fatalf("Create error: %v", err)
			}
			fmt.Printf("✓ Created %s in %s: %s\n", iss.ID, iss.Project, iss.Title)
		},
	}

	cmd.Flags().StringVarP(&repo, "repo", "r", "", "Target repository name (required)")
	cmd.Flags().StringVarP(&title, "title", "t", "", "Issue title (required)")
	cmd.Flags().StringVarP(&desc, "description", "d", "", "Issue description")
	cmd.Flags().StringVar(&issueType, "type", "task", "Issue type (task, feature, bug, chore)")
	cmd.Flags().IntVarP(&priority, "priority", "p", 2, "Priority (0=P0, 1=P1, 2=P2, 3=P3, 4=P4)")

	return cmd
}

func newWebCmd(svc *fleet.Service) *cobra.Command {
	var port string
	cmd := &cobra.Command{
		Use:   "web",
		Short: "Start the Beads Fleet web command center",
		Run: func(cmd *cobra.Command, args []string) {
			if port == "" {
				port = svc.Config().Port
			}
			if port == "" {
				port = "8420"
			}
			runWebServer(svc, port)
		},
	}
	cmd.Flags().StringVarP(&port, "port", "p", "", "Port to listen on (default 8420)")
	return cmd
}

func runReadyCmd(repo, status, search string) {
	cfg, _ := config.LoadConfig()
	svc := fleet.NewService(cfg)
	issues, _ := svc.ListFleetIssues(repo, status, search)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PRIORITY\tPROJECT\tID\tTITLE")
	for _, iss := range issues {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", iss.PriorityLabel(), iss.Project, iss.ID, iss.Title)
	}
	w.Flush()
}

func runWebServer(svc *fleet.Service, port string) {
	log.Printf("🌐 Beads Fleet Web Command Center running on http://127.0.0.1:%s", port)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// Static Assets
	staticContent, err := fs.Sub(beadsfleet.StaticFS, "static")
	if err == nil {
		r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticContent))))
	} else {
		r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	}

	// Fleet Home
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		selectedRepo := r.URL.Query().Get("repo")
		status := r.URL.Query().Get("status")
		if status == "" {
			status = "open"
		}
		projects, err := svc.GetProjectsWithStats()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		issues, err := svc.ListFleetIssues(selectedRepo, status, "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templ.Handler(pages.Index(issues, projects, selectedRepo, status, "")).ServeHTTP(w, r)
	})

	// Filter Tabs (HTMX endpoint returning updated IssueList)
	r.Get("/issues/filter", func(w http.ResponseWriter, r *http.Request) {
		selectedRepo := r.URL.Query().Get("repo")
		status := r.URL.Query().Get("status")
		issues, err := svc.ListFleetIssues(selectedRepo, status, "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templ.Handler(components.IssueList(issues)).ServeHTTP(w, r)
	})

	// Live Search (HTMX keyup endpoint)
	r.Get("/issues/search", func(w http.ResponseWriter, r *http.Request) {
		selectedRepo := r.URL.Query().Get("repo")
		query := r.URL.Query().Get("search")
		issues, err := svc.ListFleetIssues(selectedRepo, "all", query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templ.Handler(components.IssueList(issues)).ServeHTTP(w, r)
	})

	// Create Issue
	r.Post("/issues", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		repo := r.FormValue("repo")
		title := r.FormValue("title")
		desc := r.FormValue("description")
		issueType := r.FormValue("issue_type")
		priorityStr := r.FormValue("priority")
		priority := 2
		if p, err := strconv.Atoi(priorityStr); err == nil {
			priority = p
		}

		_, err := svc.CreateIssue(repo, title, desc, issueType, priority)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return updated list of open issues
		issues, _ := svc.ListFleetIssues(repo, "open", "")
		templ.Handler(components.IssueList(issues)).ServeHTTP(w, r)
	})

	// Update Issue Status
	r.Post("/issues/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		newStatus := r.URL.Query().Get("status")
		if newStatus == "" {
			newStatus = r.FormValue("status")
		}

		updated, err := svc.UpdateIssueStatus(id, newStatus)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templ.Handler(components.IssueCard(*updated)).ServeHTTP(w, r)
	})

	// Get Edit Modal
	r.Get("/issues/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		issue, err := svc.GetIssue(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		templ.Handler(components.IssueEditModal(*issue)).ServeHTTP(w, r)
	})

	// Edit Issue
	r.Post("/issues/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		title := r.FormValue("title")
		desc := r.FormValue("description")
		issueType := r.FormValue("issue_type")
		priorityStr := r.FormValue("priority")
		priority := 2
		if p, err := strconv.Atoi(priorityStr); err == nil {
			priority = p
		}

		updated, err := svc.UpdateIssue(id, title, desc, priority, issueType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templ.Handler(components.IssueCard(*updated)).ServeHTTP(w, r)
	})

	// Delete Issue
	r.Delete("/issues/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := svc.DeleteIssue(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Sync All Repos
	r.Post("/sync", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			if err := svc.SyncAll(); err != nil {
				log.Printf("❌ Fleet sync error: %v", err)
			} else {
				log.Printf("✅ Fleet sync completed")
			}
		}()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Syncing fleet..."))
	})

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
