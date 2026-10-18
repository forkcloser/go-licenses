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
- `docDiff` hands go-diff copies of the documents' runes: go-diff writes into the slices it is
  given, and the corpus documents are shared by the matches `Libraries` runs concurrently (a data race
  under `-race`).
- Held to go-licenses' lint baseline like the rest of the module, formatting included. The
  package is named `licenseclassifier`, after its directory; `AddContent` returns the tokenizer's
  error instead of dropping it; parameters nothing read (and the searchset's trace-only `origin`)
  are gone; the shadowed and the hand-rolled (`max`, `strings.Index` for `Contains`) give way to
  their standard forms. The long functions and short names it still carries are in the overlay's
  onboarding backlog.
