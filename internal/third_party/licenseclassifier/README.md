# licenseclassifier

Vendored from <https://github.com/google/licenseclassifier/tree/v2.0.0/v2> (`e6a9bb9`, the last
commit under `v2/`), the classifier go-licenses identifies licences with. Only what go-licenses runs is
kept: the engine and the `assets` package with the licence texts it embeds.

Local modifications:

- Import paths.
- Tracing removed (`trace.go`, `SetTraceConfiguration` and every trace call), and with it the
  dependency on github.com/davecgh/go-spew, which only the trace output used.
- Removed what go-licenses does not call: `Classifier.Match`, `Normalize` and `LoadLicenses`,
  two unexported helpers only upstream's tests used, and the `assets` package's `ReadLicenseFile`
  and `ReadLicenseDir`.
- The licence texts lose trailing whitespace, carriage returns and trailing blank lines (145 of
  431). The tokenizer drops whitespace, so each text yields the same document as upstream's.
