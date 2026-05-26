package runtime

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SkillSource string

const (
	SkillSourceProject     SkillSource = "project"
	SkillSourceCodex       SkillSource = "codex"
	SkillSourceClaudeCode  SkillSource = "claude_code"
	SkillSourceAgents      SkillSource = "agents"
	SkillSourcePluginCache SkillSource = "plugin_cache"
)

type InstalledSkill struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Source      SkillSource `json:"source"`
	SourceLabel string      `json:"source_label"`
	Package     string      `json:"package,omitempty"`
	Path        string      `json:"path"`
	Installed   bool        `json:"installed"`
}

type SkillDiscoveryInput struct {
	HomeDir      string
	WorkspaceDir string
	ExtraRoots   []SkillRoot
	MaxFiles     int
}

type SkillRoot struct {
	Path        string
	Source      SkillSource
	SourceLabel string
	MaxDepth    int
}

func DefaultSkillDiscoveryInput() SkillDiscoveryInput {
	homeDir, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()
	return SkillDiscoveryInput{HomeDir: homeDir, WorkspaceDir: findSkillWorkspaceDir(cwd), MaxFiles: 500}
}

func DiscoverInstalledSkills(input SkillDiscoveryInput) ([]InstalledSkill, error) {
	if input.MaxFiles <= 0 {
		input.MaxFiles = 500
	}
	roots := skillRoots(input)
	seenRoots := map[string]bool{}
	seenSkills := map[string]bool{}
	skills := []InstalledSkill{}
	for _, root := range roots {
		rootPath := filepath.Clean(root.Path)
		if rootPath == "." || rootPath == "" || seenRoots[rootPath] {
			continue
		}
		seenRoots[rootPath] = true
		info, err := os.Stat(rootPath)
		if err != nil || !info.IsDir() {
			continue
		}
		maxDepth := root.MaxDepth
		if maxDepth <= 0 {
			maxDepth = 6
		}
		err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if len(skills) >= input.MaxFiles {
				return filepath.SkipAll
			}
			if d.IsDir() {
				name := d.Name()
				if name == ".git" || name == "node_modules" || name == "vendor" {
					return filepath.SkipDir
				}
				if pathDepth(rootPath, path) > maxDepth {
					return filepath.SkipDir
				}
				return nil
			}
			if d.Name() != "SKILL.md" {
				return nil
			}
			cleanPath := filepath.Clean(path)
			if seenSkills[cleanPath] {
				return nil
			}
			seenSkills[cleanPath] = true
			skill := parseInstalledSkill(rootPath, cleanPath, root)
			skills = append(skills, skill)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(skills, func(i, j int) bool {
		if skills[i].Source != skills[j].Source {
			return skills[i].Source < skills[j].Source
		}
		return strings.ToLower(skills[i].Name) < strings.ToLower(skills[j].Name)
	})
	return skills, nil
}

func skillRoots(input SkillDiscoveryInput) []SkillRoot {
	var roots []SkillRoot
	if input.WorkspaceDir != "" {
		roots = append(roots,
			SkillRoot{Path: filepath.Join(input.WorkspaceDir, ".agents", "skills"), Source: SkillSourceProject, SourceLabel: "Project .agents", MaxDepth: 5},
			SkillRoot{Path: filepath.Join(input.WorkspaceDir, ".claude", "skills"), Source: SkillSourceClaudeCode, SourceLabel: "Project Claude Code", MaxDepth: 5},
		)
	}
	if input.HomeDir != "" {
		roots = append(roots,
			SkillRoot{Path: filepath.Join(input.HomeDir, ".codex", "skills"), Source: SkillSourceCodex, SourceLabel: "Codex", MaxDepth: 5},
			SkillRoot{Path: filepath.Join(input.HomeDir, ".agents", "skills"), Source: SkillSourceAgents, SourceLabel: "User agents", MaxDepth: 5},
			SkillRoot{Path: filepath.Join(input.HomeDir, ".claude", "skills"), Source: SkillSourceClaudeCode, SourceLabel: "Claude Code", MaxDepth: 5},
			SkillRoot{Path: filepath.Join(input.HomeDir, ".codex", "plugins", "cache"), Source: SkillSourcePluginCache, SourceLabel: "Codex plugins", MaxDepth: 10},
			SkillRoot{Path: filepath.Join(input.HomeDir, ".claude", "plugins"), Source: SkillSourceClaudeCode, SourceLabel: "Claude Code plugins", MaxDepth: 10},
		)
	}
	roots = append(roots, input.ExtraRoots...)
	return roots
}

func parseInstalledSkill(rootPath string, skillPath string, root SkillRoot) InstalledSkill {
	name, description := readSkillMetadata(skillPath)
	if name == "" {
		name = filepath.Base(filepath.Dir(skillPath))
	}
	sum := sha1.Sum([]byte(skillPath))
	return InstalledSkill{
		ID:          string(root.Source) + ":" + hex.EncodeToString(sum[:])[:12],
		Name:        name,
		Description: description,
		Source:      root.Source,
		SourceLabel: root.SourceLabel,
		Package:     skillPackage(rootPath, skillPath),
		Path:        skillPath,
		Installed:   true,
	}
}

func readSkillMetadata(path string) (string, string) {
	file, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, 16*1024))
	if err != nil {
		return "", ""
	}
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	if strings.HasPrefix(text, "---\n") {
		if end := strings.Index(text[4:], "\n---"); end >= 0 {
			frontmatter := text[4 : 4+end]
			name := frontmatterValue(frontmatter, "name")
			description := frontmatterValue(frontmatter, "description")
			if name != "" || description != "" {
				return name, description
			}
		}
	}
	var name, description string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "---" {
			continue
		}
		if name == "" && strings.HasPrefix(trimmed, "#") {
			name = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			continue
		}
		if description == "" && !strings.HasPrefix(trimmed, "#") && !strings.Contains(trimmed, ":") {
			description = trimmed
		}
		if name != "" && description != "" {
			break
		}
	}
	return name, description
}

func frontmatterValue(frontmatter string, key string) string {
	prefix := key + ":"
	for _, line := range strings.Split(frontmatter, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, prefix) {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		value = strings.Trim(value, `"'`)
		return value
	}
	return ""
}

func skillPackage(rootPath string, skillPath string) string {
	rel, err := filepath.Rel(rootPath, filepath.Dir(skillPath))
	if err != nil || rel == "." {
		return ""
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) <= 1 {
		return ""
	}
	return parts[0]
}

func pathDepth(rootPath string, path string) int {
	rel, err := filepath.Rel(rootPath, path)
	if err != nil || rel == "." {
		return 0
	}
	return len(strings.Split(rel, string(filepath.Separator)))
}

func findSkillWorkspaceDir(start string) string {
	if start == "" {
		return ""
	}
	current := filepath.Clean(start)
	for {
		if info, err := os.Stat(filepath.Join(current, ".agents", "skills")); err == nil && info.IsDir() {
			return current
		}
		if info, err := os.Stat(filepath.Join(current, "go.mod")); err == nil && !info.IsDir() {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return start
		}
		current = parent
	}
}
