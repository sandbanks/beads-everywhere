package discovery

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"beads-fleet/pkg/config"
	"beads-fleet/pkg/models"
)

type Discoverer struct {
	cfg *config.Config
}

func NewDiscoverer(cfg *config.Config) *Discoverer {
	return &Discoverer{cfg: cfg}
}

func (d *Discoverer) isRepoPermitted(name, path string) bool {
	// If allowed_repos whitelist is defined, must match at least one
	if len(d.cfg.AllowedRepos) > 0 {
		allowed := false
		for _, allow := range d.cfg.AllowedRepos {
			if strings.EqualFold(name, allow) || path == allow || strings.HasSuffix(path, "/"+allow) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	// If hidden_repos blacklist is defined, must not match any
	for _, hide := range d.cfg.HiddenRepos {
		if strings.EqualFold(name, hide) || path == hide || strings.HasSuffix(path, "/"+hide) {
			return false
		}
	}

	return true
}

func (d *Discoverer) FindRepositories() ([]models.Project, error) {
	ignoredMap := make(map[string]bool)
	for _, ign := range d.cfg.IgnoredDirs {
		ignoredMap[ign] = true
	}

	var mu sync.Mutex
	projectsMap := make(map[string]models.Project)
	var wg sync.WaitGroup

	for _, root := range d.cfg.ScanRoots {
		rootPath, err := filepath.Abs(root)
		if err != nil {
			continue
		}

		info, err := os.Stat(rootPath)
		if err != nil || !info.IsDir() {
			continue
		}

		// If root itself has .beads
		if _, err := os.Stat(filepath.Join(rootPath, ".beads")); err == nil {
			name := filepath.Base(rootPath)
			if d.isRepoPermitted(name, rootPath) {
				p := models.Project{
					Name:   name,
					Path:   rootPath,
					HasGit: false,
				}
				if _, err := os.Stat(filepath.Join(rootPath, ".git")); err == nil {
					p.HasGit = true
				}
				mu.Lock()
				projectsMap[rootPath] = p
				mu.Unlock()
			}
		}

		wg.Add(1)
		go func(searchDir string) {
			defer wg.Done()

			_ = filepath.WalkDir(searchDir, func(path string, dEntry fs.DirEntry, walkErr error) error {
				if walkErr != nil || !dEntry.IsDir() {
					return nil
				}

				name := dEntry.Name()

				// Skip ignored directories (except search root itself)
				if path != searchDir && ignoredMap[name] {
					return filepath.SkipDir
				}

				// Check if this directory contains a .beads folder
				if name == ".beads" {
					repoDir := filepath.Dir(path)
					repoName := filepath.Base(repoDir)

					if d.isRepoPermitted(repoName, repoDir) {
						hasGit := false
						if _, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
							hasGit = true
						}

						mu.Lock()
						projectsMap[repoDir] = models.Project{
							Name:   repoName,
							Path:   repoDir,
							HasGit: hasGit,
						}
						mu.Unlock()
					}

					return filepath.SkipDir
				}

				return nil
			})
		}(rootPath)
	}

	wg.Wait()

	var result []models.Project
	for _, p := range projectsMap {
		result = append(result, p)
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result, nil
}
