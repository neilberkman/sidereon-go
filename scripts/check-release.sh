#!/bin/sh
# Verify the version and source identity used by a release candidate.
#
# internal/native/lib/sidereon-c.ref is the single C source pin. It contains
# one non-empty ref, either vX.Y.Z or a 40-character commit ID.

set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
HEADER_REL=internal/native/include/sidereon.h
C_HEADER_REL=bindings/c/include/sidereon.h
C_REF_FILE=$ROOT/internal/native/lib/sidereon-c.ref
ARCHIVE_MANIFEST=$ROOT/internal/native/lib/manifest.sha256
# The version a pre-release build targets is the archive manifest release
# target, so the README claim check follows each release instead of a constant.
TARGET_VERSION=$(awk -F= '$1 == "# release_target" { print $2; exit }' "$ARCHIVE_MANIFEST" 2>/dev/null || true)
[ -n "$TARGET_VERSION" ] || {
	echo "check-release: $ARCHIVE_MANIFEST has no release_target" >&2
	exit 2
}
ALLOW_PRERELEASE=0
REQUESTED_REF=

while [ "$#" -gt 0 ]; do
	case "$1" in
		--allow-prerelease)
			ALLOW_PRERELEASE=1
			;;
		--help)
			echo "usage: $0 [--allow-prerelease] [vX.Y.Z|commit]"
			exit 0
			;;
		-*)
			echo "check-release: unknown option: $1" >&2
			exit 2
			;;
		*)
			if [ -n "$REQUESTED_REF" ]; then
				echo "check-release: only one release/source ref may be supplied" >&2
				exit 2
			fi
			REQUESTED_REF=$1
			;;
	esac
	shift
done

[ -f "$C_REF_FILE" ] || {
	echo "check-release: missing single source pin $C_REF_FILE" >&2
	exit 2
}
C_REF=$(sed -e 's/[[:space:]]*#.*$//' -e '/^[[:space:]]*$/d' "$C_REF_FILE" | awk 'NR == 1 { value=$0 } NR > 1 { count++ } END { if (count) exit 1; print value }') || {
	echo "check-release: $C_REF_FILE must contain exactly one non-empty line" >&2
	exit 2
}
C_REF=$(printf '%s' "$C_REF" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')
[ -n "$C_REF" ] || {
	echo "check-release: $C_REF_FILE is empty" >&2
	exit 2
}

if [ -n "$REQUESTED_REF" ]; then
	RELEASE_REF=$REQUESTED_REF
else
	RELEASE_REF=$C_REF
fi

is_commit_ref() {
	printf '%s\n' "$1" | grep -Eq '^[0-9a-fA-F]{40}$'
}

is_release_ref() {
	printf '%s\n' "$1" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'
}

if is_release_ref "$RELEASE_REF"; then
	EXPECTED_VERSION=${RELEASE_REF#v}
	RELEASE_MODE=1
elif is_commit_ref "$RELEASE_REF"; then
	EXPECTED_VERSION=$TARGET_VERSION
	RELEASE_MODE=0
else
	echo "check-release: ref must be vX.Y.Z or a 40-character commit ID: $RELEASE_REF" >&2
	exit 2
fi

if [ -n "$REQUESTED_REF" ] && [ "$C_REF" != "$REQUESTED_REF" ]; then
	echo "check-release: release ref $RELEASE_REF does not agree with C source ref $C_REF" >&2
	exit 1
fi

HEADER=$ROOT/$HEADER_REL
if [ ! -f "$HEADER" ]; then
	echo "check-release: missing vendored header: $HEADER_REL" >&2
	exit 1
fi

header_macro() {
	awk -v wanted="$1" '$1 == "#define" && $2 == wanted { value=$3 } END { if (value != "") print value }' "$HEADER"
}

MAJOR=$(header_macro SIDEREON_VERSION_MAJOR)
MINOR=$(header_macro SIDEREON_VERSION_MINOR)
PATCH=$(header_macro SIDEREON_VERSION_PATCH)
HEADER_VERSION=$(header_macro SIDEREON_VERSION_STRING | sed 's/^"//; s/"$//')
for value in "$MAJOR" "$MINOR" "$PATCH" "$HEADER_VERSION"; do
	if [ -z "$value" ]; then
		echo "check-release: incomplete SIDEREON_VERSION_* macros in $HEADER_REL" >&2
		exit 1
	fi
done

MODULE_PATH=$(awk '$1 == "module" { print $2; exit }' "$ROOT/go.mod" 2>/dev/null || true)
case "$MODULE_PATH" in
	*/v[0-9]*) MODULE_MAJOR=${MODULE_PATH##*/v} ;;
	*) MODULE_MAJOR= ;;
esac
case "$MODULE_MAJOR" in
	''|*[!0-9]*)
		echo "check-release: go.mod module path has no numeric major-version suffix: $MODULE_PATH" >&2
		exit 1
		;;
esac
if [ "$MODULE_MAJOR" != "$MAJOR" ]; then
	echo "check-release: go.mod module major v$MODULE_MAJOR does not agree with header major $MAJOR" >&2
	exit 1
fi
README_MODULE_PATHS=$(grep -Eo 'go get sidereon\.dev/go/v[0-9]+' "$ROOT/README.md" \
	| sed 's/^go get //' \
	| sort -u || true)
[ -n "$README_MODULE_PATHS" ] || {
	echo "check-release: README.md has no Go module import-path example" >&2
	exit 1
}
for module_path in $README_MODULE_PATHS; do
	if [ "$module_path" != "$MODULE_PATH" ]; then
		echo "check-release: README module path $module_path does not agree with go.mod module path $MODULE_PATH" >&2
		exit 1
	fi
done

if [ "$HEADER_VERSION" != "$MAJOR.$MINOR.$PATCH" ]; then
	echo "check-release: header version string $HEADER_VERSION disagrees with macros $MAJOR.$MINOR.$PATCH" >&2
	exit 1
fi

if [ "$HEADER_VERSION" != "$EXPECTED_VERSION" ]; then
	if [ "$RELEASE_MODE" -eq 0 ] && [ "$HEADER_VERSION" = "1.2.0" ]; then
		echo "check-release: pre-release header is $HEADER_VERSION; target is $TARGET_VERSION"
	else
		echo "check-release: header version $HEADER_VERSION does not agree with source ref version $EXPECTED_VERSION" >&2
		exit 1
	fi
fi

checksum() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	else
		shasum -a 256 "$1" | awk '{print $1}'
	fi
}

C_REPO=${SIDEREON_C_REPO:-}
if [ -z "$C_REPO" ] && [ -d "$ROOT/../repos/sidereon-c/.git" ]; then
	C_REPO=$ROOT/../repos/sidereon-c
elif [ -z "$C_REPO" ] && [ -d "$ROOT/../../repos/sidereon-c/.git" ]; then
	C_REPO=$ROOT/../../repos/sidereon-c
fi

if [ -n "$C_REPO" ] && git -C "$C_REPO" rev-parse --git-dir >/dev/null 2>&1; then
	if ! git -C "$C_REPO" cat-file -e "$C_REF^{commit}:$C_HEADER_REL" 2>/dev/null; then
		echo "check-release: C ref $C_REF has no $C_HEADER_REL" >&2
		exit 1
	fi
	if ! git -C "$C_REPO" show "$C_REF:$C_HEADER_REL" | cmp -s - "$HEADER"; then
		echo "check-release: vendored header is not byte-identical to C ref $C_REF" >&2
		exit 1
	fi
	echo "check-release: vendored header matches C ref $C_REF"
elif [ -f "$ARCHIVE_MANIFEST" ] &&
	[ "$(awk -F= '$1 == "# source_ref" { print $2; exit }' "$ARCHIVE_MANIFEST")" = "$C_REF" ] &&
	[ "$(awk -F= '$1 == "# source_header_sha256" { print $2; exit }' "$ARCHIVE_MANIFEST")" = "$(checksum "$HEADER")" ]; then
	echo "check-release: vendored header matches the verified archive manifest for C ref $C_REF"
else
	echo "check-release: cannot verify exact C source/header identity; set SIDEREON_C_REPO or provide the pinned sibling checkout" >&2
	exit 1
fi

# Validate only explicit Go-module release claims. README toolchain requirements
# (for example Rust 1.98.1) and unrelated dependency versions are not Go
# release claims and must not be compared with the Sidereon module version.
CLAIMS=$(grep -E 'go get sidereon\.dev/go/v[0-9]+@v[0-9]+\.[0-9]+\.[0-9]+|Go module uses' "$ROOT/README.md" \
	| grep -Eoh 'v?[0-9]+\.[0-9]+\.[0-9]+' \
	| sed 's/^v//' \
	| sort -u || true)
[ -n "$CLAIMS" ] || {
	echo "check-release: README.md has no explicit Go module release-version claim" >&2
	exit 1
}
for claim in $CLAIMS; do
	if [ "$claim" != "$TARGET_VERSION" ] && [ "$claim" != "$EXPECTED_VERSION" ]; then
		if [ "$RELEASE_MODE" -eq 0 ] && [ "$claim" = "$HEADER_VERSION" ]; then
			continue
		fi
		echo "check-release: unsupported Go module release claim: $claim" >&2
		exit 1
	fi
done
if ! grep -Fq "$EXPECTED_VERSION" "$ROOT/README.md"; then
	echo "check-release: README.md does not state the expected version $EXPECTED_VERSION" >&2
	exit 1
fi

if [ "$RELEASE_MODE" -eq 0 ]; then
	echo "check-release: PRE-RELEASE — publication remains blocked until public v$TARGET_VERSION exists with $TARGET_VERSION C header macros"
	if [ "$ALLOW_PRERELEASE" -ne 1 ]; then
		echo "check-release: pass --allow-prerelease for ordinary pre-release CI" >&2
			exit 2
		fi
fi
echo "check-release: release ref $RELEASE_REF agrees with header, module path, and Go/README claims"
