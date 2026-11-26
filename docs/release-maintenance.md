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

The 4.2.0 archive set requires C tag `v4.2.0` at commit
`0a81c8a2f99efb6d277e062ea33ef354ba3919b8` and its matching header.
The ref, header, manifest, and all seven archives must identify the same release.
Archive-format verification alone does not establish that each target links.

For a release candidate, run the normal checks and the packed consumer check:

```sh
./scripts/smoke-fixtures.sh
./scripts/test-native-archives.sh
./scripts/check-release.sh vX.Y.Z
./scripts/test-packed-module.sh
```

The packed consumer uses the package files in the working tree, places it in a
temporary module outside the checkout, resolves it through a local module
replacement, and performs a numerical solve. It does not publish or contact a
module proxy. The archive verifier checks the complete seven-target set,
manifest, hashes, archive formats, source identity, and release metadata.
