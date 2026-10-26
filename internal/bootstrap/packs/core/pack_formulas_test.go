package core

import (
	"io/fs"
	"strings"
	"testing"
)

// TestCoreFormulasAnchorWorktreeCreationToAgentHome guards against the
// regression from gastownhall/gascity#6014: a formula that creates the
// per-bead worktree under $(pwd) plants it wherever the session happens to
// be standing — which can be the tracked rig checkout, leaving it dirty and
// breaking drift detection. Worktree creation must anchor to the agent
// home ($GC_DIR), never to the process's current directory.
func TestCoreFormulasAnchorWorktreeCreationToAgentHome(t *testing.T) {
	formulas, err := fs.Glob(PackFS, "formulas/*.toml")
	if err != nil {
		t.Fatalf("Glob(formulas): %v", err)
	}
	if len(formulas) == 0 {
		t.Fatal("core pack embeds no formulas")
	}
	for _, path := range formulas {
		data, err := fs.ReadFile(PackFS, path)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", path, err)
		}
		if strings.Contains(string(data), "$(pwd)/worktrees") {
			t.Fatalf("%s derives the per-bead worktree path from $(pwd); anchor it to the agent home via $GC_DIR instead", path)
		}
	}
}

// TestCoreWorktreeFormulasUseGCDirAnchor pins the fix for the two formulas
// known to create per-bead worktrees: both must resolve the anchor from
// $GC_DIR so the worktree lands under the agent home regardless of where
// the session last cd-ed.
func TestCoreWorktreeFormulasUseGCDirAnchor(t *testing.T) {
	for _, path := range []string{
		"formulas/mol-polecat-commit.toml",
		"formulas/mol-scoped-work.toml",
	} {
		data, err := fs.ReadFile(PackFS, path)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", path, err)
		}
		if !strings.Contains(string(data), `AGENT_HOME="${GC_DIR:-}"`) {
			t.Fatalf("%s must anchor worktree creation to the agent home via $GC_DIR", path)
		}
	}
}

// TestCorePolecatCommitCleanupKeepsCwdWorktree pins the cleanup-step idiom:
// commit-and-push runs with the agent already inside the worktree, so the
// bare WORKTREE_PATH=$(pwd) there is the intended value and must not be
// "fixed" by a search-and-replace across $(pwd).
func TestCorePolecatCommitCleanupKeepsCwdWorktree(t *testing.T) {
	data, err := fs.ReadFile(PackFS, "formulas/mol-polecat-commit.toml")
	if err != nil {
		t.Fatalf("ReadFile(formulas/mol-polecat-commit.toml): %v", err)
	}
	if !strings.Contains(string(data), "WORKTREE_PATH=$(pwd)\n") {
		t.Fatal("mol-polecat-commit cleanup step must keep WORKTREE_PATH=$(pwd): the cleanup runs inside the worktree")
	}
}
