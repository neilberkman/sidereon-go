package sidereon

import (
	"reflect"
	"testing"
)

func TestCatalogValueFieldEquality(t *testing.T) {
	values := []any{CatalogSourceEntry{}, CelestrakGPPayload{}, CelestrakTLERecord{}, CelestrakChecksumWarning{}, CelestrakChecksumDetail{}, CelestrakGPRecordData{}, CelestrakGPRejectionReason{}, CelestrakTLEError{}, CelestrakTLEErrorFields{}, CelestrakElementSetError{}, CelestrakElementSetErrorFields{}, CelestrakGPRejectedRecordData{}, DataCatalogError{}, CelestrakGPRecordError{}, OMMRecordSpan{}, SGP4FitOMMHeader{}, SGP4FitOMMHeaderError{}, OMMParseError{}, OMMParseErrorFields{}, OMMParseExpected{}}
	var setNonzero func(reflect.Value)
	setNonzero = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.String:
			v.SetString("different")
		case reflect.Bool:
			v.SetBool(true)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v.SetInt(1)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v.SetUint(1)
		case reflect.Float32, reflect.Float64:
			v.SetFloat(1)
		case reflect.Pointer:
			v.Set(reflect.New(v.Type().Elem()))
		case reflect.Slice:
			v.Set(reflect.MakeSlice(v.Type(), 1, 1))
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if v.Field(i).CanSet() {
					setNonzero(v.Field(i))
					return
				}
			}
		default:
			t.Fatalf("unsupported value %s", v.Type())
		}
	}
	for _, example := range values {
		typ := reflect.TypeOf(example)
		t.Run(typ.Name(), func(t *testing.T) {
			a := reflect.New(typ).Elem()
			equal := reflect.New(typ).Elem()
			if !reflect.DeepEqual(a.Interface(), equal.Interface()) {
				t.Fatal("equal values differ")
			}
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)
				if !field.IsExported() {
					continue
				}
				t.Run(field.Name, func(t *testing.T) {
					different := reflect.New(typ).Elem()
					setNonzero(different.Field(i))
					if reflect.DeepEqual(a.Interface(), different.Interface()) {
						t.Fatal("different field compared equal")
					}
					switch left := a.Addr().Interface().(type) {
					case *DataCatalogError:
						if !left.Equal(equal.Addr().Interface().(*DataCatalogError)) || left.Equal(different.Addr().Interface().(*DataCatalogError)) {
							t.Fatal("catalog equality")
						}
					case *CelestrakGPRecordError:
						if !left.Equal(equal.Addr().Interface().(*CelestrakGPRecordError)) || left.Equal(different.Addr().Interface().(*CelestrakGPRecordError)) {
							t.Fatal("record equality")
						}
					}
				})
			}
		})
	}
}

func TestCatalogScalarValueEquality(t *testing.T) {
	cases := []struct {
		name                 string
		equal, same, unequal any
	}{
		{"format", CelestrakGPFormatTLE, CelestrakGPFormatTLE, CelestrakGPFormatJSON},
		{"special", CelestrakGPSpecialGPZ, CelestrakGPSpecialGPZ, CelestrakGPSpecialDecaying},
		{"group", CelestrakGPGroup("a"), CelestrakGPGroup("a"), CelestrakGPGroup("b")},
		{"number", CelestrakGPCatalogNumber(1), CelestrakGPCatalogNumber(1), CelestrakGPCatalogNumber(2)},
		{"designator", CelestrakGPInternationalDesignator("a"), CelestrakGPInternationalDesignator("a"), CelestrakGPInternationalDesignator("b")},
		{"name", CelestrakGPName("a"), CelestrakGPName("a"), CelestrakGPName("b")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.equal != tc.same || tc.equal == tc.unequal {
				t.Fatal("scalar equality")
			}
		})
	}
}
