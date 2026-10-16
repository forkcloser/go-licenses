# go/pkgsite/source

Vendored from <https://go.googlesource.com/pkgsite/+/beceacdece62d95d6dc41a9b5f09da7b2a021020/internal/source>,
because the source package is internal, and there's no plan to move it out anytime soon: <https://github.com/golang/go/issues/40477#issuecomment-868532845>.

The entire source folder and be downloaded via:

```bash
curl -LO https://go.googlesource.com/pkgsite/+archive/beceacdece62d95d6dc41a9b5f09da7b2a021020/internal.tar.gz
```

Local modifications:

- Update import paths.
- Removed unused functions from pkgsite/internal/stdlib, pkgsite/internal/derrors,
  pkgsite/internal/version to avoid other dependencies.
- For pkgsite/internal/source, switched to use go log package, because glog conflicts with a test
  dependency that also defines the "v" flag.
- Add a SetCommit method to type ModuleInfo in ./source/source_patch.go, more rationale explained in the method's comments.
- `NewClient` builds on primordium's retrying transport (github.com/mycophonic/primordium/network/transporter)
  instead of an opencensus one. The go-get meta lookup is https only, reading the tags whatever the
  status, as the go command does; it no longer falls back to plain http.
- gopkg.in modules resolve without a lookup (`gopkgInInfo`, in ./source/source_patch.go): its mapping to
  GitHub is fixed, and its proxy answered CI with 502s and timeouts.
