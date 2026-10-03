# TinyStories controlled reproduction

This ladder asks a narrower question than the general-foundation ladder: can
WALDO reproduce the small-model behavior reported by TinyStories on a tightly
controlled synthetic distribution? It is a functional reproduction, not a
bit-for-bit copy of the paper's GPT-Neo implementation.

The experiment holds the dataset, tokenizer recipe, context, optimizer, and
evaluation prompts fixed while increasing training horizon and then model
capacity. Stop at the first failed gate and diagnose it before advancing.

## Controlled differences from the paper

- WALDO uses its decoder-transformer implementation, grouped-query attention,
  RoPE/RMSNorm, and gated MLP instead of GPT-Neo local/global attention.
- WALDO trains a deterministic 10,000-entry byte BPE from the pinned corpus.
- WALDO creates a 512-record held-out set from the train split. The publisher's
  validation split is not ingested for this experiment.
- The paper reports constant learning rate after warmup. WALDO currently uses
  warmup-stable-warmdown, so each compose pins the same 5e-4 peak rate and a
  proportional warmdown.
- The inexpensive promotion suite uses one deterministic completion per prompt.
  It does not reproduce the paper's GPT-4 evaluation of ten temperature-1
  completions for each of roughly 50 prompts.

These differences mean success validates WALDO's small-model training path and
data hypothesis. It does not claim an exact reproduction of a published score.

## Pinned ingestion

The source is `roneneldan/TinyStories` revision
`7dfb18bd830e361f237a94bcdb466bc6e7b09050`, licensed under
CDLA-Sharing-1.0. Download exactly the four train shards into a clean handoff:

```console
handoff="$(mktemp -d)"
mkdir -p "$handoff/data"
base="https://huggingface.co/datasets/roneneldan/TinyStories/resolve/7dfb18bd830e361f237a94bcdb466bc6e7b09050/data"

curl --fail --location "$base/train-00000-of-00004-2d5a1467fff1081b.parquet" \
  --output "$handoff/data/train-00000-of-00004-2d5a1467fff1081b.parquet"
curl --fail --location "$base/train-00001-of-00004-5852b56a2bd28fd9.parquet" \
  --output "$handoff/data/train-00001-of-00004-5852b56a2bd28fd9.parquet"
curl --fail --location "$base/train-00002-of-00004-a26307300439e943.parquet" \
  --output "$handoff/data/train-00002-of-00004-a26307300439e943.parquet"
curl --fail --location "$base/train-00003-of-00004-d243063613e5a057.parquet" \
  --output "$handoff/data/train-00003-of-00004-d243063613e5a057.parquet"

cp composes/tinystories/tinystories-manifest.json "$handoff/manifest.json"
sha256sum "$handoff"/data/*.parquet
go run ./cmd/waldo/ index ingest "$handoff" core/synthetic/tinystories
```

The four expected file hashes, in filename order, are:

```text
77cf780cebe52b6e83e3a2ac84bc56d8059363113e41d17a023f1d8b2ed0fc0b
e2f173a6dcc2b5895183e8ad376e115bd5c06764eb0cfc01ce1add156b2d017f
cac73cd1a0973a7fccde022a5a5c78f8ba58ea6849541835c58e7c5fc747f2ef
862514ae75723ddd339337f1ef7617b88f82290dcddd57445e1d38d257ca7891
```

WALDO verifies the file count, total bytes, and the manifest's aggregate raw
tree hash during ingestion. Do not continue if ingestion reports a mismatch.

## Ladder

| Gate | Compose | Parameters | Tokens | Purpose |
| --- | --- | ---: | ---: | --- |
| 1 | `0001-tinystories-canary.yaml` | 8.6M | 10M | Ingestion, tokenizer, training, publication, and inference smoke test |
| 2 | `0002-tinystories-8m.yaml` | 8.6M | 500M | Small-model story-learning qualification |
| 3 | `0003-tinystories-32m.yaml` | 32.3M | 1B | Capacity scaling on the identical distribution |

The horizons are controlled budgets, not a claim that 20 tokens per parameter
is optimal for this distribution. Record WALDO's observed corpus pass count and
consumed tokens for every run.

## Run one gate at a time

```console
go run ./cmd/waldo/ model forecast \
  composes/tinystories/0001-tinystories-canary.yaml

go run ./cmd/waldo/ model train tinystories-canary-01 \
  composes/tinystories/0001-tinystories-canary.yaml \
  --hostfile ~/hostfile

./composes/tinystories/evaluate-tinystories.sh \
  tinystories-canary-01 /tmp/tinystories-canary-01-eval.jsonl
```

After Gate 1 passes, repeat with `0002-tinystories-8m.yaml` and a fresh model
name. Do not train Gate 3 until Gate 2 passes.

## Scoring

For each story, score these three dimensions from 0 to 2:

1. **Grammar:** 0 unreadable; 1 understandable with substantial errors; 2
   consistently grammatical simple English.
2. **Prompt consistency:** 0 unrelated or contradictory; 1 partly follows the
   setup; 2 preserves the named characters, objects, and situation.
3. **Plot progression:** 0 loops or makes no progress; 1 adds events without a
   clear resolution; 2 develops and resolves a simple story.

Separately mark repetition failure if a sentence or four-word phrase repeats
more than twice, and record whether generation ended with EOS before 256 tokens.
Do not truncate repeated text before scoring; repetition is part of the result.

## Promotion gates

### Gate 1: pipeline canary

- Training, checkpointing, evaluation, publication, reload, and inference pass.
- Held-out loss is finite and below initialization.
- All ten prompts produce non-empty output.
- No story-quality or EOS threshold applies at 10M tokens.

### Gate 2: 8M qualification

- At least 8/10 stories score 2 for grammar.
- At least 7/10 score 2 for prompt consistency.
- At least 6/10 score at least 1 for plot progression.
- At least 8/10 avoid repetition failure.
- At least 6/10 emit EOS before 256 tokens.
- Held-out loss is finite and improves through the selected checkpoint.

### Gate 3: 32M scaling

- At least 9/10 stories score 2 for grammar.
- At least 8/10 score 2 for prompt consistency.
- At least 8/10 score at least 1 for plot progression.
- At least 9/10 avoid repetition failure.
- At least 8/10 emit EOS before 256 tokens.
- It improves on Gate 2's mean score without worse repetition or EOS behavior.

Once a deterministic gate passes, run a secondary diversity check at
temperature 1 with several seeds. Those samples diagnose breadth; they do not
replace the deterministic promotion result:

```console
./composes/tinystories/evaluate-tinystories.sh \
  MODEL /tmp/MODEL-temp1-seed43.jsonl 1 43
```

## Questions recorded at every gate

1. Did held-out loss and story quality improve together?
2. Did the selected checkpoint and reloaded artifact agree?
3. How many corpus passes did WALDO need for the requested token budget?
4. Did the model learn coherent progression rather than only local grammar?
5. Did it learn EOS without post-training?
6. Were failures caused by looping, contradiction, grammar, or truncation?
7. Does the evidence justify the next gate's cost?

Primary references: [TinyStories paper](https://arxiv.org/abs/2305.07759),
[official dataset](https://huggingface.co/datasets/roneneldan/TinyStories), and
[CDLA-Sharing-1.0](https://cdla.dev/sharing-1-0/).
