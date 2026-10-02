// Copyright (c) 2026 OpenWALDO Project contributors
// Copyright (c) 2026 CtrlIQ, Inc.
// Copyright (c) 2026 Gregory M. Kurtzer
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"strings"
	"testing"

	waldotokenizer "github.com/openwaldo/waldo/internal/tokenizer"
)

func TestTokenizerCompressionGateUsesBytesPerTokenAsHigherIsBetter(t *testing.T) {
	baseline := waldotokenizer.Comparison{Tokenizer: "r50k", BytesPerToken: 4}
	if err := validateTokenizerCompression([]waldotokenizer.Comparison{baseline, {Tokenizer: "candidate", BytesPerToken: 3.6}}, 0.10); err != nil {
		t.Fatal(err)
	}
	if err := validateTokenizerCompression([]waldotokenizer.Comparison{baseline, {Tokenizer: "candidate", BytesPerToken: 4.5}}, 0.10); err != nil {
		t.Fatal(err)
	}
	err := validateTokenizerCompression([]waldotokenizer.Comparison{baseline, {Tokenizer: "candidate", BytesPerToken: 3.59}}, 0.10)
	if err == nil || !strings.Contains(err.Error(), "worse than r50k") {
		t.Fatalf("compression regression error = %v", err)
	}
}
