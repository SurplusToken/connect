package configuration

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"surplustoken-connect/internal/configedit"
	"testing"
)

func fixture(t *testing.T) (*Engine, string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "settings.json")
	e, err := New(filepath.Join(root, "backups"), map[string]string{"claude": path})
	if err != nil {
		t.Fatal(err)
	}
	return e, path
}
func write(t *testing.T, path, stringData string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(stringData), 0600); err != nil {
		t.Fatal(err)
	}
}
func TestApplyRestorePreservesExternalFieldsAndMasksSecrets(t *testing.T) {
	e, path := fixture(t)
	original := "{\r\n  // retained comment\r\n  \"theme\": \"light\",\r\n  \"env\": {\"CUSTOM\": \"keep\"}\r\n}\r\n"
	write(t, path, original)
	p, err := e.Plan("owner", "claude", []configedit.KV{{Path: "env.ANTHROPIC_AUTH_TOKEN", Value: "fixture-secret"}, {Path: "env.ANTHROPIC_MODEL", Value: "model-x"}})
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(p)
	if bytes.Contains(encoded, []byte("fixture-secret")) {
		t.Fatal("plan leaked secret")
	}
	if _, err = e.Apply("another-owner", p.ID); err == nil {
		t.Fatal("cross-session plan applied")
	}
	r, err := e.Apply("owner", p.ID)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	if !bytes.Contains(raw, []byte("// retained comment\r\n")) {
		t.Fatal("comment/CRLF lost")
	}
	raw, err = configedit.PatchJSON(raw, []configedit.KV{{Path: "newSetting", Value: "external"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	restore, err := e.Restore("owner", r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = e.Apply("owner", restore.ID); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if value(after, "env.ANTHROPIC_AUTH_TOKEN").Exists() || value(after, "env.ANTHROPIC_MODEL").Exists() {
		t.Fatal("managed fields survived restore")
	}
	if value(after, "newSetting").String() != "external" || value(after, "env.CUSTOM").String() != "keep" {
		t.Fatal("external fields lost")
	}
	history, err := e.History()
	if err != nil || len(history) != 2 {
		t.Fatalf("history: %v %d", err, len(history))
	}
	public, _ := json.Marshal(history)
	if bytes.Contains(public, []byte("fixture-secret")) {
		t.Fatal("history leaked secret")
	}
}
func TestStalePlanAndRestoreConflictNeverWrite(t *testing.T) {
	e, path := fixture(t)
	write(t, path, `{"model":"old"}`)
	p, err := e.Plan("x", "claude", []configedit.KV{{Path: "model", Value: "new"}})
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, `{"model":"external"}`)
	if _, err = e.Apply("x", p.ID); err == nil {
		t.Fatal("stale plan applied")
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != `{"model":"external"}` {
		t.Fatal("stale write changed file")
	}
	p, _ = e.Plan("x", "claude", []configedit.KV{{Path: "model", Value: "new"}})
	r, err := e.Apply("x", p.ID)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, `{"model":"external-again"}`)
	if _, err = e.Restore("x", r.ID); err == nil {
		t.Fatal("restored over external managed field")
	}
}
func TestRecoverCommittedPendingAndRestoreMissingOriginal(t *testing.T) {
	e, path := fixture(t)
	p, _ := e.Plan("x", "claude", []configedit.KV{{Path: "model", Value: "new"}})
	r, err := e.Apply("x", p.ID)
	if err != nil {
		t.Fatal(err)
	}
	d, err := e.load(r.ID)
	if err != nil {
		t.Fatal(err)
	}
	d.State = "pending"
	if err = e.save(d); err != nil {
		t.Fatal(err)
	}
	reopened, err := New(e.Store, map[string]string{"claude": path})
	if err != nil {
		t.Fatal(err)
	}
	records, err := reopened.History()
	if err != nil || records[0].State != "applied" {
		t.Fatal("pending commit not recovered")
	}
	rest, err := reopened.Restore("x", r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = reopened.Apply("x", rest.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("new unmodified file not removed on restore")
	}
}
func TestInvalidJSONBackupFailureAndSymlink(t *testing.T) {
	e, path := fixture(t)
	write(t, path, `{broken`)
	if _, err := e.Plan("x", "claude", []configedit.KV{{Path: "model", Value: "new"}}); err == nil {
		t.Fatal("invalid JSON planned")
	}
	write(t, path, `{"env":"not-object"}`)
	if _, err := e.Plan("x", "claude", []configedit.KV{{Path: "env.key", Value: "new"}}); err == nil {
		t.Fatal("scalar ancestor overwritten")
	}
	write(t, path, `{"model":"old"}`)
	p, _ := e.Plan("x", "claude", []configedit.KV{{Path: "model", Value: "new"}})
	store := e.Store
	e.Store = filepath.Join(path, "not-a-directory")
	if _, err := e.Apply("x", p.ID); err == nil {
		t.Fatal("backup failure still applied")
	}
	e.Store = store
	raw, _ := os.ReadFile(path)
	if string(raw) != `{"model":"old"}` {
		t.Fatal("backup failure changed target")
	}
	link := filepath.Join(filepath.Dir(path), "link.json")
	if err := os.Symlink(path, link); err != nil {
		t.Skip(err)
	}
	e.Paths["claude"] = link
	if _, err := e.Plan("x", "claude", []configedit.KV{{Path: "model", Value: "new"}}); err == nil {
		t.Fatal("symlink accepted")
	}
	if _, err := e.Plan("x", "codex", nil); err == nil {
		t.Fatal("Codex unprotected")
	}
}
func TestIdempotenceAndExactRestore(t *testing.T) {
	e, path := fixture(t)
	original := "{\n  // own\n  \"model\": \"old\"\n}\n"
	write(t, path, original)
	p, _ := e.Plan("x", "claude", []configedit.KV{{Path: "model", Value: "new"}})
	r, err := e.Apply("x", p.ID)
	if err != nil {
		t.Fatal(err)
	}
	again, err := e.Plan("x", "claude", []configedit.KV{{Path: "model", Value: "new"}})
	if err != nil || !again.Noop {
		t.Fatal("not idempotent")
	}
	rest, err := e.Restore("x", r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = e.Apply("x", rest.ID); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	if strings.Compare(string(raw), original) != 0 {
		t.Fatal("unmodified original bytes not restored")
	}
}

func TestProviderNameNotMaskedAndStructuredValueNotExposed(t *testing.T) {
	e, path := fixture(t)
	write(t, path, `{"provider":{"surplustoken":{"name":{"private":"fixture-secret"}}}}`)
	p, err := e.Plan("x", "claude", []configedit.KV{{Path: "provider.surplustoken.name", Value: "SurplusToken"}, {Path: "provider.surplustoken.options.apiKey", Value: "new-fixture-secret"}})
	if err != nil {
		t.Fatal(err)
	}
	if p.Changes[0].After != "SurplusToken" {
		t.Fatal("non-secret provider name masked")
	}
	raw, _ := json.Marshal(p)
	if bytes.Contains(raw, []byte("fixture-secret")) {
		t.Fatal("structured value exposed")
	}
}
func TestProtectedCodexHomeCannotBeAliasTarget(t *testing.T) {
	e, _ := fixture(t)
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", root)
	e.Paths["claude"] = filepath.Join(root, "auth.json")
	if _, err = e.Plan("x", "claude", []configedit.KV{{Path: "model", Value: "new"}}); err == nil {
		t.Fatal("protected home accepted under another client")
	}
}

func TestDuplicateKeysAndEmptyExistingFileRefused(t *testing.T) {
	e, path := fixture(t)
	for _, raw := range []string{`{"model":"first","model":"last"}`, `{"env":{"key":"first","key":"last"}}`, ""} {
		write(t, path, raw)
		if _, err := e.Plan("x", "claude", []configedit.KV{{Path: "model", Value: "new"}}); err == nil {
			t.Fatal("ambiguous or invalid existing config accepted")
		}
	}
}
