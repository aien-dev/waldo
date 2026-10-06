# Reference-model training ladder

This directory contains the active, gated path from a reproducible narrow
language model toward broader capability. Every rung has one hypothesis, a
pinned compose, fixed prompts, and explicit promotion criteria. A failed rung
stops the ladder; it does not justify changing several variables at once.

Historical composes remain under [`archive`](archive). The `experiments`,
`general-foundation`, and `tinystories` subdirectories preserve earlier
diagnostics and are not active ladder rungs.

## Rung 0001: Tiny Shakespeare reference — passed

[`0001-tiny-shakespeare.yaml`](0001-tiny-shakespeare.yaml) is WALDO's first
reference model. It uses the exact 1,115,394-byte Tiny Shakespeare text, the
built-in byte tokenizer, a 10.7M-parameter decoder, and a deterministic 90/10
contiguous split that does not introduce artificial line-level EOS tokens.

Validated result on 2026-10-05:

- model `tiny-shakespeare-control-01`, ID `b4f8477a55ad`;
- 5,000 optimizer steps and 81.92M consumed tokens;
- held-out loss improved from 5.3814 to 1.5117;
- step 1,250 was correctly selected and reloaded after terminal loss rose to
  2.0940;
- all eight temperature-0.8 samples preserved play formatting and produced
  locally plausible Shakespeare-like text without immediate loop collapse;
- greedy decoding exposed a repeat attractor around "season/state/seas"; and
- training completed in under 15 minutes on two H200 GPUs.

The full 81.92M-token compose is retained because it reproduces the learning
curve, overtraining evidence, and selected checkpoint. Changing its horizon to
20.48M would also change the cosine schedule and would not reproduce the same
checkpoint.

Run and evaluate it with:

```console
go run ./cmd/waldo/ model forecast composes/0001-tiny-shakespeare.yaml
go run ./cmd/waldo/ model train tiny-shakespeare-control-01 \
  composes/0001-tiny-shakespeare.yaml
./composes/evaluate-tiny-shakespeare.sh \
  tiny-shakespeare-control-01 /tmp/tiny-shakespeare-control-01-eval.jsonl
```

This rung proves that WALDO's ingestion, byte tokenization, packing, optimizer,
held-out evaluation, checkpoint selection, reload, and raw generation paths can
learn a real language distribution. EOS is not a gate because this corpus is
one continuous document.

## Rung 0002: TinyStories byte control — diagnostic complete

[`0002-tinystories-byte.yaml`](0002-tinystories-byte.yaml) asks whether the
same proven 10.7M model can move from one play-like stream to many short,
simple stories. It keeps the architecture, tokenizer, context, batch,
optimizer, learning rate, dropout, initialization, and seed from rung 0001.
The intentional changes are the corpus, record-level evaluation, shuffle
capacity, and the longer 15,000-step horizon needed to encounter diverse
stories.

The corpus is the first pinned training Parquet shard from the original
TinyStories release. This bounded quarter-corpus makes the rung practical on a
Mac while retaining hundreds of thousands of complete story records. It is a
WALDO-shaped learning control, not yet an exact reproduction of the paper's
alternating GPT-Neo attention or pruned tokenizer. The paper and published
prompts are available from the
[TinyStories project](https://huggingface.co/datasets/roneneldan/TinyStories)
and [paper](https://arxiv.org/abs/2305.07759).

Hypothesis: changing only to a constrained simple-English story distribution
will preserve grammatical generation while improving entity, causal, and
short-plot consistency beyond the Shakespeare style control.

Promotion gates:

1. Training completes without non-finite loss, and the selected checkpoint's
   reloaded held-out loss agrees with its persisted evaluation.
2. Best held-out loss improves by at least 50% from the initial evaluation.
3. At least six of eight temperature-0.8 samples remain grammatical and retain
   the prompt's people, objects, and causal setup for at least 150 generated
   byte tokens.
4. No more than two samples collapse into an immediate repeated sentence or
   phrase loop.
5. At least three samples emit EOS within 256 generated tokens. Unlike rung
   0001, every TinyStories record teaches a real document boundary.
6. Greedy output is recorded as a degeneration diagnostic, but it is not the
   sole promotion decision.

Validated results on 2026-10-05:

- the four-GPU, two-host run reached held-out loss 0.5782 after 245.8M tokens;
- the two-GPU, single-host control reached 0.5794 with the same architecture,
  data, token budget, global batch, optimizer steps, and seed;
- the 0.21% final-loss difference passes the 3% topology-equivalence gate and
  rules out multi-host data parallelism as the cause of the generation issues;
- both runs produced grammatical simple-story prose at temperature 0.8, but
  lost prompt entities and causal details, exhibited greedy repetition, and
  rarely emitted EOS within the evaluation horizon; and
- rung 0002 therefore passes learning and execution controls but fails its
  prompt-retention, repetition, and EOS capability gates.

The failure is useful evidence rather than a reason to tune several settings:
with byte tokenization, the 256-token context is exactly 256 UTF-8 bytes. The
published prompts consume much of that window before generation begins.

Run it only after `core/synthetic/tinystories-reference` has been ingested:

```console
cd ../fetchers
go run ./cmd/fetcher corpora/tinystories-reference.ini \
  /tmp/tinystories-reference
cd ../waldo
go run ./cmd/waldo/ index ingest /tmp/tinystories-reference \
  core/synthetic/tinystories-reference

go run ./cmd/waldo/ model forecast composes/0002-tinystories-byte.yaml
go run ./cmd/waldo/ model train tinystories-byte-01 \
  composes/0002-tinystories-byte.yaml
./composes/evaluate-tinystories.sh \
  tinystories-byte-01 /tmp/tinystories-byte-01-eval.jsonl
./composes/evaluate-tinystories.sh \
  tinystories-byte-01 /tmp/tinystories-byte-01-greedy.jsonl 0 42
```

## Rung 0003: TinyStories 512-byte context — ready

[`0003-tinystories-context-512.yaml`](0003-tinystories-context-512.yaml)
isolates the next hypothesis: rung 0002 failed because a 256-byte window cannot
hold the prompt and enough continuation to preserve its setup. It changes only
the architecture and training sequence lengths from 256 to 512 and reduces the
global batch from 64 to 32. Both rungs therefore retain 16,384 tokens per
optimizer update, 15,000 optimizer steps, and 245.76M total training tokens.

Promotion gates remain the six rung-0002 gates. The comparison must answer:

1. Does held-out loss remain stable or improve without changing the token
   budget or optimizer-step count?
2. Do at least six of eight temperature-0.8 samples retain the people,
   objects, and causal setup for 150 generated bytes?
3. Are immediate repetition and EOS materially better than rung 0002?

Run and evaluate it with:

```console
go run ./cmd/waldo/ model forecast composes/0003-tinystories-context-512.yaml
go run ./cmd/waldo/ model train tinystories-context-512-01 \
  composes/0003-tinystories-context-512.yaml \
  --hostfile ~/hostfile
./composes/evaluate-tinystories.sh \
  tinystories-context-512-01 /tmp/tinystories-context-512-01-eval.jsonl
./composes/evaluate-tinystories.sh \
  tinystories-context-512-01 /tmp/tinystories-context-512-01-greedy.jsonl 0 42
```

## Later rungs

Do not create rung 0004 until rung 0003 has a written gate decision. The likely
next isolated variable is a compact byte-BPE tokenizer, followed by model
scaling. General-corpus mixtures come only after these narrow reference
controls establish stable grammar, consistency, EOS, repetition, and held-out
behavior.

Every result must retain the compose, model summary, run ID, telemetry,
consumption report, temperature samples, greedy samples, and a written gate
decision.

## Forecast before training

`waldo model forecast <compose>` now reports the exact WALDO parameter
decomposition, core/token-I/O allocation, GQA projections, SwiGLU matrices,
context fitness, optimizer-step arithmetic, conventional and
architecture-aware compute, memory components, advisory warnings, and JSON
fields under `forecast.fitness`.

For a byte tokenizer, token context is also exact byte context. The current
reference models therefore expose 256 tokens as exactly 256 UTF-8 bytes; after
150 generated bytes, no more than about 106 prompt bytes can remain visible.
For BPE and multi-corpus recipes, bytes/token, record percentiles, fit rate,
packed boundaries, EOS targets, and per-corpus fertility are marked as
requiring corpus preflight rather than estimated from unrelated manifest token
counts.

Warnings do not promote, reject, or launch a run. In particular,
tokens-per-parameter and the Chinchilla allocation are not capability gates.
Loss prediction is refused until WALDO has a sufficiently large cohort with
the same corpus/held-out revision, tokenizer, architecture family, context,
objective, and optimizer recipe. See
[`docs/MODEL-TRAINING-FITNESS.md`](../docs/MODEL-TRAINING-FITNESS.md).
