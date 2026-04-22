package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/steveyegge/gastown/internal/polecat"
)

func TestSessionInfoJSONOutput(t *testing.T) {
	info := &polecat.SessionInfo{
		Polecat:   "alpha",
		SessionID: "gt-alpha",
		Running:   true,
		RigName:   "gastown",
		Attached:  false,
		Created:   time.Date(2026, 2, 20, 10, 0, 0, 0, time.UTC),
		Windows:   1,
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if parsed["polecat"] != "alpha" {
		t.Errorf("polecat = %v, want alpha", parsed["polecat"])
	}
	if parsed["session_id"] != "gt-alpha" {
		t.Errorf("session_id = %v, want gt-alpha", parsed["session_id"])
	}
	if parsed["running"] != true {
		t.Errorf("running = %v, want true", parsed["running"])
	}
	if parsed["rig_name"] != "gastown" {
		t.Errorf("rig_name = %v, want gastown", parsed["rig_name"])
	}
}

func TestSessionStatusCmdJSONFlagWiring(t *testing.T) {
	// Verify --json flag is registered on the session status command.
	// This catches regressions where flag binding is accidentally removed,
	// which would silently break formulas that depend on --json output.
	f := sessionStatusCmd.Flags().Lookup("json")
	if f == nil {
		t.Fatal("session status command missing --json flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--json default = %q, want \"false\"", f.DefValue)
	}
}

func TestSessionInfoJSONOutputNotRunning(t *testing.T) {
	info := &polecat.SessionInfo{
		Polecat:   "beta",
		SessionID: "gt-beta",
		Running:   false,
		RigName:   "testrig",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if parsed["running"] != false {
		t.Errorf("running = %v, want false", parsed["running"])
	}
}

func TestRecoverableSessionIssue(t *testing.T) {
	tests := []struct {
		name          string
		state         polecat.State
		polecatIssue  string
		polecatBranch string
		want          string
	}{
		{
			name:          "active polecat issue wins over branch",
			state:         polecat.StateWorking,
			polecatIssue:  "tg-live",
			polecatBranch: "polecat/furiosa/tg-branch@mk123",
			want:          "tg-live",
		},
		{
			name:          "issue branch rescues missing issue",
			state:         polecat.StateWorking,
			polecatBranch: "polecat/furiosa/tg-branch@mk123",
			want:          "tg-branch",
		},
		{
			name:          "modern timestamp branch has no issue",
			state:         polecat.StateWorking,
			polecatBranch: "polecat/furiosa-mk123",
			want:          "",
		},
		{
			name:          "invalid polecat branch segment is ignored",
			state:         polecat.StateWorking,
			polecatBranch: "polecat/furiosa/not-a-bead@mk123",
			want:          "",
		},
		{
			name:          "idle polecat does not recover stale branch issue",
			state:         polecat.StateIdle,
			polecatBranch: "polecat/furiosa/tg-branch@mk123",
			want:          "",
		},
		{
			name:          "non polecat branch is ignored",
			state:         polecat.StateWorking,
			polecatBranch: "codex/gamejam-webgpu-rig",
			want:          "",
		},
		{
			name:  "empty when no source exists",
			state: polecat.StateWorking,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := recoverableSessionIssue(tt.state, tt.polecatIssue, tt.polecatBranch)
			if got != tt.want {
				t.Errorf("recoverableSessionIssue() = %q, want %q", got, tt.want)
			}
		})
	}
}
