#!/usr/bin/env bash
# Build a developer archive from a C source checkout. Release artifacts use
# build-native-archives.sh and its pinned source requirements.
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)
source_dir=${SIDEREON_C_SOURCE:?Set SIDEREON_C_SOURCE to the C source checkout}
build_dir=${NATIVE_BUILD_ROOT:-$root/.native-build/developer-darwin-arm64}
export RUSTUP_TOOLCHAIN=1.98.1
export CARGO_TARGET_DIR="$build_dir"
export CARGO_PROFILE_RELEASE_LTO=off
unset CARGO_ENCODED_RUSTFLAGS
# Rust's distributed standard library itself carries embedded bitcode. Build
# it from the matching rust-src component so the final static archive contains
# machine code only, without modifying or relocating other target archives.
export RUSTC_BOOTSTRAP=1
export RUSTFLAGS="-C embed-bitcode=no -C lto=off --remap-path-prefix=$source_dir=__sidereon_c__ --remap-path-prefix=$root=__sidereon_go__"
mkdir -p "$build_dir"
git -C "$source_dir" rev-parse HEAD > "$build_dir/source-head.txt"
git -C "$source_dir" status --short > "$build_dir/source-status.txt"
git -C "$source_dir" diff HEAD --binary | shasum -a 256 > "$build_dir/source-diff.sha256"
cargo build -Z build-std=std,panic_unwind --release --target aarch64-apple-darwin --manifest-path "$source_dir/bindings/c/Cargo.toml"
archive="$build_dir/aarch64-apple-darwin/release/libsidereon.a"
xcrun otool -l "$archive" > "$build_dir/archive-load-commands.txt"
if rg -q '__LLVM|__bitcode|__GNU_LTO|\.llvmbc' "$build_dir/archive-load-commands.txt"; then
  echo "native archive contains embedded bitcode/LTO" >&2
  exit 1
fi
cp "$archive" "$root/internal/native/lib/libsidereon_darwin_arm64.a"
cp "$source_dir/bindings/c/include/sidereon.h" "$root/internal/native/include/sidereon.h"
echo "Built darwin arm64 from C HEAD $(cat "$build_dir/source-head.txt")"
cat "$build_dir/source-status.txt"
