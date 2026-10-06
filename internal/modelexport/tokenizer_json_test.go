// Copyright (c) 2026 OpenWALDO Project contributors
// Copyright (c) 2026 CtrlIQ, Inc.
// Copyright (c) 2026 Gregory M. Kurtzer
// SPDX-License-Identifier: Apache-2.0

package modelexport

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestHuggingFaceTokenizerJSONDescribesSchema1ByteTokenizer(t *testing.T) {
	data, err := huggingFaceTokenizerJSON()
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		AddedTokens []struct {
			ID      int    `json:"id"`
			Content string `json:"content"`
			Special bool   `json:"special"`
		} `json:"added_tokens"`
		Normalizer    any `json:"normalizer"`
		PreTokenizer  any `json:"pre_tokenizer"`
		PostProcessor any `json:"post_processor"`
		Decoder       struct {
			Type     string `json:"type"`
			Decoders []struct {
				Type string `json:"type"`
			} `json:"decoders"`
		} `json:"decoder"`
		Model struct {
			Type         string         `json:"type"`
			ByteFallback bool           `json:"byte_fallback"`
			Vocab        map[string]int `json:"vocab"`
			Merges       []any          `json:"merges"`
		} `json:"model"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document.Model.Type != "BPE" || !document.Model.ByteFallback {
		t.Fatalf("model = %s byte_fallback=%v", document.Model.Type, document.Model.ByteFallback)
	}
	if document.Model.Merges == nil || len(document.Model.Merges) != 0 {
		t.Fatalf("merges = %#v, want empty list", document.Model.Merges)
	}
	if len(document.Model.Vocab) != 259 {
		t.Fatalf("vocabulary has %d entries, want 259", len(document.Model.Vocab))
	}
	for token, id := range map[string]int{"<pad>": 0, "<bos>": 1, "<eos>": 2, "<0x00>": 3, "<0x0A>": 13, "<0x41>": 68, "<0xFF>": 258} {
		if document.Model.Vocab[token] != id {
			t.Errorf("vocab[%s] = %d, want %d", token, document.Model.Vocab[token], id)
		}
	}
	for value := 0; value < 256; value++ {
		token := fmt.Sprintf("<0x%02X>", value)
		if id, ok := document.Model.Vocab[token]; !ok || id != value+3 {
			t.Fatalf("vocab[%s] = %d (present %v), want %d", token, id, ok, value+3)
		}
	}
	if len(document.AddedTokens) != 3 {
		t.Fatalf("added tokens = %#v", document.AddedTokens)
	}
	for index, want := range []string{"<pad>", "<bos>", "<eos>"} {
		added := document.AddedTokens[index]
		if added.ID != index || added.Content != want || !added.Special {
			t.Errorf("added token %d = %#v", index, added)
		}
	}
	if document.Normalizer != nil || document.PreTokenizer != nil || document.PostProcessor != nil {
		t.Fatal("tokenizer must not add a normalizer, pre-tokenizer, or post-processor (no implicit BOS/EOS)")
	}
	if document.Decoder.Type != "Sequence" || len(document.Decoder.Decoders) != 2 || document.Decoder.Decoders[0].Type != "ByteFallback" || document.Decoder.Decoders[1].Type != "Fuse" {
		t.Fatalf("decoder = %#v", document.Decoder)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Fatal("tokenizer.json must end with a newline")
	}
}

func TestHuggingFaceTokenizerJSONIsDeterministic(t *testing.T) {
	first, err := huggingFaceTokenizerJSON()
	if err != nil {
		t.Fatal(err)
	}
	second, _ := huggingFaceTokenizerJSON()
	if string(first) != string(second) {
		t.Fatal("tokenizer.json output changed between calls")
	}
}

func TestHuggingFaceExportRolesListTokenizerJSON(t *testing.T) {
	for _, template := range []string{"", "x"} {
		roles := huggingFaceArtifactRoles(template)
		if roles["tokenizer.json"] != "tokenizer" {
			t.Fatalf("tokenizer.json role = %q", roles["tokenizer.json"])
		}
		for name, role := range map[string]string{"tokenizer_config.json": "tokenizer", "special_tokens_map.json": "tokenizer", "tokenization_openwaldo.py": "tokenizer-code"} {
			if roles[name] != role {
				t.Errorf("%s role = %q, want %q", name, roles[name], role)
			}
		}
		if (roles["chat_template.jinja"] == "interaction-template") != (template != "") {
			t.Errorf("chat template role wrong for template %q", template)
		}
	}
}
