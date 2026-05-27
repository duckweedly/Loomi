package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverInstalledSkillsReadsKnownRootsOnly(t *testing.T) {
	home := t.TempDir()
	workspace := t.TempDir()
	writeSkill(t, filepath.Join(home, ".codex", "skills", "review", "SKILL.md"), "---\nname: code-review\ndescription: Review local code.\n---\nsecret body")
	writeSkill(t, filepath.Join(workspace, ".agents", "skills", "speckit-implement", "SKILL.md"), "# speckit-implement\n\nExecute implementation tasks.")
	writeSkill(t, filepath.Join(home, ".codex", "plugins", "cache", "openai-curated", "github", "abc", "skills", "github", "SKILL.md"), "---\nname: github:github\ndescription: GitHub workflow helper.\n---")

	skills, err := DiscoverInstalledSkills(SkillDiscoveryInput{HomeDir: home, WorkspaceDir: workspace})
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 3 {
		t.Fatalf("len(skills) = %d, want 3: %+v", len(skills), skills)
	}
	byName := map[string]InstalledSkill{}
	for _, skill := range skills {
		byName[skill.Name] = skill
		if !skill.Installed || skill.Path == "" || skill.ID == "" {
			t.Fatalf("skill missing stable fields: %+v", skill)
		}
	}
	if byName["code-review"].Description != "Review local code." || byName["code-review"].Source != SkillSourceCodex {
		t.Fatalf("codex skill = %+v", byName["code-review"])
	}
	if byName["speckit-implement"].Description != "Execute implementation tasks." || byName["speckit-implement"].Source != SkillSourceProject {
		t.Fatalf("project skill = %+v", byName["speckit-implement"])
	}
	if byName["github:github"].Source != SkillSourcePluginCache {
		t.Fatalf("plugin skill = %+v", byName["github:github"])
	}
}

func writeSkill(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
