//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "testing"

func TestUnsupportedBodyPropagatesNativeIndexFailure(t *testing.T) {
	messages, err := BuildRTCMUnsupported(4090, []byte{0x12, 0x34})
	if err != nil {
		t.Fatal(err)
	}
	closeNativeAfterTest(t, messages)

	if body, err := messages.UnsupportedBody(1); err == nil {
		t.Fatalf("positive out-of-range index returned body %x without error", body)
	}
}
