# Model training experiments

The active plan is the [general-foundation ladder](general-foundation/README.md).
Run its 32.3M data ablation before spending compute on the 125.6M model.

## Current conclusion

WALDO's training mechanics are healthy: multi-host data parallelism, weighted
streaming, trained byte-BPE tokenizers, checkpoint selection, FP32 publication,
reload verification, and inference have all completed successfully. The
remaining problem is model capability per token, driven jointly by model size
and corpus construction.

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

Accordingly, the next gate changes the broad data mixture while holding the
32.3M architecture and 1B-token budget fixed. Only a successful data ablation
authorizes the 125.6M ladder.

## Experiment directories

- [`general-foundation`](general-foundation/README.md): active 32.3M data
  ablation followed by 125.6M pilot and qualification.
- [`tinystories`](tinystories/README.md): completed Cosmopedia proxy and its
  failed coherence hypothesis.
- [`experiments`](experiments/README.md): completed assistant-EOS post-training
  diagnostics.
- [`archive`](archive): retired ladders and preserved failure evidence.

The numbered YAML files still at this directory's root are the previous
byte-BPE foundation ladder. They remain reproducible historical inputs but are
not the active plan. Do not continue directly to
`0005-foundation-medium-pilot.yaml` or `0006-foundation-medium.yaml`.

## Rules shared by every ladder

1. Run only one gate at a time with a fresh model name.
2. Preserve compose, summary, run ID, telemetry, consumption, and evaluations.
3. Evaluate temperature 0 before sampling.
4. Falling held-out loss is necessary but never sufficient.
5. Change one causal variable per diagnostic rung.
6. Do not use post-training to conceal foundation incoherence or repetition.
7. Do not silently change, relabel, or overwrite an indexed corpus or compose.

## Corpus and post-training policy

Corpus extraction may normalize structure, remove versioned extraction
artifacts, annotate quality, deduplicate, and publish immutable derived views.
It must not silently paraphrase or fact-edit canonical source records. Generated
text is a separate synthetic corpus with complete provenance.

Post-training data should teach interaction, response shape, and EOS after the
foundation passes. It should combine high-quality factual and procedural
responses with a small format-contract component; the prior experiments show
that format examples alone create a stopped but canned model.

The detailed implementation plan, mixture, gates, and research basis are in
the active [general-foundation README](general-foundation/README.md).
