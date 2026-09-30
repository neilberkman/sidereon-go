package sidereon

import "testing"

type testCloseable interface {
	Close() error
}

func closeAfterTest(t *testing.T, value testCloseable) {
	t.Helper()
	t.Cleanup(func() {
		if err := value.Close(); err != nil {
			t.Errorf("close test resource %T: %v", value, err)
		}
	})
}
