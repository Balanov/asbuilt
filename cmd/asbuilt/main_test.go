package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitCommand(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		command string
		rest    []string
	}{
		{"no arguments", nil, "check", []string{}},
		{"plain check", []string{"check"}, "check", []string{}},
		{"manifest alone", []string{"manifest"}, "manifest", []string{}},
		{"flags after the command", []string{"manifest", "--root", "/srv"}, "manifest", []string{"--root", "/srv"}},
		{"flags before the command", []string{"--root", "/srv", "manifest"}, "manifest", []string{"--root", "/srv"}},
		{"flags on both sides", []string{"--json", "check", "--root", "/srv"}, "check", []string{"--json", "--root", "/srv"}},
		{"positional arguments", []string{"diff", "a.json", "b.json"}, "diff", []string{"a.json", "b.json"}},
		{"positional after flags", []string{"diff", "--json", "a.json", "b.json"}, "diff", []string{"--json", "a.json", "b.json"}},
		{"value with equals sign", []string{"--root=/srv", "manifest"}, "manifest", []string{"--root=/srv"}},
		{"profile value is not a command", []string{"--profile", "check", "manifest"}, "manifest", []string{"--profile", "check"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			command, rest := splitCommand(testCase.args)
			if command != testCase.command {
				t.Errorf("command = %q, want %q", command, testCase.command)
			}
			if strings.Join(rest, " ") != strings.Join(testCase.rest, " ") {
				t.Errorf("rest = %v, want %v", rest, testCase.rest)
			}
		})
	}
}

func TestGatherRejectsUnknownProfileFields(t *testing.T) {
	for _, input := range []string{
		"{\"unexpected\":true}",
		"{\"expect\":{\"unexpected\":true}}",
		"{\"expect\":{\"memory\":{\"configured_mts_x\":6400}}}",
	} {
		path := filepath.Join(t.TempDir(), "profile.json")
		if err := os.WriteFile(path, []byte(input), 0600); err != nil {
			t.Fatal(err)
		}
		_, _, err := gather(commonFlags{root: t.TempDir(), profile: path})
		if err == nil || !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "unknown field") {
			t.Fatalf("expected an unknown-field error naming %s, got %v", path, err)
		}
	}
}
