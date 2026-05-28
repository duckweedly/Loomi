package productdata

import (
	"strings"
	"testing"
)

func TestBuiltInPersonaDefaultPromptStaysConcise(t *testing.T) {
	persona := BuiltInPersonas()[0]

	for _, want := range []string{"Answer first", "Keep final answers brief", "Do not expose hidden chain-of-thought"} {
		if !strings.Contains(persona.SystemPrompt, want) {
			t.Fatalf("system prompt missing %q: %s", want, persona.SystemPrompt)
		}
	}
}
