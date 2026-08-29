#!/bin/sh
# Verify the version and source identity used by a release candidate.
#
# The archive change may add internal/native/lib/sidereon-c.ref. It must contain
# one non-empty C source ref, either vX.Y.Z or a 40-character commit ID. When
# that file is absent, this script uses exactly one fallback: the current C
# authority's commit ID below. The fallback is for local pre-release checking;
# it is not a release version.

set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
TARGET_VERSION=1.3.0
FALLBACK_C_REF=b38ecf8caf796a02f209dbb4cbebdaa4a042204c
FALLBACK_HEADER_SHA256=6e79405bbce65d91958fe591279ad4c73f79f86573c64176f7c3dd0d9c29420d
HEADER_REL=internal/native/include/sidereon.h
C_HEADER_REL=bindings/c/include/sidereon.h
C_REF_FILE=$ROOT/internal/native/lib/sidereon-c.ref
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

if [ -f "$C_REF_FILE" ]; then
	C_REF=$(sed -e 's/[[:space:]]*#.*$//' -e '/^[[:space:]]*$/d' "$C_REF_FILE" | awk 'NR == 1 { value=$0 } NR > 1 { count++ } END { if (count) exit 1; print value }') || {
		echo "check-release: $C_REF_FILE must contain exactly one non-empty line" >&2
		exit 2
	}
	C_REF=$(printf '%s' "$C_REF" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')
	if [ -z "$C_REF" ]; then
		echo "check-release: $C_REF_FILE is empty" >&2
		exit 2
	fi
else
	C_REF=$FALLBACK_C_REF
fi

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

if [ "$HEADER_VERSION" != "$MAJOR.$MINOR.$PATCH" ]; then
	echo "check-release: header version string $HEADER_VERSION disagrees with macros $MAJOR.$MINOR.$PATCH" >&2
	exit 1
fi

if [ "$HEADER_VERSION" != "$EXPECTED_VERSION" ]; then
	if [ "$RELEASE_MODE" -eq 0 ] && [ "$C_REF" = "$FALLBACK_C_REF" ] && [ "$HEADER_VERSION" = "1.2.0" ]; then
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
if [ -z "$C_REPO" ] && [ -d "$ROOT/../../repos/sidereon-c/.git" ]; then
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
elif [ "$C_REF" = "$FALLBACK_C_REF" ] && [ "$(checksum "$HEADER")" = "$FALLBACK_HEADER_SHA256" ]; then
	echo "check-release: vendored header matches the recorded current C authority digest"
else
	echo "check-release: cannot verify exact C source/header identity; set SIDEREON_C_REPO or provide the pinned sibling checkout" >&2
	exit 1
fi

CLAIM_FILES=
if [ -f "$ROOT/README.md" ]; then
	CLAIM_FILES="$ROOT/README.md"
fi
for file in $(find "$ROOT" -type f -name '*.go' -not -path "$ROOT/.git/*"); do
	if [ -f "$file" ]; then
		CLAIM_FILES="$CLAIM_FILES $file"
	fi
done

CLAIMS=$(for file in $CLAIM_FILES; do
	grep -Eoh '[0-9]+\.[0-9]+\.[0-9]+' "$file" || true
done | sort -u)
for claim in $CLAIMS; do
	if [ "$claim" != "$TARGET_VERSION" ] && [ "$claim" != "$EXPECTED_VERSION" ]; then
		if [ "$RELEASE_MODE" -eq 0 ] && [ "$claim" = "$HEADER_VERSION" ]; then
			continue
		fi
		echo "check-release: unsupported Go/README version claim: $claim" >&2
		exit 1
	fi
done
if ! grep -Fq "$TARGET_VERSION" "$ROOT/README.md"; then
	echo "check-release: README.md does not state the target version $TARGET_VERSION" >&2
	exit 1
fi

if [ "$RELEASE_MODE" -eq 0 ]; then
	echo "check-release: PRE-RELEASE — publication remains blocked until public v$TARGET_VERSION exists with $TARGET_VERSION C header macros"
	if [ "$ALLOW_PRERELEASE" -ne 1 ]; then
		echo "check-release: pass --allow-prerelease for ordinary pre-release CI" >&2
		exit 2
	fi
else
	echo "check-release: release ref $RELEASE_REF agrees with header and Go/README claims"
fi
