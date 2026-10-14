package httpapi

import (
	"github.com/tidwall/gjson"
	"strings"
	"surplustoken-connect/internal/configedit"
	"surplustoken-connect/internal/gateway"
	"testing"
)

func TestOpenCodeV2TemplatePreservesModelChoices(t *testing.T) {
	patches := openCodeV2Patches(gateway.Key{Value: "test-key-secret"}, gateway.Group{Platform: "openai"}, []string{"gpt-5.5", "other/model.v2"})
	raw, err := configedit.PatchJSON([]byte("{\n // user comment\n \"model\":\"original/default\", \"agents\":{\"title\":{\"model\":\"original/small\"},\"custom\":{\"model\":\"keep/model\"}}\n}"), patches, nil)
	if err != nil {
		t.Fatal(err)
	}
	checks := map[string]string{"providers.surplustoken.package": "@opencode/ai/providers/openai-compatible/responses", "providers.surplustoken.settings.apiKey": "test-key-secret", "model": "original/default", "agents.title.model": "original/small", "agents.custom.model": "keep/model", "providers.surplustoken.models.other/model\\.v2.modelID": "other/model.v2"}
	for field, want := range checks {
		if value := gjson.GetBytes(raw, field).String(); value != want {
			t.Fatalf("%s mismatch", field)
		}
	}
	if !strings.Contains(string(raw), "// user comment") || gjson.GetBytes(raw, "provider").Exists() || gjson.GetBytes(raw, "small_model").Exists() {
		t.Fatal("legacy format written or comments lost")
	}
	if got := openCodeV2Patches(gateway.Key{}, gateway.Group{Platform: "composite"}, nil); got != nil {
		t.Fatal("ambiguous composite protocol guessed")
	}
}

func TestOpenCodeTemplateDoesNotAddDefaultOrTitleModel(t *testing.T) {
	patches := openCodeV2Patches(gateway.Key{Value: "fixture"}, gateway.Group{Platform: "openai"}, []string{"gpt-5.5"})
	raw, err := configedit.PatchJSON([]byte("{}"), patches, nil)
	if err != nil {
		t.Fatal(err)
	}
	if gjson.GetBytes(raw, "model").Exists() || gjson.GetBytes(raw, "agents").Exists() {
		t.Fatal("template assigned model choices")
	}
	if gjson.GetBytes(raw, "providers.surplustoken.settings.transport").String() != "http" {
		t.Fatal("gateway template must use HTTP")
	}
}
