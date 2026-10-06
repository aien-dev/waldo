// Copyright (c) 2026 OpenWALDO Project contributors
// Copyright (c) 2026 CtrlIQ, Inc.
// Copyright (c) 2026 Gregory M. Kurtzer
// SPDX-License-Identifier: Apache-2.0

package modelexport

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// huggingFaceTokenizerJSON describes the schema-1 WALDO byte tokenizer in the
// standard Hugging Face tokenizers format: <pad>=0, <bos>=1, <eos>=2, and byte
// b is id b+3 spelled "<0xHH>". There are no merges and no implicit BOS or EOS.
func huggingFaceTokenizerJSON() ([]byte, error) {
	specials := []string{"<pad>", "<bos>", "<eos>"}
	added := make([]map[string]any, 0, len(specials))
	var vocabulary bytes.Buffer
	vocabulary.WriteString("{")
	tokens := append([]string{}, specials...)
	for value := 0; value < 256; value++ {
		tokens = append(tokens, fmt.Sprintf("<0x%02X>", value))
	}
	for id, token := range tokens {
		if id > 0 {
			vocabulary.WriteString(",")
		}
		key, err := json.Marshal(token)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(&vocabulary, "%s:%d", key, id)
	}
	vocabulary.WriteString("}")
	for id, token := range specials {
		added = append(added, map[string]any{
			"id": id, "content": token, "single_word": false, "lstrip": false,
			"rstrip": false, "normalized": false, "special": true,
		})
	}
	document := struct {
		Version       string           `json:"version"`
		Truncation    any              `json:"truncation"`
		Padding       any              `json:"padding"`
		AddedTokens   []map[string]any `json:"added_tokens"`
		Normalizer    any              `json:"normalizer"`
		PreTokenizer  any              `json:"pre_tokenizer"`
		PostProcessor any              `json:"post_processor"`
		Decoder       any              `json:"decoder"`
		Model         any              `json:"model"`
	}{
		Version:     "1.0",
		AddedTokens: added,
		Decoder: map[string]any{"type": "Sequence", "decoders": []any{
			map[string]any{"type": "ByteFallback"}, map[string]any{"type": "Fuse"},
		}},
		Model: struct {
			Type                    string          `json:"type"`
			Dropout                 any             `json:"dropout"`
			UnkToken                any             `json:"unk_token"`
			ContinuingSubwordPrefix any             `json:"continuing_subword_prefix"`
			EndOfWordSuffix         any             `json:"end_of_word_suffix"`
			FuseUnk                 bool            `json:"fuse_unk"`
			ByteFallback            bool            `json:"byte_fallback"`
			IgnoreMerges            bool            `json:"ignore_merges"`
			Vocab                   json.RawMessage `json:"vocab"`
			Merges                  []string        `json:"merges"`
		}{Type: "BPE", ByteFallback: true, Vocab: vocabulary.Bytes(), Merges: []string{}},
	}
	data, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}
	var indented bytes.Buffer
	if err := json.Indent(&indented, data, "", "  "); err != nil {
		return nil, err
	}
	indented.WriteByte('\n')
	return indented.Bytes(), nil
}

func huggingFaceArtifactRoles(interactionTemplate string) map[string]string {
	roles := map[string]string{
		"model.safetensors": "weights", "config.json": "configuration",
		"generation_config.json": "generation-configuration",
		"tokenizer_config.json":  "tokenizer", "special_tokens_map.json": "tokenizer",
		"tokenizer.json":            "tokenizer",
		"tokenization_openwaldo.py": "tokenizer-code", "architecture.py": "architecture-code",
		"README.md": "documentation", "EU-BOM.json": "regulatory-disclosure",
	}
	if interactionTemplate != "" {
		roles["chat_template.jinja"] = "interaction-template"
	}
	return roles
}
