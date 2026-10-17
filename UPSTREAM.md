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
| On Windows, vendored modules keep their licence URL, and licence URLs use `/` | https://github.com/forkcloser/go-licenses/pull/5 |
| The licence classifier is vendored without its tracing, and go-spew leaves the dependency graph | https://github.com/forkcloser/go-licenses/pull/7 |
| gopkg.in modules resolve without a network lookup; other vanity lookups are https only and retried | https://github.com/forkcloser/go-licenses/pull/8 |
