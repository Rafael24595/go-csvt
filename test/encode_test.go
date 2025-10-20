package test

import (
	"strings"
	"testing"

	"github.com/Rafael24595/go-csvt/csvt"
	"github.com/Rafael24595/go-csvt/test/support"
)

func TestMarshal_LangStructure(t *testing.T) {
	lang1 := support.Lang{
		Name: "Go",
		Release: support.Release{
			Version: "1.25.3",
			Stable:  true,
		},
		Tags: []string{"go", "golang"},
		Attributes: map[string]string{
			"oop":        "some",
			"procedural": "true",
			"functional": "false",
		},
	}

	lang2 := support.Lang{
		Name: "Zig",
		Release: support.Release{
			Version: "0.16.0-dev.747+493ad58ff",
			Stable:  false,
		},
		Tags: []string{"zig", "ziglang"},
		Attributes: map[string]string{
			"oop":        "false",
			"procedural": "true",
			"functional": "false",
		},
	}

	result, err := csvt.Marshal(lang1, lang2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(result)

	if !strings.Contains(output, "Lang&") {
		t.Errorf("expected Lang table, not found: %s", output)
	}
	if !strings.Contains(output, "Release&") {
		t.Errorf("expected Release table, not found: %s", output)
	}
	if !strings.Contains(output, "common-array") {
		t.Errorf("expected common-array table, not found: %s", output)
	}
	if !strings.Contains(output, "common-map") {
		t.Errorf("expected common-map table, not found: %s", output)
	}

	if !strings.Contains(output, "Name;Release;Tags;Attributes") {
		t.Errorf("expected Lang headers, got: %s", output)
	}

	if !strings.Contains(output, "Version;Stable") {
		t.Errorf("expected Release headers, got: %s", output)
	}

	if !strings.Contains(output, "$Release") {
		t.Errorf("expected Release pointer references")
	}
	if !strings.Contains(output, "$common-array") {
		t.Errorf("expected common-array pointer references")
	}
	if !strings.Contains(output, "$common-map") {
		t.Errorf("expected common-map pointer references")
	}
}
