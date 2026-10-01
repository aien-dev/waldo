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
PressBooks shard so the systems canary stays cheap. Gates 1-4 use one fixed,
prose-heavy mixture to qualify the 16M model before attempting the 76M model.
Each size gets a short pilot before a run near 20 training tokens per parameter.
Parameters remain float32 and execution remains eager. No conversation or
instruction data enters this ladder.

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
7. Do not create a medium-model ladder until `0004-foundation-small.yaml`
   passes. Do not add conversational tuning until a later medium foundation
   qualifies.

For seed repeats, copy the qualifying compose into the saved run evidence and
change only `seed`; do not silently edit the numbered reference compose.

## Rungs

| Rung | Suggested model name | Size | Tokens | Tokens/parameter | Question answered |
| --- | --- | ---: | ---: | ---: | --- |
| `0000-foundation-canary.yaml` | `foundation-canary-02` | 16M | 5M | 0.31 | Does the complete pipeline work using one assessed shard? |
| `0001-foundation-tiny-pilot.yaml` | `foundation-tiny-pilot-02` | 16M | 160M | 9.99 | Does the full recipe produce coherent deterministic generation? |
| `0002-foundation-tiny.yaml` | `foundation-tiny-01` | 16M | 320M | 19.98 | Can a properly exposed tiny model learn coherent language? |
| `0003-foundation-small-pilot.yaml` | `foundation-small-pilot-01` | 76M | 760M | 9.95 | Does the 76M architecture follow the proven learning curve? |
| `0004-foundation-small.yaml` | `foundation-small-01` | 76M | 1.5B | 19.63 | Does the small foundation justify designing a medium ladder? |

The tiny pilot and qualification use the same architecture and recipe. The
small pair does likewise. Only the token budget, batch, and observation cadence
change between sizes. The canary corpus is intentionally smaller and is not
evidence for the quality of the three-corpus recipe. The first Gate 1 run has a
one-time materialization cost of roughly 23.6 GiB; later rungs reuse that cache.

## Run procedure

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

Run all 15 prompts against the initial checkpoint, intermediate checkpoints,
the selected checkpoint, and the published artifact when available. A
published artifact must produce materially equivalent loss and generations to
its selected checkpoint.

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

### Gate 1: tiny pilot

- Gate 0 still passes.
- Fixed-prompt score is at least **12/30**.
- At least **11/15** responses avoid repetition failure.
- The later checkpoints improve both held-out loss and prompt score over the
  early checkpoint. If loss improves while prompt score degrades, stop.

The original 50M-token `foundation-tiny-pilot-01` run
`246cdae7ed71a425` failed this gate. Four-GPU accounting and 5:2:1 corpus
exposure were correct, and held-out loss improved from 10.8795 to a final-best
4.9693, but deterministic answers were off-topic and severely repetitive.
At 3.12 tokens per parameter, this was an underexposed learning-curve point,
not evidence for scaling. The replacement pilot requests 160M tokens and must
use the fresh `foundation-tiny-pilot-02` model name.

### Gate 2: tiny qualification

- Fixed-prompt score is at least **18/30**.
- At least 7 of the 10 factual prompts score at least 1.
- At least **13/15** responses avoid repetition failure.
- The selected checkpoint beats the tiny pilot; continued training has not
  crossed a visible capability peak.
- Repeat the qualifying recipe with seeds 43 and 44. All three runs must score
  at least 16/30 and avoid broad repetition collapse before scaling the model.

### Gate 3: small pilot

- Fixed-prompt score is at least **16/30**.
- At least **13/15** responses avoid repetition failure.
- Loss and prompt score improve together across checkpoints.
- The pilot must match or beat the qualified tiny model before receiving the
  full small-model budget.

### Gate 4: small qualification

- Fixed-prompt score is at least **22/30**.
- Every factual prompt scores at least 1.
- No response has a repetition failure.
- Repeat the qualifying recipe with seeds 43 and 44. All three runs must score
  at least 20/30 and pass artifact integrity.
- Only after this gate passes may a separate medium foundation ladder be
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
