# Release maintenance

The Go module, C binding, and canonical engine use one lockstep version. For a
release ref `vX.Y.Z`, the C header macros, Go version claims, and bundled static
archives must all identify `X.Y.Z`. Run the release check from the repository
root:

```sh
./scripts/check-release.sh vX.Y.Z
```

The archive change may provide the shared C source ref in the one-line file
`internal/native/lib/sidereon-c.ref`. The check consumes that file when it is
present. A checkout without that file uses the documented current-reference
fallback inside the check script. Keep the file to one non-empty line and do
not add a second source-ref setting.

The current preparation uses a commit ref because the public `v1.3.0` tag does
not exist. The check verifies the committed C header against that commit and
reports publication as blocked until a public `v1.3.0` C ref carries matching
`1.3.0` macros. This is an expected pre-release result, not a release.

For a release candidate, run the normal checks and the packed consumer check:

```sh
./scripts/smoke-fixtures.sh
./scripts/check-release.sh vX.Y.Z
./scripts/test-packed-module.sh
```

The packed consumer uses only tracked repository content, places it in a
temporary module outside the checkout, downloads it through a local module
replacement, and performs a numerical solve. It does not publish or contact a
module proxy.
