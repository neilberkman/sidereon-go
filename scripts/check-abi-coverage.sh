#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
HEADER="$ROOT/internal/native/include/sidereon.h"
MAP="$ROOT/audit/ABI_IMPLEMENTATION_MAP.md"
MODE=${1:-check}

if [[ "$MODE" != "check" && "$MODE" != "--write" ]]; then
	printf 'usage: %s [check|--write]\n' "$0" >&2
	exit 2
fi

tmp_base=${TMPDIR:-"$ROOT/../tmp"}
if [[ ! -d "$tmp_base" ]]; then
	tmp_base="$ROOT/../tmp"
	mkdir -p "$tmp_base"
fi
work=$(mktemp -d "$tmp_base/sidereon-abi-coverage.XXXXXX")
trap 'rm -rf "$work"' EXIT

cd "$ROOT"

grep -oE 'sidereon_[A-Za-z0-9_]+[[:space:]]*\(' "$HEADER" \
	| sed -E 's/[[:space:]]*\($//' \
	| sort -u >"$work/header"

while IFS= read -r source; do
	{
		grep -oE 'C\.sidereon_[A-Za-z0-9_]+' "$source" || true
	} | awk -v source="${source#./}" '{ sub(/^C\./, ""); print $0 "\t" source }'
done < <(find . -type f -name '*.go' ! -name '*_test.go' ! -path './.git/*' | LC_ALL=C sort) \
	| sort -u >"$work/direct-all"

awk -F '\t' '!seen[$1]++ { print $1 "\t`" $2 "` contains the production cgo call." }' \
	"$work/direct-all" >"$work/direct-proof"
cut -f1 "$work/direct-proof" >"$work/direct"

printf '%s\t%s\n' \
	'sidereon_bias_sinex_load' '`LoadBiasSINEX` in `bias.go` reads or decompresses the path in Go, then uses the byte parser.' \
	'sidereon_bias_sinex_load_lossy' '`LoadBiasSINEXLossy` in `bias.go` reads or decompresses the path in Go, then uses the lossy byte parser.' \
	'sidereon_code_dcb_load' '`LoadCodeDCB` in `bias.go` reads the path in Go, then uses the strict byte parser.' \
	'sidereon_code_dcb_load_lossy' '`LoadCodeDCBLossy` in `bias.go` reads the path in Go, then uses the lossy byte parser.' \
	'sidereon_broadcast_ephemeris_load_nav' '`LoadBroadcastEphemeris` in `corrections.go` reads the path in Go, then uses the NAV byte parser.' \
	'sidereon_rinex_obs_load' '`LoadRINEXObservation` in `observation.go` reads the path in Go, then uses the RINEX observation byte parser.' \
	'sidereon_precise_interpolant_artifact_from_path' '`OpenPreciseInterpolantArtifactFile` in `remaining_products.go` reads the path in Go, then opens the detached artifact bytes.' \
	'sidereon_precise_interpolant_artifact_open_borrowed' '`OpenPreciseInterpolantArtifactBorrowed` in `remaining_products.go` delegates to the owned-byte artifact opener; Go slice lifetime is not borrowed across the call.' \
	'sidereon_write_dted_tile_list_to_mmap_store' '`WriteDTEDTileListToMMapStore` in `geodesy_environment.go` builds store bytes through the ABI and writes them with Go filesystem ownership.' \
	'sidereon_write_dted_tree_to_mmap_store' '`WriteDTEDTreeToMMapStore` in `geodesy_environment.go` builds store bytes through the ABI and writes them with Go filesystem ownership.' \
	>"$work/composed-proof"
sort -u "$work/composed-proof" -o "$work/composed-proof"
cut -f1 "$work/composed-proof" >"$work/composed"

while IFS=$'\t' read -r function source; do
	if ! grep -qE "^func[[:space:]]+${function}\\(" "$source"; then
		printf 'ABI coverage: composed route %s is absent from %s\n' "$function" "$source" >&2
		exit 1
	fi
done <<'EOF'
LoadBiasSINEX	bias.go
LoadBiasSINEXLossy	bias.go
LoadCodeDCB	bias.go
LoadCodeDCBLossy	bias.go
LoadBroadcastEphemeris	corrections.go
LoadRINEXObservation	observation.go
OpenPreciseInterpolantArtifactFile	remaining_products.go
OpenPreciseInterpolantArtifactBorrowed	remaining_products.go
WriteDTEDTileListToMMapStore	geodesy_environment.go
WriteDTEDTreeToMMapStore	geodesy_environment.go
EOF

printf '%s\t%s\n' \
	'sidereon_precise_interpolant_artifact_from_path_attested' 'Excluded: this path-only deferred-attestation route cannot preserve the Go-owned byte-transport contract; the ABI has no equivalent owned-byte attested opener.' \
	>"$work/excluded-proof"
cut -f1 "$work/excluded-proof" >"$work/excluded"

header_count=$(wc -l <"$work/header" | tr -d ' ')
direct_count=$(wc -l <"$work/direct" | tr -d ' ')
composed_count=$(wc -l <"$work/composed" | tr -d ' ')
excluded_count=$(wc -l <"$work/excluded" | tr -d ' ')

[[ "$header_count" == 1503 ]] || {
	printf 'ABI coverage: expected 1503 header declarations, found %s\n' "$header_count" >&2
	exit 1
}
[[ "$direct_count" == 1492 ]] || {
	printf 'ABI coverage: expected 1492 production cgo routes, found %s\n' "$direct_count" >&2
	exit 1
}
[[ "$composed_count" == 10 && "$excluded_count" == 1 ]] || {
	printf 'ABI coverage: expected 10 compositions and 1 exclusion, found %s and %s\n' "$composed_count" "$excluded_count" >&2
	exit 1
}

if ! comm -13 "$work/header" "$work/direct" >"$work/calls-outside-header"; then
	exit 1
fi
if [[ -s "$work/calls-outside-header" ]]; then
	printf 'ABI coverage: production calls absent from the vendored header:\n' >&2
	sed 's/^/  /' "$work/calls-outside-header" >&2
	exit 1
fi

for left in direct composed excluded; do
	for right in direct composed excluded; do
		[[ "$left" < "$right" ]] || continue
		comm -12 "$work/$left" "$work/$right" >"$work/overlap"
		if [[ -s "$work/overlap" ]]; then
			printf 'ABI coverage: %s/%s dispositions overlap:\n' "$left" "$right" >&2
			sed 's/^/  /' "$work/overlap" >&2
			exit 1
		fi
	done
done

sort -u "$work/direct" "$work/composed" "$work/excluded" >"$work/union"
if ! cmp -s "$work/header" "$work/union"; then
	printf 'ABI coverage: disposition union does not equal the vendored header\n' >&2
	printf 'Missing dispositions:\n' >&2
	comm -23 "$work/header" "$work/union" | sed 's/^/  /' >&2
	printf 'Unknown dispositions:\n' >&2
	comm -13 "$work/header" "$work/union" | sed 's/^/  /' >&2
	exit 1
fi

awk -F '\t' '{ print $1 "\tdirect\t" $2 }' "$work/direct-proof" >"$work/dispositions"
awk -F '\t' '{ print $1 "\tcomposed\t" $2 }' "$work/composed-proof" >>"$work/dispositions"
awk -F '\t' '{ print $1 "\texcluded\t" $2 }' "$work/excluded-proof" >>"$work/dispositions"
sort -u "$work/dispositions" -o "$work/dispositions"

pin=$(tr -d '[:space:]' <internal/native/lib/sidereon-c.ref)
{
	printf '# Current C ABI implementation map\n\n'
	printf 'This map is generated from the vendored header and production cgo calls for pinned public `sidereon-c` commit `%s`. Run `./scripts/check-abi-coverage.sh` to prove that every declaration has exactly one disposition.\n\n' "$pin"
	printf 'Summary: **1,503 total = 1,492 direct + 10 composed + 1 excluded**. Direct rows name a production cgo source. Composed rows keep filesystem acquisition or persistence in Go and delegate bytes to an implemented ABI route.\n\n'
	printf '| # | C symbol | Disposition | Implementation proof |\n'
	printf '|---:|---|---|---|\n'
	awk -F '\t' '{ printf "| %d | `%s` | %s | %s |\n", NR, $1, $2, $3 }' "$work/dispositions"
} >"$work/expected-map"

if [[ "$MODE" == "--write" ]]; then
	cp "$work/expected-map" "$MAP"
else
	if [[ ! -f "$MAP" ]] || ! cmp -s "$work/expected-map" "$MAP"; then
		printf 'ABI coverage: %s is missing or stale; run %s --write\n' "$MAP" "$0" >&2
		exit 1
	fi
fi

printf 'ABI coverage: %s total; %s direct, %s composed, %s excluded; map exact\n' \
	"$header_count" "$direct_count" "$composed_count" "$excluded_count"
