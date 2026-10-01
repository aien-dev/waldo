# Foundation model ladder

This directory is a stop/go experiment, not a queue of models to run. Train
one rung, record its evidence, and continue only after it passes every gate.
The failed `conversation6` ladder is preserved unchanged in
[`archive/2026-09-conversation6`](archive/2026-09-conversation6/).

## Why we restarted

`conversation6` did not first fail during conversation tuning. Its 12B-token
foundation checkpoint completed cleanly and had plausible held-out loss, but
raw deterministic generation already invented a political biography for
`Linux is`. Later tuning merely changed that failure into repetitive,
content-poor answers. This proves that loss, successful execution, and more
tokens are not sufficient promotion criteria.

The new ladder isolates foundation learning. Gate 0 uses one small, assessed
PressBooks shard so the systems canary stays cheap. The failed 16M capability
experiment is preserved in
[`archive/2026-10-tiny-r50k-failure`](archive/2026-10-tiny-r50k-failure/README.md).
Gates 1-4 use one fixed mixture across 76M and 337M architectures. Each size
gets a 10-token-per-parameter pilot before a run near 20. Parameters remain
float32 and execution remains eager. No conversation or instruction data enters
this ladder.

The qualification mixture contains only assessed schema-2 Wikimedia, PLOS,
and PressBooks shards. DOAB and Gutenberg remain excluded until they are
re-ingested with content assessments; otherwise the declared repetition and
boilerplate filters are silently unavailable for those records.

## Rules

1. Use a fresh model name for every run. Never append a changed recipe to an
   existing model.
2. Run only the next unqualified rung. A failure stops the ladder.
3. Evaluate the selected artifact and intermediate checkpoints, not merely the
   final model name.
4. Use temperature zero and the exact prompts below. Do not tune the questions
   after seeing an answer.
5. Preserve the compose, run summary, run IDs, scores, loss history, and sample
   output for every decision.
6. Change one experimental variable at a time after a failure. Repeat the
   failed rung; do not advance.
7. For every capability rung, tied token embeddings must consume no more than
   50% of total parameters. Report the allocation explicitly.
8. Do not add conversational tuning until `0004-foundation-medium.yaml`
   qualifies.

For seed repeats, copy the qualifying compose into the saved run evidence and
change only `seed`; do not silently edit the numbered reference compose.

## Rungs

| Rung | Suggested model name | Size | Tokens | Tokens/parameter | Question answered |
| --- | --- | ---: | ---: | ---: | --- |
| `0000-foundation-canary.yaml` | `foundation-canary-02` | 16M | 5M | 0.31 | Does the complete pipeline work using one assessed shard? |
| `0001-foundation-small-pilot.yaml` | `foundation-small-pilot-01` | 76M | 760M | 9.95 | Does the first balanced architecture learn recognizable language? |
| `0002-foundation-small.yaml` | `foundation-small-01` | 76M | 1.5B | 19.63 | Does the small model reach honest babbling competence? |
| `0003-foundation-medium-pilot.yaml` | `foundation-medium-pilot-01` | 337M | 3.4B | 10.10 | Does the medium model begin producing locally coherent text? |
| `0004-foundation-medium.yaml` | `foundation-medium-01` | 337M | 6.7B | 19.90 | Is the foundation coherent enough to consider later tuning? |

The small pair allocates 32.2M of 76.4M parameters (42.1%) to embeddings. The
medium pair allocates 57.9M of 336.6M (17.2%). Pilot and qualification within
each pair keep architecture and corpus recipe fixed. The canary corpus is
intentionally smaller and is not capability evidence. Gate 1 has a one-time
materialization cost of roughly 23.6 GiB; later rungs reuse that cache.

## Run procedure

For this ladder, retain the 23.6 GiB assessed corpus selection between runs.
On rank 0, configure a bound with headroom; hostfile launch propagates it to
secondary workers:

```console
waldo config set lookaside.cache.retain-completed true
waldo config set lookaside.cache.max-size 30GiB
waldo lookaside cache status
```

For each rung:

```console
waldo model forecast composes/0000-foundation-canary.yaml
waldo model train foundation-canary-02 composes/0000-foundation-canary.yaml
```

Record every completed run ID. Test a specific artifact rather than whatever
run happens to be selected by the model name:

```console
waldo model chat foundation-canary-02 \
  --run-id RUN_ID --raw --temperature 0 --max-tokens 80 \
  "The Linux kernel is"
```

Track held-out loss at every recorded checkpoint. Run all 15 prompts against
the selected checkpoint artifact and record its run ID. A separately published
artifact must produce materially equivalent loss and generations.

## Fixed evaluation prompts

Factual and definitional continuations:

1. `The Linux kernel is`
2. `Linux is an operating system whose kernel was created by`
3. `The capital city of France is`
4. `At sea level, water freezes at`
5. `Earth orbits`
6. `A central processing unit (CPU) is`
7. `The programming language Python was created by`
8. `Two plus two equals`
9. `Plants use sunlight to`
10. `An operating system manages`

Coherence continuations:

11. `Once upon a time`
12. `The experiment failed because`
13. `To install software on Linux,`
14. `A backup is useful because`
15. `The scientist compared the results and concluded`

Score each response before looking at the aggregate:

- **0:** incoherent, contradictory, unrelated, or fabricated in a way that
  defeats the prompt.
- **1:** grammatical and relevant, but incomplete, vague, or partly wrong.
- **2:** coherent, materially correct, and directly continues the prompt.

A response has a repetition failure if a phrase or sentence loops more than
twice or the answer makes no progress across successive sentences.

## Promotion gates

### Gate 0: canary

- Training, evaluation, checkpointing, resume, publication, and inference all
  complete without non-finite values.
- Materialization resolves one assessed shard (about 181 MB), not the full
  multi-corpus dataset.
- Token accounting is exact and held-out loss moves downward.
- The selected checkpoint and published artifact agree within the existing
  artifact-integrity tolerance.
- All 15 prompts produce non-empty output. No knowledge score is required.

Observed result: `foundation-canary-02` run `100f0082a33af66f` completed all
306 steps on two H200 GPUs. Held-out loss fell from 10.8531 to 6.2503, and the
selected, master, and reloaded-artifact losses agreed at 6.2503. Deterministic
generation was non-empty but collapsed into repeated phrases and bullets. At
only 0.31 tokens per parameter, this is a systems pass and a capability fail by
design; it must not be used to judge the corpus recipe.

### Gate 1: small pilot

- Gate 0 still passes.
- At least **8/15** responses avoid severe repetition failure.
- At least **10/15** responses contain recognizable English word sequences
  related to the prompt.
- No factual accuracy threshold applies. This rung is expected to babble.
- Held-out loss is still improving at the selected checkpoint.

### Gate 2: small qualification

- Fixed-prompt score is at least **10/30**.
- At least **10/15** responses avoid repetition failure.
- At least **10/15** begin with a grammatical sentence or sentence fragment.
- The selected checkpoint improves materially over the small pilot without
  crossing a visible loss or generation peak.

### Gate 3: medium pilot

- Fixed-prompt score is at least **15/30**.
- At least **12/15** responses avoid repetition failure.
- At least 7 of the 10 factual prompts are relevant, even if not fully correct.
- Loss and prompt quality both improve relative to the small qualification.

### Gate 4: medium qualification

- Fixed-prompt score is at least **18/30**.
- At least **13/15** responses avoid repetition failure.
- At least 8 of the 10 factual prompts score at least 1.
- Repeat once with seed 43 only after the seed-42 run passes. Both runs must
  pass artifact integrity and avoid broad repetition collapse.
- Only after this gate passes may a separate conversational-tuning ladder be
  designed. These prompts remain permanent regression tests.

## Questions to answer at every rung

Record short, evidence-backed answers with the run results:

1. Did held-out loss and fixed-prompt capability improve together?
2. At which checkpoint did prompt score peak?
3. Did any corpus dominate sampled tokens beyond its declared weight?
4. Did any host, rank, or resume event change token accounting or loss?
5. Are selected-checkpoint and published-artifact results equivalent?
6. Which failures are factual, incoherent, repetitive, or off-topic?
7. Does the result justify the cost of the next rung?

If a rung fails, investigate the smallest relevant variable: corpus samples
and weights, tokenizer, initialization, learning rate and schedule, effective
global batch, or architecture. Do not compensate for a broken foundation with
instruction data.

`holding/tool-use.yaml` remains intentionally outside this ladder.
