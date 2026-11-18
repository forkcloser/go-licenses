# Upstream

Fork of https://github.com/google/go-licenses (`master`).

- Forked from: [`3e084b0`](https://github.com/google/go-licenses/commit/3e084b0caf710f7bfead967567539214f598c0a2), committed 2025-09-08
- Reviewed through: [`3e084b0`](https://github.com/google/go-licenses/commit/3e084b0caf710f7bfead967567539214f598c0a2), reviewed 2026-11-17
- The licence classifier is vendored from https://github.com/google/licenseclassifier, its `v2/` at [`e6a9bb9`](https://github.com/google/licenseclassifier/commit/e6a9bb99b5a6f71d5a34336b8245e305f5430f99), the last commit under `v2/`; reviewed 2026-11-16: seven commits on its `main` since, none under `v2/`. What is changed in it is listed in [its README](./internal/third_party/licenseclassifier/README.md)

## Incorporated

| Upstream commit | Here | What |
|---|---|---|

## Not incorporated

| Upstream commit | Why |
|---|---|

## Changes upstream does not have

| Change | Here |
| --- | --- |
| `golang.org/x/text` past GO-2026-5970, which `check` reaches | https://github.com/forkcloser/go-licenses/pull/1 |
| On Windows, vendored modules are recognised as vendored | https://github.com/forkcloser/go-licenses/pull/5 |
| The licence classifier is vendored without its tracing, and go-spew leaves the dependency graph | https://github.com/forkcloser/go-licenses/pull/7 |
| Installed from `github.com/forkcloser/go-licenses/cmd/go-licenses`, and released as signed binaries | https://github.com/forkcloser/go-licenses/pull/15 |
| No `report` or `csv` command, and the vendored pkgsite with them: go-licenses itself sends no HTTP request | https://github.com/forkcloser/go-licenses/pull/21 |
| Flags parsed by kong, upstream's names kept; logs through `log/slog` (`LOG_LEVEL`; colored on a terminal, JSON otherwise) instead of klog, whose flags (`-v`, `--logtostderr`, ...) are gone | https://github.com/forkcloser/go-licenses/pull/23 |
| Licence files classified concurrently without a data race in the classifier | https://github.com/forkcloser/go-licenses/pull/24 |
| A library of sibling packages sharing a name prefix is named by their common directory (`m/pkg` beside `m/pkgextra` is `m`, where it was `m/pkg`); the licence search is bounded by a path relation, not a string prefix, so a sibling module version's licence is never handed over; `check --disallowed_types` with an unknown type is an error naming it and the types there are, where it went on with the defaults; `save --force` judges every library before removing the previous save, where a failing run had destroyed it | https://github.com/forkcloser/go-licenses/pull/31 |
| The classifier names two licence texts that share a line, whichever is closer to its licence, where it dropped one; and no longer names a licence for a text that lies inside another's last line: the ISC text with a paragraph a line was ISC and BSD-0-Clause, and `check --allowed_licenses=ISC` refused it | https://github.com/forkcloser/go-licenses/pull/89 |

## Upstream defects known and kept

| Defect | Here |
| --- | --- |
| licenseclassifier: a text stops being named before its confidence falls to the threshold. The classifier first wants the text to hold 0.8 of the licence's distinct words, and only then measures the confidence, which must reach 0.8 too; the first test is the stricter. The MIT text short of its last 28 words is named nothing, at a confidence of 0.827. `frequencies.go` says of the first test that a value under the threshold "can't possibly match", which is not so | Kept: it errs toward naming nothing, and `check` then fails for a licence not found. Tests hold where it falls, https://github.com/forkcloser/go-licenses/pull/87 |
| licenseclassifier: two licence texts on one line whose seam is two words they share, written once, are not both named | Kept, as upstream has it: none of the 55 such pairs of the corpus is named whole before or after https://github.com/forkcloser/go-licenses/pull/89 |
