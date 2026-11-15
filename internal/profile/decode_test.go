package profile

import (
	"os"
	"strings"
	"testing"
)

func TestParseUnknownFieldPath(t *testing.T) {
	_, err := Parse([]byte("{\"expect\":{\"memory\":{\"configured_mts_x\":6400}}}"))
	if err == nil || !strings.Contains(err.Error(), "expect.memory.configured_mts_x") {
		t.Fatalf("got %v", err)
	}
}
func TestParseShippedProfile(t *testing.T) {
	data, err := os.ReadFile("../../profiles/xd-node.json")
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(data)
	if err != nil || p.Expect.Memory == nil {
		t.Fatalf("profile=%+v err=%v", p, err)
	}
}
func TestParseRejectsTrailingJSON(t *testing.T) {
	if _, err := Parse([]byte("{} {}")); err == nil {
		t.Fatal("accepted two profiles")
	}
}
