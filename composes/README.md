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
prose-heavy mixture across two model sizes, with short pilots before full runs,
float32 portable parameters, eager execution, and deterministic checkpoint
evaluation. No conversation or instruction data enters this ladder.

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
7. Do not add conversational tuning until `0004-foundation-medium.yaml` passes.

For seed repeats, copy the qualifying compose into the saved run evidence and
change only `seed`; do not silently edit the numbered reference compose.

## Rungs

| Rung | Suggested model name | Approximate size | Token request | Question answered |
| --- | --- | ---: | ---: | --- |
| `0000-foundation-canary.yaml` | `foundation-canary-02` | 16M | 5M | Does the complete pipeline work using one assessed shard? |
| `0001-foundation-small-pilot.yaml` | `foundation-small-pilot-01` | 76M | 50M | Does this recipe begin learning coherent language cheaply? |
| `0002-foundation-small.yaml` | `foundation-small-01` | 76M | 600M | Can the small model meet a real capability floor? |
| `0003-foundation-medium-pilot.yaml` | `foundation-medium-pilot-01` | 337M | 300M | Does scaling the architecture improve the fixed evaluation? |
| `0004-foundation-medium.yaml` | `foundation-medium-01` | 337M | 2.4B | Is the foundation good enough to justify conversational tuning? |

The small pilot and qualification use the same architecture and recipe. The
medium pair does likewise. Only the token budget and observation cadence
change within each pair. The canary corpus is intentionally smaller and is not
evidence for the quality of the four-corpus recipe.

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

### Gate 1: small pilot

- Gate 0 still passes.
- Fixed-prompt score is at least **15/30**.
- At least **12/15** responses avoid repetition failure.
- The later checkpoints improve both held-out loss and prompt score over the
  early checkpoint. If loss improves while prompt score degrades, stop.

### Gate 2: small qualification

- Fixed-prompt score is at least **22/30**.
- Every factual prompt scores at least 1.
- No response has a repetition failure.
- The selected checkpoint beats the small pilot; continued training has not
  crossed a visible capability peak.
- Repeat the qualifying recipe with seeds 43 and 44. All three runs must score
  at least 20/30 and avoid repetition collapse before scaling the model.

### Gate 3: medium pilot

- Fixed-prompt score is at least **22/30**, with no repetition failure.
- It matches the small qualification after half as many training tokens, or
  shows a clear checkpoint trend likely to exceed it at the full budget.
- No more than two individual prompts regress relative to the qualified small
  model.

### Gate 4: medium qualification

- Fixed-prompt score is at least **27/30**.
- At least 9 of the 10 factual prompts score 2; every factual prompt scores at
  least 1.
- No response has a repetition failure.
- Repeat the qualifying recipe with seeds 43 and 44. All three runs must score
  at least 25/30 and pass artifact integrity.
- Only after this gate passes may a separate conversational-tuning ladder be
  designed. Foundation prompts remain permanent regression tests.

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
