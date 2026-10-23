# Upstream

Fork of https://github.com/google/go-licenses (`master`).

- Forked from: [`3e084b0`](https://github.com/google/go-licenses/commit/3e084b0caf710f7bfead967567539214f598c0a2), committed 2025-09-08
- Reviewed through: [`3e084b0`](https://github.com/google/go-licenses/commit/3e084b0caf710f7bfead967567539214f598c0a2), reviewed 2026-10-15

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
