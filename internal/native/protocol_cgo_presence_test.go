//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "testing"

func TestBroadcastRecordOptionalPresenceCConversion(t *testing.T) {
	for _, test := range []struct {
		name                      string
		record                    NativeBroadcastRecord
		wantIssue, wantSVAccuracy bool
	}{
		{
			name: "CNAV absent zero fields",
			record: NativeBroadcastRecord{
				SatelliteID: "G01", Message: 1,
				CNAV: NativeBroadcastCNAV{Present: true},
			},
		},
		{
			name:      "present zero values",
			record:    NativeBroadcastRecord{HasIssue: true, HasSVAccuracyM: true},
			wantIssue: true, wantSVAccuracy: true,
		},
		{
			name:      "legacy nonzero values imply presence",
			record:    NativeBroadcastRecord{Issue: 1, SVAccuracyM: 2},
			wantIssue: true, wantSVAccuracy: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			test.record.SatelliteID = "G01"
			cRecord, err := broadcastRecordToC(test.record)
			if err != nil {
				t.Fatalf("convert broadcast record to C: %v", err)
			}
			got := broadcastRecordFromC(cRecord)
			if got.HasIssue != test.wantIssue || got.Issue != test.record.Issue || got.HasSVAccuracyM != test.wantSVAccuracy || got.SVAccuracyM != test.record.SVAccuracyM {
				t.Fatalf("C record round trip optional fields = (%t,%d,%t,%g), want (%t,%d,%t,%g)", got.HasIssue, got.Issue, got.HasSVAccuracyM, got.SVAccuracyM, test.wantIssue, test.record.Issue, test.wantSVAccuracy, test.record.SVAccuracyM)
			}
		})
	}
}
