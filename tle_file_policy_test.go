package sidereon

import (
	"bytes"
	"strings"
	"testing"
)

func TestTLEFilePolicyAndRejectedRecordRetention(t *testing.T) {
	bad := append([]byte(nil), readPositioningFixture(t, "iss.tle")...)
	checksum := -1
	for i := 0; i+1 < len(bad); i++ {
		if bad[i] == '1' && bad[i+1] == ' ' {
			for j := i; j < len(bad) && bad[j] != '\n'; j++ {
				if j-i == 68 && bad[j] >= '0' && bad[j] <= '9' {
					checksum = j
					break
				}
			}
			break
		}
	}
	if checksum < 0 {
		t.Fatal("fixture line 1 has no checksum digit")
	}
	bad[checksum] = '0' + (bad[checksum]-'0'+1)%10

	strict, err := ParseTLEFileWithPolicy(bad, TLEFilePolicyStrict)
	if err != nil {
		t.Fatalf("strict parse: %v", err)
	}
	closeAfterTest(t, strict)
	if count, err := strict.Count(); err != nil || count != 0 {
		t.Fatalf("strict accepted=%d err=%v", count, err)
	}
	if skipped, err := strict.Skipped(); err != nil || skipped != 1 {
		t.Fatalf("strict skipped=%d err=%v", skipped, err)
	}
	rejected, err := strict.Rejected(0)
	if err != nil || rejected.LineNumber != 1 || rejected.Issue != TLERecordIssueInvalid || rejected.Name != "" || rejected.Error == "" {
		t.Fatalf("strict rejected record=%+v err=%v", rejected, err)
	}

	lenient, err := ParseTLEFileWithPolicy(bad, TLEFilePolicyLenient)
	if err != nil {
		t.Fatalf("lenient parse: %v", err)
	}
	closeAfterTest(t, lenient)
	if count, err := lenient.Count(); err != nil || count != 1 {
		t.Fatalf("lenient accepted=%d err=%v", count, err)
	}
	if skipped, err := lenient.Skipped(); err != nil || skipped != 0 {
		t.Fatalf("lenient skipped=%d err=%v", skipped, err)
	}
	satellite, err := lenient.Satellite(0)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, satellite)
	warnings, err := satellite.ChecksumWarnings()
	if err != nil || len(warnings) != 1 || warnings[0].LineNumber != 1 {
		t.Fatalf("lenient checksum warnings=%+v err=%v", warnings, err)
	}
	if _, err := strict.Rejected(-1); err == nil {
		t.Fatal("negative rejected index accepted")
	}
}

func TestTLELoadPolicyAndAcceptedRecordLineNumber(t *testing.T) {
	bad := append([]byte(nil), readPositioningFixture(t, "iss.tle")...)
	line1Start, checksum := -1, -1
	for i := 0; i+1 < len(bad); i++ {
		if bad[i] != '1' || bad[i+1] != ' ' {
			continue
		}
		line1Start = i
		for j := i; j < len(bad) && bad[j] != '\n'; j++ {
			if j-i == 68 && bad[j] >= '0' && bad[j] <= '9' {
				checksum = j
				break
			}
		}
		break
	}
	if line1Start < 0 || checksum < 0 {
		t.Fatal("fixture line 1 has no checksum digit")
	}
	wantLineNumber := bytes.Count(bad[:line1Start], []byte{'\n'}) + 1
	bad[checksum] = '0' + (bad[checksum]-'0'+1)%10

	lines := strings.Split(strings.TrimRight(string(bad), "\r\n"), "\n")
	var line1, line2 string
	for i, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if strings.HasPrefix(line, "1 ") && i+1 < len(lines) {
			line1 = line
			line2 = strings.TrimSuffix(lines[i+1], "\r")
			break
		}
	}
	if line1 == "" || !strings.HasPrefix(line2, "2 ") {
		t.Fatalf("could not isolate line pair from fixture: %q / %q", line1, line2)
	}
	if _, err := ParseTLEWithOpsModeAndPolicy(line1, line2, OpsModeAFSPC, TLEFilePolicyStrict); err == nil {
		t.Fatal("strict pair parser accepted a checksum mismatch")
	}
	parsed, err := ParseTLEWithOpsModeAndPolicy(line1, line2, OpsModeAFSPC, TLEFilePolicyLenient)
	if err != nil {
		t.Fatalf("lenient pair parser rejected checksum mismatch: %v", err)
	}
	closeAfterTest(t, parsed)
	warnings, err := parsed.ChecksumWarnings()
	if err != nil || len(warnings) != 1 || warnings[0].LineNumber != 1 || warnings[0].Found == warnings[0].Computed {
		t.Fatalf("lenient pair checksum warning=%+v err=%v", warnings, err)
	}

	file, err := ParseTLEFileWithPolicy(bad, TLEFilePolicyLenient)
	if err != nil {
		t.Fatalf("lenient file parser: %v", err)
	}
	closeAfterTest(t, file)
	lineNumber, err := file.LineNumber(0)
	if err != nil || lineNumber != wantLineNumber {
		t.Fatalf("accepted record line number=%d, want %d: %v", lineNumber, wantLineNumber, err)
	}
	if _, err := file.LineNumber(-1); err == nil {
		t.Fatal("negative accepted-record index was accepted")
	}
}
