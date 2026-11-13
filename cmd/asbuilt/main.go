// Command asbuilt reports the as-built truth of a server.
//
// It collects what the machine actually is, judges it against the rules that
// matter for data centre hardware, and prints the result as text or JSON.
// There is no daemon, no database, no configuration file and no telemetry.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Balanov/asbuilt/internal/collect"
	"github.com/Balanov/asbuilt/internal/diff"
	"github.com/Balanov/asbuilt/internal/profile"
	"github.com/Balanov/asbuilt/internal/report"
	"github.com/Balanov/asbuilt/internal/rules"
	"github.com/Balanov/asbuilt/internal/sysfs"
	"github.com/Balanov/asbuilt/internal/version"
)

const usage = `asbuilt - the as-built truth of a server

Usage:
  asbuilt [check] [flags]        audit this machine
  asbuilt manifest [flags]       print the canonical JSON manifest
  asbuilt diff A.json B.json     compare two manifests
  asbuilt rules                  list the rule catalog
  asbuilt explain RULE_ID        explain one rule

Flags:
  --json            machine-readable output
  --profile FILE    check the machine against an expected profile (JSON)
  --root DIR        read the machine under DIR instead of / (fixtures, containers)
  --fail-on LEVEL   exit 1 when findings reach LEVEL (critical|warn|info|none)
  --version         print the version
`

func main() {
	os.Exit(run(os.Args[1:]))
}

// flagsWithValue are the flags that consume the argument after them.
var flagsWithValue = map[string]bool{
	"profile": true,
	"root":    true,
	"fail-on": true,
}

// splitCommand pulls the subcommand out of an argument list that may hold
// flags before it, after it, or on both sides.
//
// The standard flag package stops at the first non-flag argument, which
// would turn `asbuilt --root /srv manifest` into a plain check. The command
// is therefore extracted first, and the flag values it swallows are taken
// with it.
func splitCommand(args []string) (command string, rest []string) {
	command = "check"
	found := false
	rest = make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			if !found {
				command = arg
				found = true
				continue
			}
			rest = append(rest, arg)
			continue
		}

		rest = append(rest, arg)
		name := strings.TrimLeft(arg, "-")
		if index := strings.Index(name, "="); index >= 0 {
			continue
		}
		if flagsWithValue[name] && i+1 < len(args) {
			rest = append(rest, args[i+1])
			i++
		}
	}

	return command, rest
}

func run(args []string) int {
	if len(args) > 0 {
		switch args[0] {
		case "-h", "--help", "help":
			fmt.Print(usage)
			return 0
		case "--version", "version":
			fmt.Println("asbuilt " + version.Version)
			return 0
		}
	}

	command, args := splitCommand(args)

	switch command {
	case "-h", "--help", "help":
		fmt.Print(usage)
		return 0
	case "--version", "version":
		fmt.Println("asbuilt " + version.Version)
		return 0
	case "check":
		return cmdCheck(args)
	case "manifest":
		return cmdManifest(args)
	case "diff":
		return cmdDiff(args)
	case "rules":
		return cmdRules(args)
	case "explain":
		return cmdExplain(args)
	default:
		fmt.Fprintf(os.Stderr, "asbuilt: unknown command %q\n\n%s", command, usage)
		return 2
	}
}

type commonFlags struct {
	json    bool
	profile string
	root    string
	failOn  string
}

func addFlags(fs *flag.FlagSet, c *commonFlags) {
	fs.BoolVar(&c.json, "json", false, "machine-readable output")
	fs.StringVar(&c.profile, "profile", "", "expected profile (JSON)")
	fs.StringVar(&c.root, "root", "", "read the machine under this directory instead of /")
	fs.StringVar(&c.failOn, "fail-on", "none", "exit 1 when findings reach this severity")
}

func cmdCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	var c commonFlags
	addFlags(fs, &c)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	m, prof, err := gather(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "asbuilt: %v\n", err)
		return 2
	}

	findings := rules.Run(&rules.Context{Manifest: m, Profile: prof})

	if c.json {
		out := struct {
			Manifest *collect.Manifest `json:"manifest"`
			Findings []rules.Finding   `json:"findings"`
			Counts   map[string]int    `json:"counts"`
		}{Manifest: m, Findings: findings, Counts: counts(findings)}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			fmt.Fprintf(os.Stderr, "asbuilt: %v\n", err)
			return 2
		}
	} else {
		fmt.Print(report.Manifest(m))
		fmt.Println()
		fmt.Print(report.Findings(findings))
	}

	return exitCode(findings, c.failOn)
}

func cmdManifest(args []string) int {
	fs := flag.NewFlagSet("manifest", flag.ContinueOnError)
	var c commonFlags
	addFlags(fs, &c)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	m, _, err := gather(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "asbuilt: %v\n", err)
		return 2
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		fmt.Fprintf(os.Stderr, "asbuilt: %v\n", err)
		return 2
	}
	return 0
}

func cmdDiff(args []string) int {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	var c commonFlags
	fs.BoolVar(&c.json, "json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	rest := fs.Args()
	if len(rest) != 2 {
		fmt.Fprintf(os.Stderr, "asbuilt: diff needs exactly two manifest files\n")
		return 2
	}
	a, err := loadManifest(rest[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "asbuilt: %v\n", err)
		return 2
	}
	b, err := loadManifest(rest[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "asbuilt: %v\n", err)
		return 2
	}
	diffs := diff.Manifests(a, b)
	if c.json {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(diffs); err != nil {
			fmt.Fprintf(os.Stderr, "asbuilt: %v\n", err)
			return 2
		}
	} else {
		fmt.Print(report.Diff(diffs))
	}
	if len(diffs) > 0 {
		return 1
	}
	return 0
}

func cmdRules(args []string) int {
	fs := flag.NewFlagSet("rules", flag.ContinueOnError)
	var c commonFlags
	fs.BoolVar(&c.json, "json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	catalog := rules.Catalog()
	if c.json {
		type ruleJSON struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Severity string `json:"severity"`
			Source   string `json:"source,omitempty"`
			Remedy   string `json:"remedy,omitempty"`
		}
		out := make([]ruleJSON, 0, len(catalog))
		for _, r := range catalog {
			out = append(out, ruleJSON{r.ID, r.Title, string(r.Severity), r.Source, r.Remedy})
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			fmt.Fprintf(os.Stderr, "asbuilt: %v\n", err)
			return 2
		}
		return 0
	}
	fmt.Print(report.Catalog(catalog))
	return 0
}

func cmdExplain(args []string) int {
	fs := flag.NewFlagSet("explain", flag.ContinueOnError)
	var c commonFlags
	fs.BoolVar(&c.json, "json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintf(os.Stderr, "asbuilt: explain needs exactly one rule id\n")
		return 2
	}
	for _, r := range rules.Catalog() {
		if !strings.EqualFold(r.ID, rest[0]) {
			continue
		}
		if c.json {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(map[string]string{
				"id": r.ID, "title": r.Title, "severity": string(r.Severity),
				"source": r.Source, "remedy": r.Remedy,
			})
			return 0
		}
		fmt.Printf("%s\n  %s\n  severity: %s\n", r.ID, r.Title, r.Severity)
		if r.Remedy != "" {
			fmt.Printf("  remedy: %s\n", r.Remedy)
		}
		if r.Source != "" {
			fmt.Printf("  source: %s\n", r.Source)
		}
		return 0
	}
	fmt.Fprintf(os.Stderr, "asbuilt: no rule with id %q\n", rest[0])
	return 2
}

func gather(c commonFlags) (*collect.Manifest, *profile.Profile, error) {
	m := collect.Gather(sysfs.New(c.root))

	var prof *profile.Profile
	if c.profile != "" {
		raw, err := os.ReadFile(c.profile)
		if err != nil {
			return nil, nil, fmt.Errorf("profile %s: %w", c.profile, err)
		}
		var p profile.Profile
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, nil, fmt.Errorf("profile %s: %w", c.profile, err)
		}
		prof = &p
	}
	return m, prof, nil
}

func loadManifest(path string) (*collect.Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var m collect.Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &m, nil
}

func counts(findings []rules.Finding) map[string]int {
	out := map[string]int{"critical": 0, "warn": 0, "info": 0}
	for _, f := range findings {
		out[string(f.Severity)]++
	}
	return out
}

func exitCode(findings []rules.Finding, failOn string) int {
	threshold, ok := map[string]rules.Severity{
		"critical": rules.Critical,
		"warn":     rules.Warn,
		"info":     rules.Info,
	}[failOn]
	if !ok {
		return 0
	}
	ids := make([]string, 0, len(findings))
	for _, f := range findings {
		ids = append(ids, f.ID)
	}
	sort.Strings(ids)
	for _, f := range findings {
		if f.Severity.Rank() >= threshold.Rank() {
			return 1
		}
	}
	return 0
}
