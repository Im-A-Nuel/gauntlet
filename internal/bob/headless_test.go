package bob

import (
	"bytes"
	"strings"
	"testing"
)

func TestDefaultArgsReferencesSurvivorsFile(t *testing.T) {
	args := DefaultArgs(".gauntlet/survivors.md")
	if len(args) != 4 || args[0] != "run" || args[1] != "--format" || args[2] != "json" {
		t.Fatalf("DefaultArgs = %v, want [run --format json <prompt>]", args)
	}
	prompt := args[len(args)-1]
	if !strings.Contains(prompt, "@.gauntlet/survivors.md") {
		t.Fatalf("prompt = %q, want it to reference the survivors file via @path", prompt)
	}
	if !strings.Contains(prompt, "strengthen-tests") {
		t.Fatalf("prompt = %q, want it to name the strengthen-tests skill", prompt)
	}
}

func TestAdapterInvokeRejectsEmptyExecutable(t *testing.T) {
	a := Adapter{}
	if err := a.Invoke(t.TempDir(), &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatalf("Invoke with no executable = nil error, want refusal")
	}
}

// These exercise the adapter's process-launching mechanics (argument-array
// exec, no shell, error propagation) against the `go` binary as a stand-in,
// since IBM Bob is not installed in this environment. They do not verify
// real Bob CLI compatibility — see package doc comment in headless.go.
func TestAdapterInvokeSucceedsOnKnownGoodCommand(t *testing.T) {
	a := Adapter{Executable: "go", Args: []string{"version"}}
	var out bytes.Buffer
	if err := a.Invoke(t.TempDir(), &out, &out); err != nil {
		t.Fatalf("Invoke: %v (output: %s)", err, out.String())
	}
	if !strings.Contains(out.String(), "go version") {
		t.Fatalf("output = %q, want it to contain 'go version'", out.String())
	}
}

func TestAdapterInvokeReportsFailureOnBadSubcommand(t *testing.T) {
	a := Adapter{Executable: "go", Args: []string{"definitely-not-a-real-subcommand"}}
	var out bytes.Buffer
	if err := a.Invoke(t.TempDir(), &out, &out); err == nil {
		t.Fatalf("Invoke with bad subcommand = nil error, want failure")
	}
}
