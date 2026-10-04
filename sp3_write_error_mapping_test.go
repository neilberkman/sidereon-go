package sidereon

import (
	"math"
	"reflect"
	"testing"

	"sidereon.dev/go/v3/internal/native"
)

type sp3WriteErrorSetter func(*native.SP3WriteError, *SP3WriteError)

func sp3Field(value string) sp3WriteErrorSetter {
	return func(in *native.SP3WriteError, want *SP3WriteError) {
		in.HasField, in.Field = true, value
		want.HasField, want.Field = true, value
	}
}

func sp3Text(value string) sp3WriteErrorSetter {
	return func(in *native.SP3WriteError, want *SP3WriteError) {
		in.HasTextValue, in.TextValue = true, value
		want.HasTextValue, want.TextValue = true, value
	}
}

func sp3Satellite(value string) sp3WriteErrorSetter {
	return func(in *native.SP3WriteError, want *SP3WriteError) {
		in.HasSatelliteID, in.SatelliteID = true, value
		want.HasSatelliteID, want.SatelliteID = true, value
	}
}

func sp3Epoch(value int) sp3WriteErrorSetter {
	return func(in *native.SP3WriteError, want *SP3WriteError) {
		in.HasEpochIndex, in.EpochIndex = true, value
		want.HasEpochIndex, want.EpochIndex = true, value
	}
}

func sp3Columns(value int) sp3WriteErrorSetter {
	return func(in *native.SP3WriteError, want *SP3WriteError) {
		in.HasColumns, in.Columns = true, value
		want.HasColumns, want.Columns = true, value
	}
}

func sp3Decimals(value int) sp3WriteErrorSetter {
	return func(in *native.SP3WriteError, want *SP3WriteError) {
		in.HasDecimals, in.Decimals = true, value
		want.HasDecimals, want.Decimals = true, value
	}
}

func applySP3Setters(in *native.SP3WriteError, want *SP3WriteError, setters ...sp3WriteErrorSetter) {
	for _, set := range setters {
		set(in, want)
	}
}

func TestPublicSP3WriteOutcomePreservesEveryRefusalPayload(t *testing.T) {
	type testCase struct {
		name string
		kind SP3WriteErrorKind
		set  sp3WriteErrorSetter
	}
	cases := []testCase{
		{"TextNotColumnSafe", SP3WriteErrorTextNotColumnSafe, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("agency"), sp3Text("A\nB"))
		}},
		{"TextNotColumnStable", SP3WriteErrorTextNotColumnStable, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("orbit type"), sp3Text(" FIT "))
		}},
		{"BlankDescriptor", SP3WriteErrorBlankDescriptor, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("data used"), sp3Text(""))
		}},
		{"EmptyComment", SP3WriteErrorEmptyComment, func(in *native.SP3WriteError, want *SP3WriteError) {
			in.HasCommentIndex, in.CommentIndex, in.HasTextValue, in.TextValue = true, 3, true, ""
			want.HasCommentIndex, want.CommentIndex, want.HasTextValue, want.TextValue = true, 3, true, ""
		}},
		{"TextTooWide", SP3WriteErrorTextTooWide, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("agency"), sp3Columns(4), sp3Text("ABCDE"))
		}},
		{"IntegerTooWide", SP3WriteErrorIntegerTooWide, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("epoch count"), sp3Columns(7))
			in.HasIntegerValue, in.IntegerValue = true, math.MaxUint64
			want.HasIntegerValue, want.IntegerValue = true, math.MaxUint64
		}},
		{"NonFinite", SP3WriteErrorNonFinite, func(in *native.SP3WriteError, want *SP3WriteError) { applySP3Setters(in, want, sp3Field("interval")) }},
		{"NumberTooWide", SP3WriteErrorNumberTooWide, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("clock base"), sp3Columns(10), sp3Decimals(7))
			in.HasNumber, in.Number = true, 12345.25
			want.HasNumber, want.Number = true, 12345.25
		}},
		{"PrecisionNotRepresentable", SP3WriteErrorPrecisionNotRepresentable, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("position base"), sp3Columns(10), sp3Decimals(7))
			in.HasNumber, in.Number = true, 1.25000001
			want.HasNumber, want.Number = true, 1.25000001
		}},
		{"AccuracyNotRepresentable", SP3WriteErrorAccuracyNotRepresentable, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Satellite("G07"), sp3Epoch(5), sp3Field("position"))
			in.HasExponent, in.Exponent = true, -12
			want.HasExponent, want.Exponent = true, -12
		}},
		{"AccuracyRecordMismatch", SP3WriteErrorAccuracyRecordMismatch, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Satellite("G07"), sp3Epoch(6))
		}},
		{"AccuracyBasisMissing", SP3WriteErrorAccuracyBasisMissing, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Satellite("G07"), sp3Epoch(8))
		}},
		{"YearNotRepresentable", SP3WriteErrorYearNotRepresentable, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Epoch(9))
			in.HasYear, in.Year = true, -12345
			want.HasYear, want.Year = true, -12345
		}},
		{"EpochNotRestatable", SP3WriteErrorEpochNotRestatable, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Epoch(10))
			in.HasFieldSeconds, in.FieldSeconds, in.HasResidualS, in.ResidualS = true, 59.125, true, 0.00000001
			want.HasFieldSeconds, want.FieldSeconds, want.HasResidualS, want.ResidualS = true, 59.125, true, 0.00000001
		}},
		{"EpochTimeScaleMismatch", SP3WriteErrorEpochTimeScaleMismatch, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Epoch(11))
			in.HasEpochTimeScale, in.EpochTimeScale, in.HasHeaderTimeScale, in.HeaderTimeScale = true, uint32(GPST), true, uint32(UTC)
			want.HasEpochTimeScale, want.EpochTimeScale, want.HasHeaderTimeScale, want.HeaderTimeScale = true, GPST, true, UTC
		}},
		{"HeaderTimeScaleMismatch", SP3WriteErrorHeaderTimeScaleMismatch, func(in *native.SP3WriteError, want *SP3WriteError) {
			in.HasTimeSystem, in.TimeSystem, in.HasHeaderTimeScale, in.HeaderTimeScale = true, "GAL", true, uint32(GPST)
			want.HasTimeSystem, want.TimeSystem, want.HasHeaderTimeScale, want.HeaderTimeScale = true, "GAL", true, GPST
		}},
		{"EpochCountMismatch", SP3WriteErrorEpochCountMismatch, func(in *native.SP3WriteError, want *SP3WriteError) {
			in.HasDeclaredEpochs, in.DeclaredEpochs, in.HasEpochs, in.Epochs = true, math.MaxUint64, true, 12
			want.HasDeclaredEpochs, want.DeclaredEpochs, want.HasEpochs, want.Epochs = true, math.MaxUint64, true, 12
		}},
		{"AccuracyCodeCountMismatch", SP3WriteErrorAccuracyCodeCountMismatch, func(in *native.SP3WriteError, want *SP3WriteError) {
			in.HasSatellites, in.Satellites, in.HasCodes, in.Codes = true, 13, true, 12
			want.HasSatellites, want.Satellites, want.HasCodes, want.Codes = true, 13, true, 12
		}},
		{"DuplicateSatellite", SP3WriteErrorDuplicateSatellite, func(in *native.SP3WriteError, want *SP3WriteError) { applySP3Setters(in, want, sp3Satellite("G07")) }},
		{"SatelliteNotRepresentable", SP3WriteErrorSatelliteNotRepresentable, func(in *native.SP3WriteError, want *SP3WriteError) { applySP3Setters(in, want, sp3Satellite("G100")) }},
		{"EpochArrayLengthMismatch", SP3WriteErrorEpochArrayLengthMismatch, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("clocks"))
			in.HasEpochs, in.Epochs, in.HasEntries, in.Entries = true, 14, true, 13
			want.HasEpochs, want.Epochs, want.HasEntries, want.Entries = true, 14, true, 13
		}},
		{"UndeclaredSatelliteRecord", SP3WriteErrorUndeclaredSatelliteRecord, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Satellite("G07"), sp3Epoch(15))
		}},
		{"ConflictingRecords", SP3WriteErrorConflictingRecords, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Satellite("G07"), sp3Epoch(16))
		}},
		{"VelocityInPositionProduct", SP3WriteErrorVelocityInPositionProduct, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("velocity x"), sp3Satellite("G07"), sp3Epoch(17))
		}},
		{"RecordValueNonFinite", SP3WriteErrorRecordValueNonFinite, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("clock"), sp3Satellite("G07"), sp3Epoch(18))
		}},
		{"RecordValueTooWide", SP3WriteErrorRecordValueTooWide, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("position x"), sp3Satellite("G07"), sp3Epoch(19), sp3Columns(14), sp3Decimals(6))
			in.HasColumnValue, in.ColumnValue = true, 123456789.25
			want.HasColumnValue, want.ColumnValue = true, 123456789.25
		}},
		{"RecordValueNotRepresentable", SP3WriteErrorRecordValueNotRepresentable, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("clock"), sp3Satellite("G07"), sp3Epoch(20), sp3Columns(14), sp3Decimals(6))
			in.HasStored, in.Stored, in.HasColumnValue, in.ColumnValue = true, 0.00000125, true, 1.25000001
			want.HasStored, want.Stored, want.HasColumnValue, want.ColumnValue = true, 0.00000125, true, 1.25000001
		}},
		{"RecordReadsAsAbsent", SP3WriteErrorRecordReadsAsAbsent, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("clock"), sp3Satellite("G07"), sp3Epoch(21))
			in.HasColumnValue, in.ColumnValue = true, 999999.999999
			want.HasColumnValue, want.ColumnValue = true, 999999.999999
		}},
		{"RecordFieldsDisagree", SP3WriteErrorRecordFieldsDisagree, func(in *native.SP3WriteError, want *SP3WriteError) {
			applySP3Setters(in, want, sp3Field("velocity y"), sp3Satellite("G07"), sp3Epoch(22))
			in.HasStored, in.Stored, in.HasNative, in.Native = true, math.Copysign(0, -1), true, -2.5
			want.HasStored, want.Stored, want.HasNative, want.Native = true, math.Copysign(0, -1), true, -2.5
		}},
	}

	if len(cases) != 29 {
		t.Fatalf("case count = %d", len(cases))
	}
	for index, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			message := "SP3 refusal " + tc.name
			in := native.SP3WriteError{Kind: uint32(tc.kind), Message: message}
			want := SP3WriteError{Kind: tc.kind, Message: message}
			tc.set(&in, &want)
			outcome := publicSP3WriteOutcome(native.SP3WriteOutcome{IsOK: false, Status: uint32(100 + index), Error: in})
			if outcome.IsOK || outcome.Status != uint32(100+index) {
				t.Fatalf("outcome envelope = %+v", outcome)
			}
			if !reflect.DeepEqual(outcome.Error, want) {
				t.Fatalf("mapped error\n got: %#v\nwant: %#v", outcome.Error, want)
			}
			if tc.kind == SP3WriteErrorRecordFieldsDisagree && math.Float64bits(outcome.Error.Stored) != math.Float64bits(math.Copysign(0, -1)) {
				t.Fatalf("stored signed zero bits = %016x", math.Float64bits(outcome.Error.Stored))
			}
		})
	}
}
