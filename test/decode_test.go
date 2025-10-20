package test

import (
	"testing"

	"github.com/Rafael24595/go-csvt/csvt"
	"github.com/Rafael24595/go-csvt/test/support"
)

func TestUnmarshalSimpleStruct(t *testing.T) {
	data := support.LoadFile(t, "../support/lang_table.csvt")

	var result []support.Lang
	err := csvt.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expLen := 2
	if len(result) != expLen {
		t.Fatalf("expected %d items, got %d", expLen, len(result))
	}

	lang1 := result[0]

	expName := "Go"
	if lang1.Name != expName {
		t.Errorf("expected Name '%s', got '%s'", expName, lang1.Name)
	}
	expVersion := "1.25.3"
	if lang1.Release.Version != expVersion {
		t.Errorf("expected Version '%s', got '%s'", expVersion, lang1.Release.Version)
	}
	expStable := true
	if lang1.Release.Stable != expStable {
		t.Errorf("expected Stable %v, got %v", expStable, lang1.Release.Stable)
	}
	expTags := []string{"go", "golang"}
	if len(lang1.Tags) != len(expTags) || lang1.Tags[0] != expTags[0] || lang1.Tags[1] != expTags[1] {
		t.Errorf("unexpected Tags: %v", lang1.Tags)
	}

	if lang1.Attributes["oop"] != "some" ||
		lang1.Attributes["procedural"] != "true" ||
		lang1.Attributes["functional"] != "false" {
		t.Errorf("unexpected Attributes: %v", lang1.Attributes)
	}

	lang2 := result[1]
	
	expName = "Zig"
	if lang2.Name != expName {
		t.Errorf("expected Name '%s', got '%s'", expName, lang2.Name)
	}
	expVersion = "0.16.0-dev.747+493ad58ff"
	if lang2.Release.Version != expVersion {
		t.Errorf("expected second Release Version '%s', got '%s'", expVersion, lang2.Release.Version)
	}

	expStable = false
	if lang2.Release.Stable != expStable {
		t.Errorf("expected second Stable %v, got %v", expStable, lang2.Release.Stable)
	}

	expTags = []string{"zig", "ziglang"}
	if len(lang2.Tags) != len(expTags) || lang2.Tags[0] != expTags[0] || lang2.Tags[1] != expTags[1] {
		t.Errorf("unexpected Tags for second entry: %v", lang2.Tags)
	}

	if lang2.Attributes["oop"] != "false" ||
		lang2.Attributes["procedural"] != "true" ||
		lang2.Attributes["functional"] != "false" {
		t.Errorf("unexpected Attributes for second entry: %v", lang2.Attributes)
	}
}

func TestUnmarshalStrictMissingField(t *testing.T) {
	data := support.LoadFile(t, "../support/lang_table_missing_field.csvt")

	var result []support.Lang
	opts := csvt.UnmarshalOptions{
		Strict: true,
	}

	err := csvt.UnmarshalOpts(data, &result, opts)
	if err == nil {
		t.Fatalf("expected error when strict=true and a field is missing, but got nil")
	}

	missing := csvt.IsMissingField(err)
	expField := "Release"
	if missing == nil || expField != missing.Field {
		t.Fatalf("expected MissingField error for field '%s', but got '%s'", expField, missing.Field)
	} 

	t.Logf("received expected error: %v", err)
}

func TestUnmarshalDefaultMissingField(t *testing.T) {
	data := support.LoadFile(t, "../support/lang_table_missing_field.csvt")

	var result []support.Lang
	err := csvt.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unexpected error when strict=false: %v", err)
	}

	expLen := 1
	if len(result) != expLen {
		t.Fatalf("expected %d items, got %d", expLen, len(result))
	}

	expVersion := ""
	if result[0].Release.Version != expVersion {
		t.Errorf("expected Version '%s', got '%s'", expVersion, result[0].Release.Version)
	}

	expStable := false
	if result[0].Release.Stable != expStable {
		t.Errorf("expected Stable '%v', got '%v'", expStable, result[0].Release.Stable)
	}
}