package sidereon

import (
	"reflect"
	"testing"
)

type sbasTestBitField struct {
	width int
	value uint64
}

func encodeSBAS226ForTest(messageType uint8, fields ...sbasTestBitField) []byte {
	data := make([]byte, 29)
	offset := 0
	write := func(width int, value uint64) {
		for bit := 0; bit < width; bit++ {
			if value&(uint64(1)<<uint(width-bit-1)) != 0 {
				position := offset + bit
				data[position/8] |= byte(1 << uint(7-position%8))
			}
		}
		offset += width
	}
	write(8, 0x53)
	write(6, uint64(messageType))
	for _, field := range fields {
		write(field.width, field.value)
	}
	if offset > 226 {
		panic("test SBAS body exceeds 226 bits")
	}
	return data
}

func TestSBASUnassignedMaskCorrectionsRetainTypedRows(t *testing.T) {
	maskFields := []sbasTestBitField{}
	mask := make([]bool, 210)
	mask[0], mask[70], mask[119] = true, true, true
	for _, active := range mask {
		value := uint64(0)
		if active {
			value = 1
		}
		maskFields = append(maskFields, sbasTestBitField{width: 1, value: value})
	}
	maskFields = append(maskFields, sbasTestBitField{width: 2, value: 1})
	maskBlock, err := DecodeSBASBlock(encodeSBAS226ForTest(1, maskFields...), SBASWireBody226)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, maskBlock)
	if info, err := maskBlock.Info(); err != nil || info.Kind != SBASMessagePRNMask || info.MessageType != 1 {
		t.Fatalf("mask block info=%+v err=%v", info, err)
	}
	fastFields := []sbasTestBitField{{width: 2, value: 0}, {width: 2, value: 1}}
	for _, prc := range []uint64{8, 16, 24} {
		fastFields = append(fastFields, sbasTestBitField{width: 12, value: prc})
	}
	for i := 3; i < 13; i++ {
		fastFields = append(fastFields, sbasTestBitField{width: 12})
	}
	for i := 0; i < 13; i++ {
		fastFields = append(fastFields, sbasTestBitField{width: 4})
	}
	fastBlock, err := DecodeSBASBlock(encodeSBAS226ForTest(2, fastFields...), SBASWireBody226)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, fastBlock)
	if info, err := fastBlock.Info(); err != nil || info.Kind != SBASMessageFastCorrections || info.MessageType != 2 {
		t.Fatalf("fast block info=%+v err=%v", info, err)
	}
	store, err := NewSBASCorrectionStore()
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, store)
	epoch := GNSSWeekTow{System: GPST, Week: 2400, TOWSeconds: 10}
	if err := store.Ingest(maskBlock, "S20", epoch); err != nil {
		t.Fatal(err)
	}
	epoch.TOWSeconds = 20
	if err := store.Ingest(fastBlock, "S20", epoch); err != nil {
		t.Fatal(err)
	}
	rows, present, err := store.UnassignedMaskCorrections("S20")
	if correction, found, correctionErr := store.FastCorrection("S20", "G01"); correctionErr != nil || !found {
		t.Fatalf("decoded first fast correction=%+v present=%v err=%v", correction, found, correctionErr)
	}
	want := []SBASUnassignedMaskCorrection{{MaskNumber: 71, Count: 1}}
	if err != nil || !present || !reflect.DeepEqual(rows, want) {
		t.Fatalf("unassigned SBAS mask rows = %+v, present=%v, err=%v; want %+v", rows, present, err, want)
	}
	readDone := make(chan error, 1)
	go func() {
		for i := 0; i < 40; i++ {
			if _, _, readErr := store.UnassignedMaskCorrections("S20"); readErr != nil {
				readDone <- readErr
				return
			}
			if _, _, readErr := store.FastCorrection("S20", "G01"); readErr != nil {
				readDone <- readErr
				return
			}
		}
		readDone <- nil
	}()
	for i := 0; i < 8; i++ {
		epoch.TOWSeconds = float64(30 + i)
		if ingestErr := store.Ingest(fastBlock, "S20", epoch); ingestErr != nil {
			t.Fatalf("concurrent SBAS ingest %d: %v", i, ingestErr)
		}
	}
	if err := <-readDone; err != nil {
		t.Fatalf("concurrent SBAS reader: %v", err)
	}
}
