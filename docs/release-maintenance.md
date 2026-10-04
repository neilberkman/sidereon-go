# Release maintenance

The Go module, C binding, and canonical engine use one lockstep version. For a
release ref `vX.Y.Z`, the C header macros, Go version claims, and bundled static
archives must all identify `X.Y.Z`. Run the release check from the repository
root:

```sh
./scripts/check-release.sh vX.Y.Z
```

The shared C source ref is the one-line file
`internal/native/lib/sidereon-c.ref`. Both the archive builder and release
check consume that file. Keep it to one non-empty line and do not add a second
source-ref setting.

The prior `3.0.0` archive set was built from C tag `v3.0.0` at commit
`a41ef678f8c81ce1ce58221bb3128ccd64511f23`, with its matching generated
header. For each release, set the single C source ref to the final reviewed C
release tag, verify that it resolves to the intended source, and rebuild and
validate all seven archives against that ref and its generated header. Keep the
ref, header, manifest, and archives synchronized. A successful archive-format
verification alone does not establish that each target builds or links
correctly.

For a release candidate, run the normal checks and the packed consumer check:

```sh
./scripts/smoke-fixtures.sh
./scripts/test-native-archives.sh
./scripts/check-release.sh vX.Y.Z
./scripts/test-packed-module.sh
```

The packed consumer uses only tracked repository content, places it in a
temporary module outside the checkout, resolves it through a local module
replacement, and performs a numerical solve. It does not publish or contact a
module proxy. The archive verifier checks the complete seven-target set,
manifest, hashes, archive formats, source identity, and release metadata.
