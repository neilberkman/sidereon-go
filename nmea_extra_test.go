package sidereon

import (
	"errors"
	"math"
	"testing"
)

func TestNMEAAccumulatorChunkingFinishAndCopies(t *testing.T) {
	accumulator, err := NewNMEAAccumulator()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := accumulator.Close(); err != nil {
			t.Error(err)
		}
	})

	data := []byte(nmeaFixture)
	first, err := accumulator.Push(data[:18])
	if err != nil {
		t.Fatal(err)
	}
	if first.SentenceCount != 0 || first.CompletedEpochCount != 0 || first.RetainedLength != 18 {
		t.Fatalf("first chunk = %+v", first)
	}
	for i := range data[:18] {
		data[i] = 'X'
	}
	second, err := accumulator.Push(data[18:])
	if err != nil {
		t.Fatal(err)
	}
	if second.SentenceCount != 2 || second.CompletedEpochCount != 1 || second.RetainedLength != 0 {
		t.Fatalf("second chunk = %+v", second)
	}
	retained, err := accumulator.RetainedLength()
	if err != nil || retained != 0 {
		t.Fatalf("retained length = %d, %v", retained, err)
	}
	finished, err := accumulator.Finish()
	if err != nil {
		t.Fatal(err)
	}
	if finished.CompletedEpochCount != 1 || finished.RetainedLength != 0 {
		t.Fatalf("finish = %+v", finished)
	}
	summary, err := accumulator.Summary()
	if err != nil {
		t.Fatal(err)
	}
	if summary.SentenceCount != 2 || summary.EpochCount != 2 || summary.SkipCount != 0 || summary.WarningCount != 0 {
		t.Fatalf("summary = %+v", summary)
	}
	epochs, err := accumulator.Epochs()
	if err != nil {
		t.Fatal(err)
	}
	if len(epochs) != 2 || !epochs[0].HasGGA || !epochs[0].HasPosition {
		t.Fatalf("epochs = %+v", epochs)
	}
}

func TestWriteNMEAGGA(t *testing.T) {
	options := DefaultNMEAGGAOptions()
	options.UTCSecondsOfDay = 3661.239
	options.Position = Geodetic{LatitudeRad: 40 * math.Pi / 180, LongitudeRad: -105 * math.Pi / 180, HeightM: 1600}
	value, err := WriteNMEAGGA(options)
	if err != nil {
		t.Fatal(err)
	}
	const expected = "$GPGGA,010101.23,4000.0000000,N,10500.0000000,W,1,10,1.00,1600.0,M,0.0,M,,*49\r\n"
	if string(value) != expected {
		t.Fatalf("GGA = %q, want %q", value, expected)
	}
	options.Talker = "G"
	if _, err := WriteNMEAGGA(options); err == nil {
		t.Fatal("one-byte talker was accepted")
	}
}

func TestNMEAAccumulatorCloseIsIdempotent(t *testing.T) {
	accumulator, err := NewNMEAAccumulator()
	if err != nil {
		t.Fatal(err)
	}
	if err := accumulator.Close(); err != nil {
		t.Fatal(err)
	}
	if err := accumulator.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := accumulator.Summary(); !errors.Is(err, ErrClosed) {
		t.Fatalf("summary after close = %v, want ErrClosed", err)
	}
}
