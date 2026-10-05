# Tiny Shakespeare control

This is a focused end-to-end language-model control using the exact
1,115,394-byte Tiny Shakespeare text distributed by `karpathy/char-rnn`. It
tests WALDO's canonical ingestion, byte tokenization, contiguous packing,
held-out evaluation, checkpoint selection, reload, and generation without the
data-mixture ambiguity of the general-foundation ladder.

## Recipe

- Corpus: `core/reference/tiny-shakespeare` only.
- Tokenizer: built-in byte tokenizer. Tiny Shakespeare is ASCII, so its text
  bytes correspond to the character-level workload.
- Model: 6 layers, width 384, 6 attention heads, approximately 10.7M
  parameters.
- Context: 256 tokens.
- Budget: 81.92M tokens, exactly 5,000 global optimizer steps at batch 64.
- Evaluation: `contiguous-tail-v1` reserves the final 10% of the single
  canonical record. The first 90% remains in original order for training.
- Optimizer: AdamW, peak learning rate 1e-3, cosine schedule, 100 warmup steps.

The contiguous split is essential. Splitting the source into lines would add
an artificial EOS boundary after every line, while record-level evaluation
would otherwise hold out nothing because the corpus contains one record.

## Run

```console
go run ./cmd/waldo/ model forecast \
  composes/tiny-shakespeare/0001-tiny-shakespeare-control.yaml

go run ./cmd/waldo/ model train tiny-shakespeare-control-01 \
  composes/tiny-shakespeare/0001-tiny-shakespeare-control.yaml

./composes/tiny-shakespeare/evaluate-tiny-shakespeare.sh \
  tiny-shakespeare-control-01 \
  /tmp/tiny-shakespeare-control-01-eval.jsonl
```

The compose uses `parallelism: auto`, so the same command is valid locally.
Add `--hostfile ~/hostfile` only when intentionally running across the GPU
hosts.

## Success criteria

1. Forecast reports roughly 10.7M parameters, 81.92M requested tokens, and a
   runnable local backend.
2. Preflight reports one included record, one partially held-out record, and a
   held-out set of approximately 111.5 KiB.
3. Held-out loss is finite, falls materially from initialization, and does not
   turn upward persistently near the selected checkpoint.
4. At least six of eight sampled continuations preserve play-like formatting,
   speaker turns, and locally coherent English for at least 100 tokens.
5. No more than two of eight samples collapse into an immediate repeated line
   or phrase loop.

EOS is recorded but is not a promotion gate. This corpus contains one long
document and therefore supplies only one natural EOS example per corpus pass.
