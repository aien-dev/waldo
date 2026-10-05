# Model training experiments

All current training ladders are frozen. Do not start a larger foundation run
from this directory. The repository-wide audit found that the project must
first prove numerical conformance, reproduce independent learning controls,
and establish quantitative evaluation. The authoritative sequence is the
[training validation and capability plan](../docs/TRAINING-ROBUSTNESS-PLAN.md).

## Current conclusion

WALDO's systems paths have completed successfully: multi-host data parallelism,
weighted streaming, trained byte-BPE tokenizers, checkpoint selection, FP32
publication, reload verification, and inference. These runs do not independently
prove the model math or optimizer update, and the remaining problem cannot yet
be attributed only to model size or corpus construction.

The experiments completed so far show:

- The old r50k models devoted too much small-model capacity to token I/O.
- The 76.6M byte-BPE prose model improved loss through 1.5B tokens but retained
  deterministic repetition, so a 20-tokens-per-parameter ratio was not a
  capability guarantee.
- Assistant-only post-training increased EOS from 0/10 to 7–8/10, proving that
  stopping can be taught. Contract-heavy SFT introduced canned language, while
  broader SFT still could not repair foundation correctness or repetition.
- The Cosmopedia proxy improved from loss 4.9915 at 8.6M/10M tokens, to 2.6408
  at 8.6M/500M, to 2.1102 at 32.3M/1B. Nevertheless, the final model produced
  0/10 EOS in both greedy and temperature-0.7 suites, repeated, and frequently
  abandoned prompts. Lower loss and greater capacity learned the corpus style
  without producing stable semantics.
- The 32.3M/1B general mixture accurately consumed its intended
  55/25/15/5 shares and reached loss 2.9481, but only 2/15 deterministic probes
  emitted EOS; the rest hit the token limit with severe factual errors,
  repetition, and visible Stack Exchange and PLOS artifacts.
- Twenty tokens per parameter is a compute-allocation heuristic, not a
  capability guarantee. Public small general models are commonly trained far
  beyond that point.

Accordingly, another mixture ablation or larger rung is not authorized. The
next work is validation code and reference controls, not a new compose.

## Experiment directories

- [`general-foundation`](general-foundation/README.md): frozen after the failed
  32.3M data ablation; do not run its 125.6M composes.
- [`tinystories`](tinystories/README.md): completed Cosmopedia proxy and its
  failed coherence hypothesis.
- [`experiments`](experiments/README.md): completed assistant-EOS post-training
  diagnostics.
- [`archive`](archive): retired ladders and preserved failure evidence.

The numbered YAML files still at this directory's root are the previous
byte-BPE foundation ladder. They remain reproducible historical inputs but are
not the active plan. Do not continue directly to
`0005-foundation-medium-pilot.yaml` or `0006-foundation-medium.yaml`.

## Rules shared by every future experiment

1. Register one falsifiable hypothesis and one changed variable before a run.
2. Preserve compose, summary, run ID, telemetry, consumption, and evaluations.
3. Compare against an independent or public baseline with the same metric.
4. Use quantitative gates and declared uncertainty; samples are diagnostic.
5. Falling held-out loss is necessary but never sufficient.
6. Failed gates stop dependent runs.
7. Do not use post-training to conceal foundation incoherence or repetition.
8. Do not silently change, relabel, or overwrite an indexed corpus or compose.

## Corpus and post-training policy

Corpus extraction may normalize structure, remove versioned extraction
artifacts, annotate quality, deduplicate, and publish immutable derived views.
It must not silently paraphrase or fact-edit canonical source records. Generated
text is a separate synthetic corpus with complete provenance.

Post-training data should teach interaction, response shape, and EOS after the
foundation passes. It should combine high-quality factual and procedural
responses with a small format-contract component; the prior experiments show
that format examples alone create a stopped but canned model.

The detailed diagnosis, implementation sequence, gates, and research basis are
in the [training validation and capability plan](../docs/TRAINING-ROBUSTNESS-PLAN.md).
