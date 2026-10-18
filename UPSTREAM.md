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
