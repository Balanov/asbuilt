package rules

import (
	"os"
	"strings"
	"testing"
)

func TestRuleReferenceMatchesCatalog(t *testing.T) {
	data, err := os.ReadFile("../../docs/rules.md")
	if err != nil {
		t.Fatal(err)
	}
	documented := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Split(line, "|")
		if len(fields) < 3 {
			continue
		}
		id := strings.TrimSpace(fields[1])
		if !strings.HasPrefix(id, "`") {
			continue
		}
		id = strings.Trim(id, "`")
		if documented[id] {
			t.Errorf("duplicate rule %s", id)
		}
		documented[id] = true
	}
	for _, rule := range Catalog() {
		if !documented[rule.ID] {
			t.Errorf("missing rule %s", rule.ID)
		}
		delete(documented, rule.ID)
	}
	for id := range documented {
		t.Errorf("unknown rule %s", id)
	}
}
