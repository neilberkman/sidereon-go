#!/usr/bin/env bash
set -Eeuo pipefail
IFS=$'\n\t'

# Build the seven public sidereon-c static libraries consumed by sidereon-go.
# The only source selector is internal/native/lib/sidereon-c.ref. It contains
# the current C commit today and changes to v1.3.0 when that public tag exists.
#
# Target matrix:
#   darwin/arm64       aarch64-apple-darwin
#   darwin/amd64       x86_64-apple-darwin
#   linux/amd64 glibc  x86_64-unknown-linux-gnu, glibc floor 2.17
#   linux/arm64 glibc  aarch64-unknown-linux-gnu, glibc floor 2.17
#   linux/amd64 musl   x86_64-unknown-linux-musl
#   linux/arm64 musl   aarch64-unknown-linux-musl
#   windows/amd64 GNU  x86_64-pc-windows-gnu
#
# cargo-zigbuild supplies the linker/sysroot for glibc-floor, musl, and Windows
# GNU builds. The host libc is never used to select a Linux target libc.

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)
ROOT_DIR=$(cd "$SCRIPT_DIR/.." && pwd -P)
LIB_DIR="$ROOT_DIR/internal/native/lib"
BUILD_ROOT="${NATIVE_BUILD_ROOT:-$ROOT_DIR/.native-build}"
TARGET_DIR="${CARGO_TARGET_DIR:-$BUILD_ROOT/target}"
C_REF_FILE="$LIB_DIR/sidereon-c.ref"

[[ -f "$C_REF_FILE" ]] || { echo "native archive error: missing $C_REF_FILE" >&2; exit 1; }
SIDEREON_C_REF=$(awk '
  /^[[:space:]]*(#|$)/ { next }
  { value=$0; count++ }
  END { if (count != 1) exit 1; print value }
' "$C_REF_FILE") || { echo "native archive error: $C_REF_FILE must contain exactly one source ref" >&2; exit 1; }
SIDEREON_C_REF=$(printf '%s' "$SIDEREON_C_REF" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')
SIDEREON_C_REPOSITORY="${SIDEREON_C_REPOSITORY:-https://github.com/neilberkman/sidereon-c.git}"
SIDEREON_C_SOURCE="${SIDEREON_C_SOURCE:-}"
RELEASE_VERSION="${SIDEREON_RELEASE_VERSION:-1.3.3}"
GLIBC_FLOOR="${SIDEREON_GLIBC_FLOOR:-2.17}"
CARGO_ZIGBUILD_VERSION="${CARGO_ZIGBUILD_VERSION:-0.23.3}"

CARGO="${CARGO:-$HOME/.cargo/bin/cargo}"
RUSTC="${RUSTC:-$HOME/.cargo/bin/rustc}"
RUSTUP="${RUSTUP:-$HOME/.cargo/bin/rustup}"
ZIG="$HOME/.local/share/mise/installs/zig/0.15.2/bin/zig"

# Keep the requested tool precedence. CARGO_HOME/bin is appended so an
# installed cargo-zigbuild is found without displacing the required paths.
export PATH="/opt/homebrew/bin:$HOME/.cargo/bin:$HOME/.local/share/mise/installs/zig/0.15.2/bin:${CARGO_HOME:-$ROOT_DIR/.cargo-home}/bin:$PATH"
export CARGO_HOME="${CARGO_HOME:-$ROOT_DIR/.cargo-home}"
export GOPATH="${GOPATH:-$ROOT_DIR/.gopath}"
export GOMODCACHE="${GOMODCACHE:-$GOPATH/pkg/mod}"
export GOCACHE="${GOCACHE:-$ROOT_DIR/.gocache}"
export TMPDIR="${TMPDIR:-/tmp}"
export CARGO_TARGET_DIR="$TARGET_DIR"
export ZIG_LOCAL_CACHE_DIR="$BUILD_ROOT/zig-local-cache"
export ZIG_GLOBAL_CACHE_DIR="$BUILD_ROOT/zig-global-cache"
export CARGO_ZIGBUILD_CACHE_DIR="$BUILD_ROOT/cargo-zigbuild-cache"
export RUSTC

MODE="build"
if [[ "${1:-}" == "--verify" ]]; then
  MODE="verify"
  shift
fi
if [[ $# -ne 0 ]]; then
  echo "usage: $0 [--verify]" >&2
  exit 2
fi

die() {
  echo "native archive error: $*" >&2
  exit 1
}

note() {
  echo "native archive: $*"
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

file_size() {
  wc -c < "$1" | tr -d '[:space:]'
}

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

require_hash_command() {
  command -v sha256sum >/dev/null 2>&1 || command -v shasum >/dev/null 2>&1 || die "sha256sum or shasum is required"
}

header_macro() {
  local header=$1
  local macro=$2
  awk -v wanted="$macro" '
    $1 == "#define" && $2 == wanted { value = $3; count++ }
    END { if (count != 1) exit 1; print value }
  ' "$header"
}

header_version() {
  local header=$1
  local major minor patch string
  major=$(header_macro "$header" SIDEREON_VERSION_MAJOR) || die "missing SIDEREON_VERSION_MAJOR in $header"
  minor=$(header_macro "$header" SIDEREON_VERSION_MINOR) || die "missing SIDEREON_VERSION_MINOR in $header"
  patch=$(header_macro "$header" SIDEREON_VERSION_PATCH) || die "missing SIDEREON_VERSION_PATCH in $header"
  string=$(header_macro "$header" SIDEREON_VERSION_STRING) || die "missing SIDEREON_VERSION_STRING in $header"
  string=${string#\"}
  string=${string%\"}
  [[ "$string" == "$major.$minor.$patch" ]] || die "header version macros disagree in $header"
  printf '%s.%s.%s\n' "$major" "$minor" "$patch"
}

validate_archive() {
  local archive=$1
  [[ -f "$archive" && ! -L "$archive" ]] || die "missing or non-regular archive: $archive"
  local size magic members
  size=$(file_size "$archive")
  [[ "$size" =~ ^[0-9]+$ && "$size" -ge 1024 ]] || die "archive is empty or implausibly small: $archive ($size bytes)"
  magic=$(dd if="$archive" bs=1 count=8 2>/dev/null | od -An -t x1 | tr -d '[:space:]')
  [[ "$magic" == "213c617263683e0a" ]] || die "archive has no ar magic: $archive"
  members=$(ar -t "$archive" 2>/dev/null | wc -l | tr -d '[:space:]')
  [[ "$members" =~ ^[0-9]+$ && "$members" -gt 0 ]] || die "archive has no readable ar members: $archive"
  if strings "$archive" | grep -Eq '(/Volumes/|/Users/|/private/var/|/home/[^/[:space:]]+/)'; then
    die "archive contains an absolute host path: $archive"
  fi
}

source_is_public() {
  local source=$1
  local remote
  remote=$(git -C "$source" config --get remote.origin.url || true)
  [[ -z "$remote" ]] && return 0
  case "$remote" in
    https://github.com/neilberkman/sidereon-c|https://github.com/neilberkman/sidereon-c.git|git@github.com:neilberkman/sidereon-c.git|ssh://git@github.com/neilberkman/sidereon-c.git)
      ;;
    *)
      die "source remote is not the public sidereon-c repository: $remote"
      ;;
  esac
}

verify_source() {
  local source=$1
  [[ -d "$source/.git" || -f "$source/.git" ]] || die "source is not a Git checkout: $source"
  [[ -z "$(git -C "$source" status --porcelain --untracked-files=all)" ]] || die "source checkout is not clean: $source"
  source_is_public "$source"
  local commit expected
  commit=$(git -C "$source" rev-parse --verify HEAD^{commit}) || die "cannot resolve source HEAD"
  expected=$(git -C "$source" rev-parse --verify "${SIDEREON_C_REF}^{commit}" 2>/dev/null) || die "source does not contain ref $SIDEREON_C_REF"
  [[ "$commit" == "$expected" ]] || die "source HEAD $commit does not match $SIDEREON_C_REF ($expected)"
  if [[ "$SIDEREON_C_REF" =~ ^[0-9a-fA-F]{40}$ ]]; then
    local lower_ref
    lower_ref=$(printf '%s' "$SIDEREON_C_REF" | tr '[:upper:]' '[:lower:]')
    [[ "$commit" == "$lower_ref" ]] || die "hash source ref is not the checked out commit"
  fi
  printf '%s\n' "$commit"
}

checkout_source() {
  if [[ -n "$SIDEREON_C_SOURCE" ]]; then
    SOURCE_DIR=$(cd "$SIDEREON_C_SOURCE" && pwd -P)
    return
  fi
  require_command git
  mkdir -p "$BUILD_ROOT"
  local parent source
  parent=$(mktemp -d "$TMPDIR/sidereon-c.XXXXXX")
  source="$parent/source"
  SOURCE_TEMP_PARENT="$parent"
  git clone --quiet "$SIDEREON_C_REPOSITORY" "$source"
  git -C "$source" checkout --quiet --detach "$SIDEREON_C_REF"
  SOURCE_DIR="$source"
}

manifest_meta() {
  local key=$1
  awk -F= -v wanted="# $key" '$1 == wanted { sub(/^# [^=]+=/, ""); print; exit }' "$LIB_DIR/manifest.sha256"
}

expected_archive_names() {
  printf '%s\n' \
    libsidereon_darwin_arm64.a \
    libsidereon_darwin_amd64.a \
    libsidereon_linux_amd64_glibc.a \
    libsidereon_linux_arm64_glibc.a \
    libsidereon_linux_amd64_musl.a \
    libsidereon_linux_arm64_musl.a \
    libsidereon_windows_amd64_gnu.a
}

expected_archive_metadata() {
  case "$1" in
    libsidereon_darwin_arm64.a) printf '%s\t%s\n' aarch64-apple-darwin none ;;
    libsidereon_darwin_amd64.a) printf '%s\t%s\n' x86_64-apple-darwin none ;;
    libsidereon_linux_amd64_glibc.a) printf '%s\t%s\n' x86_64-unknown-linux-gnu glibc ;;
    libsidereon_linux_arm64_glibc.a) printf '%s\t%s\n' aarch64-unknown-linux-gnu glibc ;;
    libsidereon_linux_amd64_musl.a) printf '%s\t%s\n' x86_64-unknown-linux-musl musl ;;
    libsidereon_linux_arm64_musl.a) printf '%s\t%s\n' aarch64-unknown-linux-musl musl ;;
    libsidereon_windows_amd64_gnu.a) printf '%s\t%s\n' x86_64-pc-windows-gnu gnu ;;
    *) die "unknown archive name: $1" ;;
  esac
}

verify_manifest_rows() {
  local manifest=$1
  local rows
  rows=$(awk -F '\t' '$0 !~ /^#/ && NF > 0 { count++ } END { print count + 0 }' "$manifest")
  [[ "$rows" == 7 ]] || die "manifest has $rows archive rows; expected 7"

  local name row hash size archive target libc ref commit actual_size actual_hash expected_target expected_libc
  while IFS= read -r name; do
    row=$(awk -F '\t' -v wanted="$name" '$0 !~ /^#/ && $3 == wanted { print; exit }' "$manifest")
    [[ -n "$row" ]] || die "manifest is missing $name"
    IFS=$'\t' read -r hash size archive target libc ref commit <<< "$row"
    [[ "$archive" == "$name" ]] || die "manifest archive column mismatch for $name"
    [[ "$hash" =~ ^[0-9a-f]{64}$ ]] || die "invalid SHA-256 for $name"
    [[ "$size" =~ ^[0-9]+$ && "$size" -ge 1024 ]] || die "invalid size for $name"
    [[ "$ref" == "$(manifest_meta source_ref)" ]] || die "source ref mismatch in $name"
    [[ "$commit" == "$(manifest_meta source_commit)" ]] || die "source commit mismatch in $name"
    IFS=$'\t' read -r expected_target expected_libc <<< "$(expected_archive_metadata "$name")"
    [[ "$target" == "$expected_target" && "$libc" == "$expected_libc" ]] || die "target metadata mismatch for $name"
    validate_archive "$LIB_DIR/$name"
    actual_size=$(file_size "$LIB_DIR/$name")
    actual_hash=$(sha256 "$LIB_DIR/$name")
    [[ "$actual_size" == "$size" ]] || die "size mismatch for $name: manifest $size, actual $actual_size"
    [[ "$actual_hash" == "$hash" ]] || die "SHA-256 mismatch for $name"
    note "verified $name ($actual_size bytes, $actual_hash)"
  done < <(expected_archive_names)

  local path base known
  for path in "$LIB_DIR"/*.a; do
    [[ -e "$path" ]] || continue
    base=${path##*/}
    known=$(expected_archive_names | awk -v wanted="$base" '$0 == wanted { print; exit }')
    [[ "$known" == "$base" ]] || die "unexpected archive in $LIB_DIR: $base"
  done
}

verify_version_rules() {
  local manifest=$1
  local ref version commit release
  ref=$(manifest_meta source_ref)
  version=$(manifest_meta header_version)
  commit=$(manifest_meta source_commit)
  release=$(manifest_meta release_target)
  [[ "$(manifest_meta manifest_version)" == 1 ]] || die "unsupported manifest version"
  [[ "$ref" =~ ^(v[0-9]+\.[0-9]+\.[0-9]+|[0-9a-f]{40})$ ]] || die "invalid source_ref in manifest: $ref"
  [[ "$commit" =~ ^[0-9a-f]{40}$ ]] || die "invalid source_commit in manifest: $commit"
  [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "invalid header_version in manifest: $version"
  [[ "$release" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "invalid release_target in manifest: $release"
  [[ "$(manifest_meta source_header_sha256)" =~ ^[0-9a-f]{64}$ ]] || die "invalid source header SHA-256"
  [[ "$ref" == "$SIDEREON_C_REF" ]] || die "manifest source_ref $ref disagrees with $C_REF_FILE"
  if [[ "$ref" =~ ^v([0-9]+\.[0-9]+\.[0-9]+)$ ]]; then
    [[ "$version" == "${BASH_REMATCH[1]}" ]] || die "tag $ref requires matching header macros; found $version"
    note "release tag $ref has matching header version $version"
  else
    note "pre-release source ref $ref has pinned header version $version; release target is $release"
  fi
}

verify_only() {
  require_command ar
  [[ -f "$LIB_DIR/manifest.sha256" ]] || die "missing $LIB_DIR/manifest.sha256"
  local manifest=$LIB_DIR/manifest.sha256
  verify_version_rules "$manifest"
  local vendored_header="$ROOT_DIR/internal/native/include/sidereon.h"
  [[ -f "$vendored_header" && ! -L "$vendored_header" ]] || die "missing or non-regular vendored header: $vendored_header"
  [[ "$(sha256 "$vendored_header")" == "$(manifest_meta source_header_sha256)" ]] || die "vendored header digest differs from manifest"
  [[ "$(header_version "$vendored_header")" == "$(manifest_meta header_version)" ]] || die "vendored header version differs from manifest"
  [[ "$(manifest_meta glibc_floor)" == "$GLIBC_FLOOR" ]] || die "glibc floor mismatch"
  verify_manifest_rows "$manifest"
  note "target completeness, archive format, hashes, and version/ref rules passed"
}

build_one() {
  local os=$1 arch=$2 libc=$3 target=$4 mode=$5 zig_target=$6 output_name=$7
  note "building $output_name for $target${zig_target:+ (zig target $zig_target)}"
  if [[ "$mode" == "zig" ]]; then
    "$CARGO" zigbuild --manifest-path "$SOURCE_DIR/bindings/c/Cargo.toml" --package sidereon-c --locked --release --target "$zig_target"
  else
    "$CARGO" build --manifest-path "$SOURCE_DIR/bindings/c/Cargo.toml" --package sidereon-c --locked --release --target "$target"
  fi

  local candidate
  for candidate in \
    "$TARGET_DIR/$target/release/libsidereon.a" \
    "$TARGET_DIR/$zig_target/release/libsidereon.a"; do
    if [[ -f "$candidate" ]]; then
      validate_archive "$candidate"
      cp "$candidate" "$STAGING_DIR/$output_name"
      chmod 0644 "$STAGING_DIR/$output_name"
      validate_archive "$STAGING_DIR/$output_name"
      return
    fi
  done
  die "static archive was not produced for $target"
}

build_all() {
  require_command awk
  require_command ar
  require_command cp
  require_command dd
  require_command grep
  require_command git
  require_command od
  require_command strings
  require_command tr
  require_hash_command
  [[ -x "$CARGO" ]] || die "usable Rust cargo not found at $CARGO"
  [[ -x "$RUSTC" ]] || die "usable Rust compiler not found at $RUSTC"
  [[ -x "$RUSTUP" ]] || die "rustup not found at $RUSTUP"
  [[ -x "$ZIG" ]] || die "Zig 0.15.2 not found at $ZIG"
  [[ "$($ZIG version)" == "0.15.2" ]] || die "unexpected Zig version at $ZIG"

  mkdir -p "$BUILD_ROOT" "$TARGET_DIR" "$LIB_DIR" "$GOPATH" "$GOMODCACHE" "$GOCACHE" "$ZIG_LOCAL_CACHE_DIR" "$ZIG_GLOBAL_CACHE_DIR" "$CARGO_ZIGBUILD_CACHE_DIR"
  SOURCE_TEMP_PARENT=""
  checkout_source
  trap 'if [[ -n "${SOURCE_TEMP_PARENT:-}" ]]; then rm -rf "$SOURCE_TEMP_PARENT"; fi' EXIT
  SOURCE_COMMIT=$(verify_source "$SOURCE_DIR")
  SOURCE_HEADER="$SOURCE_DIR/bindings/c/include/sidereon.h"
  [[ -f "$SOURCE_HEADER" ]] || die "pinned source header is missing: $SOURCE_HEADER"
  HEADER_VERSION=$(header_version "$SOURCE_HEADER")
  HEADER_SHA256=$(sha256 "$SOURCE_HEADER")

  VENDORED_HEADER="${SIDEREON_VENDORED_HEADER:-$ROOT_DIR/internal/native/include/sidereon.h}"
  if [[ -e "$VENDORED_HEADER" ]]; then
    [[ -f "$VENDORED_HEADER" && ! -L "$VENDORED_HEADER" ]] || die "vendored header is not a regular file: $VENDORED_HEADER"
    cmp -s "$SOURCE_HEADER" "$VENDORED_HEADER" || die "vendored header differs from pinned source header"
    note "vendored header matches pinned source exactly"
  else
    note "no Go vendored header is present; checked the pinned source header at $HEADER_VERSION"
  fi

  if [[ "$SIDEREON_C_REF" =~ ^v([0-9]+\.[0-9]+\.[0-9]+)$ ]]; then
    [[ "$HEADER_VERSION" == "${BASH_REMATCH[1]}" ]] || die "tag $SIDEREON_C_REF requires matching header macros; found $HEADER_VERSION"
    note "release tag $SIDEREON_C_REF has matching header version $HEADER_VERSION"
  else
    note "pre-release source ref $SIDEREON_C_REF reports header version $HEADER_VERSION; release target remains $RELEASE_VERSION"
  fi

  # Keep source and registry paths out of archives so two clean builds do not
  # record workstation-specific locations in Rust panic/debug strings.
  local remap_flags
  remap_flags="--remap-path-prefix=$SOURCE_DIR=__sidereon_c__ --remap-path-prefix=$CARGO_HOME=__cargo_home__ --remap-path-prefix=$ROOT_DIR=__sidereon_go__"
  export RUSTFLAGS="${RUSTFLAGS:+$RUSTFLAGS }$remap_flags"

  if ! command -v cargo-zigbuild >/dev/null 2>&1; then
    note "installing cargo-zigbuild $CARGO_ZIGBUILD_VERSION into $CARGO_HOME"
    "$CARGO" install --locked --version "$CARGO_ZIGBUILD_VERSION" cargo-zigbuild
  fi
  require_command cargo-zigbuild
  local cargo_zigbuild_version
  cargo_zigbuild_version=$(cargo-zigbuild --version)
  [[ "$cargo_zigbuild_version" == *"$CARGO_ZIGBUILD_VERSION"* ]] || die "unexpected cargo-zigbuild version: $cargo_zigbuild_version"

  local installed target
  for target in \
    aarch64-apple-darwin \
    x86_64-apple-darwin \
    x86_64-unknown-linux-gnu \
    aarch64-unknown-linux-gnu \
    x86_64-unknown-linux-musl \
    aarch64-unknown-linux-musl \
    x86_64-pc-windows-gnu; do
    installed=$($RUSTUP target list --installed | awk -v wanted="$target" '$1 == wanted { print; exit }')
    if [[ "$installed" != "$target" ]]; then
      note "installing Rust target $target"
      "$RUSTUP" target add "$target"
    fi
  done

  STAGING_DIR=$(mktemp -d "$BUILD_ROOT/staging.XXXXXX")
  trap 'if [[ -n "${SOURCE_TEMP_PARENT:-}" ]]; then rm -rf "$SOURCE_TEMP_PARENT"; fi; if [[ -n "${STAGING_DIR:-}" ]]; then rm -rf "$STAGING_DIR"; fi' EXIT

  # os|arch|libc|Rust target|build mode|cargo-zigbuild target|archive name
  local row os arch libc rust_target build_mode zig_target archive_name
  while IFS='|' read -r os arch libc rust_target build_mode zig_target archive_name; do
    if [[ "$libc" == "glibc" ]]; then
      zig_target="$rust_target.$GLIBC_FLOOR"
    fi
    build_one "$os" "$arch" "$libc" "$rust_target" "$build_mode" "$zig_target" "$archive_name"
  done <<'TARGETS'
darwin|arm64|none|aarch64-apple-darwin|cargo||libsidereon_darwin_arm64.a
darwin|amd64|none|x86_64-apple-darwin|cargo||libsidereon_darwin_amd64.a
linux|amd64|glibc|x86_64-unknown-linux-gnu|zig||libsidereon_linux_amd64_glibc.a
linux|arm64|glibc|aarch64-unknown-linux-gnu|zig||libsidereon_linux_arm64_glibc.a
linux|amd64|musl|x86_64-unknown-linux-musl|zig|x86_64-unknown-linux-musl|libsidereon_linux_amd64_musl.a
linux|arm64|musl|aarch64-unknown-linux-musl|zig|aarch64-unknown-linux-musl|libsidereon_linux_arm64_musl.a
windows|amd64|gnu|x86_64-pc-windows-gnu|zig|x86_64-pc-windows-gnu|libsidereon_windows_amd64_gnu.a
TARGETS

  local cargo_version rustc_version zig_version
  cargo_version=$("$CARGO" --version)
  rustc_version=$("$RUSTC" --version)
  zig_version=$("$ZIG" version)
  {
    printf '# sidereon-go static archive manifest\n'
    printf '# manifest_version=1\n'
    printf '# release_target=%s\n' "$RELEASE_VERSION"
    printf '# source_ref=%s\n' "$SIDEREON_C_REF"
    printf '# source_commit=%s\n' "$SOURCE_COMMIT"
    printf '# header_version=%s\n' "$HEADER_VERSION"
    printf '# source_header_sha256=%s\n' "$HEADER_SHA256"
    printf '# glibc_floor=%s\n' "$GLIBC_FLOOR"
    printf '# cargo=%s\n' "$cargo_version"
    printf '# rustc=%s\n' "$rustc_version"
    printf '# cargo_zigbuild=%s\n' "$cargo_zigbuild_version"
    printf '# zig=%s\n' "$zig_version"
    printf '# columns=sha256 bytes archive rust_target libc source_ref source_commit\n'
    while IFS='|' read -r os arch libc rust_target build_mode zig_target archive_name; do
      local archive_hash archive_size
      archive_hash=$(sha256 "$STAGING_DIR/$archive_name")
      archive_size=$(file_size "$STAGING_DIR/$archive_name")
      printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
        "$archive_hash" "$archive_size" "$archive_name" "$rust_target" "$libc" "$SIDEREON_C_REF" "$SOURCE_COMMIT"
    done <<'TARGETS'
darwin|arm64|none|aarch64-apple-darwin|cargo||libsidereon_darwin_arm64.a
darwin|amd64|none|x86_64-apple-darwin|cargo||libsidereon_darwin_amd64.a
linux|amd64|glibc|x86_64-unknown-linux-gnu|zig||libsidereon_linux_amd64_glibc.a
linux|arm64|glibc|aarch64-unknown-linux-gnu|zig||libsidereon_linux_arm64_glibc.a
linux|amd64|musl|x86_64-unknown-linux-musl|zig|x86_64-unknown-linux-musl|libsidereon_linux_amd64_musl.a
linux|arm64|musl|aarch64-unknown-linux-musl|zig|aarch64-unknown-linux-musl|libsidereon_linux_arm64_musl.a
windows|amd64|gnu|x86_64-pc-windows-gnu|zig|x86_64-pc-windows-gnu|libsidereon_windows_amd64_gnu.a
TARGETS
  } > "$STAGING_DIR/manifest.sha256"

  # Check the staging paths before the final copies. The manifest verifier is
  # parameterized by LIB_DIR, so temporarily point it at the staging set.
  local final_lib_dir=$LIB_DIR
  LIB_DIR=$STAGING_DIR
  verify_manifest_rows "$STAGING_DIR/manifest.sha256"
  LIB_DIR=$final_lib_dir
  local name
  while IFS= read -r name; do
    cp "$STAGING_DIR/$name" "$LIB_DIR/$name"
    chmod 0644 "$LIB_DIR/$name"
  done < <(expected_archive_names)
  cp "$STAGING_DIR/manifest.sha256" "$LIB_DIR/manifest.sha256"
  chmod 0644 "$LIB_DIR/manifest.sha256"
  verify_only
  note "wrote seven archives and manifest under $LIB_DIR"
}

if [[ "$MODE" == "verify" ]]; then
  require_command awk
  require_command ar
  require_command dd
  require_command grep
  require_command od
  require_command strings
  require_command tr
  require_hash_command
  verify_only
else
  build_all
fi
