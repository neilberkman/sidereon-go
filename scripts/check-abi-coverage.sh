#!/usr/bin/env bash
set -euo pipefail

# Every sort below must order bytes, not locale collation. Without this the
# generated map differs between a C-collation environment and a UTF-8 one
# (for example sidereon_covariance6_* against sidereon_covariance_*), so the
# committed artifact would only validate on the machine that wrote it.
export LC_ALL=C

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

# Production bridge helpers are private inline C definitions, rather than
# symbols supplied by the vendored library. Require their local definitions
# and account for them separately from public ABI routes.
cut -f1 "$work/direct-all" | sort -u >"$work/all-call-symbols"
comm -13 "$work/header" "$work/all-call-symbols" >"$work/private-calls"
printf '%s\n' sidereon_enter_c_thread sidereon_get_c_thread_depth sidereon_leave_c_thread | sort >"$work/private-expected"
if ! cmp -s "$work/private-expected" "$work/private-calls"; then
    printf 'ABI coverage: unexpected production calls outside the vendored header\n' >&2
    cat "$work/private-calls" >&2
    exit 1
fi
while IFS= read -r helper; do
    if ! grep -qE "^static inline (int|void) ${helper}\\(void\\)" internal/native/bridge.go; then
        printf 'ABI coverage: private bridge helper %s has no inline definition\n' "$helper" >&2
        exit 1
    fi
done <"$work/private-calls"
awk -F '\t' 'NR==FNR {header[$1]=1; next} header[$1] && !seen[$1]++ {print $1 "\t`" $2 "` contains the production cgo call."}' \
    "$work/header" "$work/direct-all" >"$work/direct-proof"
cut -f1 "$work/direct-proof" >"$work/direct"

# Every composition is a reviewed operation recipe, with a concrete public
# entry and production native replacements. Unknown functions remain missing
# dispositions and cause failure; exclusions are not accepted.
# The table is manually reviewed. This checker verifies declared public entries
# and production replacement call sites; it does not infer or prove semantic
# equivalence. Validate the six-column contract before parsing; Bash read would
# otherwise fold surplus tab-separated fields into the final proof column.
awk -F '\t' 'NF == 0 || $0 ~ /^[[:space:]]*#/ { next } NF != 6 { print "ABI coverage: expected exactly 6 TSV columns at line " NR ", found " NF > "/dev/stderr"; bad=1 } END { exit bad }' audit/ABI_COMPOSITIONS.tsv
# Reject duplicate symbols before any normalization can hide them.
awk -F '\t' 'NF == 0 || $1 ~ /^#/ { next } seen[$1]++ { print "ABI coverage: duplicate composition symbol " $1 > "/dev/stderr"; bad=1 } END { exit bad }' audit/ABI_COMPOSITIONS.tsv
: >"$work/composed-proof"
while IFS=$'\t' read -r symbol function source native_source replacements proof; do
    [[ "$symbol" == \#* || -z "$symbol" ]] && continue
    if [[ -z "$function" || -z "$source" || -z "$native_source" || -z "$replacements" || -z "$proof" ]]; then
        printf 'ABI coverage: incomplete composition row for %s\n' "$symbol" >&2
        exit 1
    fi
    for entry in $function; do
        entry_found=false
        for public_path in $source; do
            if grep -qE "^func[[:space:]]+(\\([^)]*\\)[[:space:]]+)?${entry}\\(" "$public_path"; then
                entry_found=true
            fi
        done
        if [[ "$entry_found" != true ]]; then
            printf 'ABI coverage: composed entry %s is absent from listed public sources %s\n' "$entry" "$source" >&2
            exit 1
        fi
    done
    for replacement in $replacements; do
        replacement_found=false
        for native_path in $native_source; do
            if grep -Fq "C.${replacement}(" "$native_path"; then
                replacement_found=true
            fi
        done
        if [[ "$replacement_found" != true ]] || ! grep -Fxq "$replacement" "$work/direct"; then
            printf 'ABI coverage: composition %s lacks direct replacement %s in %s\n' "$symbol" "$replacement" "$native_source" >&2
            exit 1
        fi
    done
    printf '%s\t`%s` in `%s`, through `%s`: %s\n' "$symbol" "$function" "$source" "$native_source" "$proof" >>"$work/composed-proof"
done <audit/ABI_COMPOSITIONS.tsv
LC_ALL=C sort "$work/composed-proof" -o "$work/composed-proof"
cut -f1 "$work/composed-proof" >"$work/composed"

: >"$work/excluded-proof"
: >"$work/excluded"

header_count=$(wc -l <"$work/header" | tr -d ' ')
direct_count=$(wc -l <"$work/direct" | tr -d ' ')
composed_count=$(wc -l <"$work/composed" | tr -d ' ')
excluded_count=$(wc -l <"$work/excluded" | tr -d ' ')

# The reviewed header adds complete SSR projection and RINEX finding-detail routes
# while retaining legacy projection symbols; every declaration still has one disposition.
[[ "$header_count" == 2059 ]] || {
    printf 'ABI coverage: expected 2059 pinned v3 header declarations, found %s\n' "$header_count" >&2
    exit 1
}

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
	printf 'Summary: **%s total = %s direct + %s composed + 0 excluded**. Direct rows name production cgo calls. Composed rows describe manually reviewed Go recipes and native replacements; this checker verifies declared entries and replacement call sites, not semantic equivalence. Private inline thread helpers are accounted for separately. This is C-symbol coverage; canonical Rust API parity is a separate audit.\n\n' "$header_count" "$direct_count" "$composed_count"
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
