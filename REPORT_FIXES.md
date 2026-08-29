# Review finding fixes

Date: 2026-08-29

Branch: `fix-review-findings`

## Defects and regressions

### Error propagation and nil-result invariants

- `FusionFilter.UpdateTightSP3` now returns the error from `tightUpdate` through
  the SP3 handle callback instead of overwriting it with the callback's `nil`
  result.
- `SolveFDESPP` and `SolveFDEBroadcast` now return `solveFDE` errors through
  their source-handle callbacks.
- All four public FDE constructors use `newFDEResult`, which rejects a nil
  native `FDESolution` with `errNilNativeHandle`.
- `RangeFDEResult.Output` now returns the excluded-list copy error from its
  handle callback.
- Commit: `e08cf64e Preserve native operation errors`.

Regression coverage:

- `TestUpdateTightSP3PropagatesErrorsInEveryMode` covers a closed filter and an
  embedded-NUL tight satellite ID in direct, recorded, and time-sync modes.
- `TestNonRobustFDEPropagatesInputAndNativeErrors` covers embedded-NUL
  observations and detailed native solve failures through both SP3 and
  broadcast entry points. It also requires a nil result at the failing solve
  call, preventing a deferred, misleading `ErrClosed`.
- `TestNewFDEResultRejectsNilNativeHandle` directly protects the public result
  invariant.
- `TestHandleCallbacksDoNotEraseCapturedErrors` parses every production Go file
  in `internal/native` and rejects a callback that assigns to its receiving
  outer error while returning a literal. This also protects the otherwise
  unreachable excluded-list accessor failure in `RangeFDEResult.Output`.

Before the fix, the tight test reported `nil` for all six closed/NUL cases, the
four FDE source/failure combinations returned `&FDEResult{handle:nil}` with no
solve error, and the callback audit reported all four sites listed below.

### Recorded-history locking

- Fusion tight recorded updates and `TrackFilter.PredictRecorded` now acquire
  the history builder with `surfaceHandle.write`.
- Commit: `2375cd6c Serialize recorded history mutations`.

Regression coverage:

- `TestPredictRecordedRequestsExclusiveBuilderLock` and
  `TestFusionTightRecordedRequestsExclusiveBuilderLock` hold a builder reader
  lock and prove that each real recorded entry point waits for exclusive
  access.
- `TestRecordedMutationExcludesFinishAndSecondMutation` holds the same
  exclusive lock used around a native recorded callback and proves that both
  `Finish` and a second mutation on the shared builder remain blocked.

These are lock-level tests. A native callback barrier was not available in the
pinned static C archive, while direct access to the package-private
`surfaceHandle` lock makes the requested shared-versus-exclusive property
deterministic. Before the fix, both real recorded entry points completed while
the test held a builder reader lock.

### FDE option allocation cleanup

- `nativeFDEOptions` validates every satellite weight ID before allocating the
  C weight array or any C strings. A later embedded NUL therefore cannot strand
  strings allocated for earlier entries.
- Commit: `d7f38ec1 Validate FDE weights before allocation`.

Regression coverage:

- `TestFDEInvalidLaterWeightDoesNotLeakEarlierCString` exercises the sorted
  two-key map `{"A": 1, "Z\x00bad": 1}` 4,096 times through the already-correct
  robust FDE error path.
- The same test repeats a one-MiB first key 64 times and checks native resident
  memory high-water growth on Darwin/Linux. Before the fix this deterministic
  probe grew native resident memory by 68,780,032 bytes; after prevalidation it
  remains below the 32-MiB failure threshold.

Go's `-asan` mode was also attempted for the focused leak regression, but Go
1.27 reports `-asan is not supported on darwin/arm64` on this host. No
sanitizer-enabled build of the pinned native archive is bundled.

## Complete callback error audit

The whole-package AST audit found exactly these callbacks that assigned to the
same outer error receiving the handle callback result and then returned a
literal:

1. `internal/native/fusion_surface.go`: `FusionFilter.UpdateTightSP3`.
2. `internal/native/fde_surface.go`: `SolveFDESPP`.
3. `internal/native/fde_surface.go`: `SolveFDEBroadcast`.
4. `internal/native/integrity_surface.go`: `RangeFDEResult.Output` excluded-list
   copy.

All four are fixed. No additional error-erasing occurrence remains. Callbacks
that store a C status in a distinct `opErr` and explicitly check it after the
handle callback are not error-erasing and were left unchanged.

## Complete builder acquisition audit

`internal/native` contains two history builder types.

`FusionRTSHistoryBuilder` acquisitions:

- `PropagateRecorded`: `write`.
- `UpdateLooseRecorded`: `write`.
- `UpdateStationaryRecorded`: `write`.
- `UpdateNonHolonomicRecorded`: `write`.
- Recorded SP3 and broadcast tight updates through `tightUpdate`: `write`
  (fixed).
- `Finish`: `read`; correct because the bundled C declaration takes a const
  builder and clones/reads it.

`TrackRTSHistoryBuilder` acquisitions:

- `PredictRecorded`: `write` (fixed).
- `RecordPredictionOnly`: `write`.
- `UpdatePositionRecorded`: `write`.
- `UpdatePositionGatedRecorded`: `write`.
- `Finish`: `read`; correct because the bundled C declaration takes a const
  builder.

The remaining history acquisition in `SmoothTrackRTS` is a read of an immutable
finished `TrackRTSHistory`, not a builder. Builder-from-filter constructors read
the filter being snapshotted and do not acquire an existing builder. No other
builder mutation uses a shared lock.

## Final gate output

All commands exited 0.

### `gofmt -l .`

```text
(no output)
```

### `go vet ./...`

```text
(no output)
```

### `golangci-lint run ./...`

CI-pinned `golangci-lint` v2.12.2 was installed into a temporary directory
because it was not initially on `PATH`.

```text
0 issues.
```

### `./scripts/check-doc-comments.sh`

```text
documentation: 3051 exported declarations and interface methods plus 2352 documented exported struct fields have conventional comments
```

### `LC_ALL=C ./scripts/check-abi-coverage.sh`

```text
ABI coverage: 1503 total; 1492 direct, 10 composed, 1 excluded; map exact
```

### `env -u LC_ALL LANG=en_US.UTF-8 ./scripts/check-abi-coverage.sh`

```text
ABI coverage: 1503 total; 1492 direct, 10 composed, 1 excluded; map exact
```

### `GOEXPERIMENT=cgocheck2 go test ./... -race -count=1`

```text
ok  	github.com/neilberkman/sidereon-go	14.660s
?   	github.com/neilberkman/sidereon-go/internal/checkdoc	[no test files]
ok  	github.com/neilberkman/sidereon-go/internal/native	2.499s
```

### `./scripts/smoke-fixtures.sh`

```text
ok  	github.com/neilberkman/sidereon-go	1.424s
?   	github.com/neilberkman/sidereon-go/internal/checkdoc	[no test files]
ok  	github.com/neilberkman/sidereon-go/internal/native	0.497s [no tests to run]
ok  	github.com/neilberkman/sidereon-go	1.409s
?   	github.com/neilberkman/sidereon-go/internal/checkdoc	[no test files]
ok  	github.com/neilberkman/sidereon-go/internal/native	0.475s [no tests to run]
```

### `./scripts/test-packed-module.sh`

```text
go: creating new go.mod: module example.com/sidereon-consumer
go: added github.com/neilberkman/sidereon-go v0.0.0
numerical solve: {[4.484127992418275e+06 550581.6853698192 4.487560540876184e+06] 0.00010006922594612232 8 [G08 G10 G16 G18 G20 G21 G26 G27] [-3.166496753692627e-07 -0.0007014274597167969 -0.00010690838098526001 0.000359412282705307 -0.00011412426829338074 -0.00015893951058387756 0.0005824938416481018 -3.2376497983932495e-05] 0x4362f6fce240 0x4362f6fb2168 {7 true 2 false false 0 false 0 8 1 4 true {3 4 4 11.56425099674148 4.5258697252838935 true true}}}
```
