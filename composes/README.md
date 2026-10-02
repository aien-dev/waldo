# Foundation model ladder

This is a stop/go experiment. Run one gate, preserve its evidence, and advance
only when it passes. The previous r50k ladder is preserved in
[`archive/2026-10-r50k-ladder-retired`](archive/2026-10-r50k-ladder-retired/README.md).

## What changed

The old 76.4M model trained correctly and reached held-out loss 3.2127, but its
generations still collapsed. Its 50,259-token embedding table consumed 32.2M
parameters (42.1%), leaving only about 44.2M for transformer blocks. More loss
reduction on that design was not evidence that it would become useful.

The new ladder separates four questions:

1. Is the tokenizer compact, deterministic, and portable?
2. Do training, publication, inference, and weighted sampling remain correct?
3. Can a ~76M diagnostic model acquire non-repetitive language?
4. Can a ~297M model meet basic general-foundation capability gates?

There is no conversation tuning in this ladder. A small-model failure stops the
experiment; it is not repaired with instruction data.

## Gate 0: train and review the tokenizer

Pull the commit containing this ladder, then run on rank 0:

```console
mkdir -p composes/tokenizers
go run ./cmd/waldo/ model train-tokenizer \
  core/common-pile/wikimedia \
  core/common-pile/pressbooks \
  science/plos \
  --vocabulary-size 16000 \
  --sample-bytes 268435456 \
  --output composes/tokenizers/foundation-16k.json
```

The command samples the three corpus paths evenly, regardless of their later
training weights. Stop if it does not produce exactly 16,000 tokens, if the
artifact fails validation, or if its reported bytes/token is more than 10%
worse than r50k on the same sample. Preserve the complete output. Once accepted,
commit `composes/tokenizers/foundation-16k.json`; it is part of the model.

The compose files use `.yaml.tmpl` only because the content-addressed tokenizer
does not exist in Git yet. They are directly runnable after Gate 0; no rendering
or substitution is required.

## Rungs

| Gate | Compose | Approx. size | Tokens | Purpose |
| --- | --- | ---: | ---: | --- |
| 1 | `0001-foundation-pipeline-canary.yaml.tmpl` | 7.3M | 10M | Trained-tokenizer, training, checkpoint, publication, and inference smoke test |
| 2 | `0002-foundation-mixture-canary.yaml.tmpl` | 7.3M | 50M | Exact 5:2:1 weighted-stream accounting |
| 3 | `0003-foundation-small-language.yaml.tmpl` | 76.6M | 760M | Language diagnostic using only Wikimedia and PressBooks |
| 4 | `0004-foundation-small-general.yaml.tmpl` | 76.6M | 1.5B | General-mixture diagnostic; not a useful-assistant claim |
| 5 | `0005-foundation-medium-pilot.yaml.tmpl` | 297.2M | 3.0B | First general-capability pilot |
| 6 | `0006-foundation-medium.yaml.tmpl` | 297.2M | 6.0B | General-foundation qualification |

The 76.6M architecture allocates 10.2M parameters (13.4%) to tied token
embeddings. The 297.2M architecture allocates 18.4M (6.2%). Gate 3 deliberately
removes PLOS and uses a 1:1 Wikimedia/PressBooks stream to test ordinary prose
before adding the scientific domain. Gates 2, 4, 5, and 6 use 62.5% Wikimedia,
25% PressBooks, and 12.5% PLOS.

## Rules

1. Use a fresh model name for each run and a specific run ID for evaluation.
2. Never run a later gate while an earlier gate is unqualified.
3. Preserve compose, summary, run ID, loss history, consumption, and samples.
4. Evaluate temperature 0 first. Sampling cannot rescue deterministic collapse.
5. A falling held-out loss is necessary but never sufficient for promotion.
6. Change one variable after failure and repeat that gate.
7. Do not add conversation data until Gate 6 passes twice with different seeds.

## Run procedure

After Gate 0, start only Gate 1:

```console
go run ./cmd/waldo/ model forecast \
  composes/0001-foundation-pipeline-canary.yaml.tmpl

go run ./cmd/waldo/ model train foundation-pipeline-canary-01 \
  composes/0001-foundation-pipeline-canary.yaml.tmpl \
  --hostfile ~/hostfile
```

Extract the selected run ID and always evaluate that immutable artifact:

```console
run_id="$(go run ./cmd/waldo/ --json model summary MODEL_NAME | jq -r '.bom.current_run_id')"
go run ./cmd/waldo/ model chat MODEL_NAME \
  --run-id "$run_id" --raw --temperature 0 --max-tokens 80 \
  "The Linux kernel is"
```

## Fixed prompts

Use the same prompts at every capability gate:

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
11. `Once upon a time`
12. `The experiment failed because`
13. `To install software on Linux,`
14. `A backup is useful because`
15. `The scientist compared the results and concluded`

Score each response: **0** means incoherent, unrelated, or materially wrong;
**1** means relevant but incomplete, vague, or partly wrong; **2** means coherent
and materially correct. Separately mark a repetition failure when a phrase or
sentence loops more than twice or successive sentences make no progress.

For Gates 3-6, also sample prompts 1, 3, 11, and 14 at temperature 0.7. These
samples diagnose distributional collapse but do not replace deterministic
scores.

## Promotion gates

### Gate 1: pipeline canary

- Training, checkpointing, evaluation, publication, and inference complete.
- Token accounting is exact; loss is finite and lower than initialization.
- Reloaded artifact loss matches the selected checkpoint.
- All 15 prompts return non-empty output. No knowledge score is required.

### Gate 2: mixture canary

- Gate 1 still passes.
- Exposure is within 0.1 percentage points of 62.5% Wikimedia, 25% PressBooks,
  and 12.5% PLOS; every corpus contributes.
- No capability threshold applies to this 7.3M model.

### Gate 3: small language diagnostic

- At least 10/15 deterministic outputs begin with a grammatical English
  sentence or fragment.
- At least 10/15 remain related to the prompt for the first sentence.
- At least 8/15 avoid repetition failure.
- At least 3/4 sampled probes remain relevant and avoid repetition.
- No factual-score threshold applies.

### Gate 4: small general diagnostic

- Fixed-prompt score is at least 10/30.
- At least 10/15 outputs avoid repetition and 6/10 factual prompts are relevant.
- It improves over Gate 3 on the fixed suite without a loss or generation peak.
- Passing authorizes the medium pilot; it does not claim the model is useful.

### Gate 5: medium pilot

- Fixed-prompt score is at least 15/30.
- At least 12/15 outputs avoid repetition.
- At least 7/10 factual prompts are relevant and at least 5 score 1 or better.
- Loss and prompt quality both improve over Gate 4.

### Gate 6: medium qualification

- Fixed-prompt score is at least 18/30.
- At least 13/15 outputs avoid repetition.
- At least 8/10 factual prompts score 1 or better.
- Repeat with seed 43 only after seed 42 passes; both runs must qualify.
- Only then design a separate conversation-tuning ladder.

## Questions recorded at every gate

1. Did held-out loss and fixed-prompt quality improve together?
2. Which checkpoint had the best prompt score?
3. Did observed corpus exposure match the declared weights?
4. Did host count, rank, or resume change accounting or loss?
5. Did the selected checkpoint and published artifact agree?
6. Were failures factual, incoherent, repetitive, or off-topic?
7. Does the evidence justify the next gate's cost?
