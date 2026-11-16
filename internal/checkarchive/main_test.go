package main

import (
	"bytes"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func encode(t *testing.T, values ...any) []byte {
	t.Helper()
	var b bytes.Buffer
	for _, v := range values {
		if err := binary.Write(&b, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	return b.Bytes()
}

func elfObject(t *testing.T, name string, kind uint32) []byte {
	t.Helper()
	names := []byte("\x00.shstrtab\x00" + name + "\x00")
	h := elf.Header64{Type: uint16(elf.ET_REL), Machine: uint16(elf.EM_X86_64), Version: 1, Shoff: 64, Ehsize: 64, Shentsize: 64, Shnum: 3, Shstrndx: 1}
	copy(h.Ident[:], "\x7fELF\x02\x01\x01")
	return encode(t, h, elf.Section64{}, elf.Section64{Name: 1, Type: uint32(elf.SHT_STRTAB), Off: 256, Size: uint64(len(names))}, elf.Section64{Name: 11, Type: kind}, names)
}

func machoObject(t *testing.T, segment, section string) []byte {
	t.Helper()
	// A 64-bit Mach-O header, one LC_SEGMENT_64, and one section.
	h := []uint32{macho.Magic64, uint32(macho.CpuArm64), 0, uint32(macho.TypeObj), 1, 152, 0, 0}
	load := make([]byte, 72)
	binary.LittleEndian.PutUint32(load, uint32(macho.LoadCmdSegment64))
	binary.LittleEndian.PutUint32(load[4:], 152)
	binary.LittleEndian.PutUint32(load[64:], 1)
	var s macho.Section64
	copy(s.Name[:], section)
	copy(s.Seg[:], segment)
	return encode(t, h, load, s)
}

func coffObject(t *testing.T, name string) []byte {
	t.Helper()
	h := pe.FileHeader{Machine: pe.IMAGE_FILE_MACHINE_AMD64, NumberOfSections: 1}
	var s pe.SectionHeader32
	copy(s.Name[:], name)
	return encode(t, h, s)
}

func TestArchiveSections(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		lto  bool
	}{
		{"macho-arm64-bitcode", machoObject(t, "__LLVM", "__bitcode"), true},
		{"macho-gnu-lto", machoObject(t, "__GNU_LTO", "code"), true},
		{"macho-machine-code", machoObject(t, "__TEXT", "__text"), false},
		{"elf-llvm-bitcode", elfObject(t, ".llvmbc", 1), true},
		{"elf-gcc-lto", elfObject(t, ".gnu.lto_code", 1), true},
		{"elf-debug-lto", elfObject(t, ".gnu.debuglto_code", 1), true},
		{"elf-lto-type", elfObject(t, ".code", 0x6fff4c0c), true},
		{"elf-machine-code", elfObject(t, ".text", 1), false},
		{"coff-llvm-bitcode", coffObject(t, ".llvmbc"), true},
		{"coff-machine-code", coffObject(t, ".text"), false},
		{"raw-bitcode", []byte{'B', 'C', 0xc0, 0xde}, true},
		{"wrapped-bitcode", []byte{0xde, 0xc0, 0x17, 0x0b}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, extended := range []bool{false, true} {
				name := "code.o/"
				data := append([]byte(nil), tc.data...)
				if extended {
					name = "#1/6"
					data = append([]byte("code.o"), data...)
				}
				var ar bytes.Buffer
				fmt.Fprint(&ar, "!<arch>\n")
				fmt.Fprintf(&ar, "%-16s%-12d%-6d%-6d%-8s%-10d`\n", name, 0, 0, 0, "644", len(data))
				ar.Write(data)
				if len(data)%2 != 0 {
					ar.WriteByte('\n')
				}
				path := filepath.Join(t.TempDir(), "object.a")
				if err := os.WriteFile(path, ar.Bytes(), 0600); err != nil {
					t.Fatal(err)
				}
				err := checkArchive(path)
				if tc.lto {
					if err == nil || !strings.Contains(err.Error(), "disallowed") {
						t.Fatalf("extended=%v: expected LTO refusal, got %v", extended, err)
					}
				} else if err != nil {
					t.Fatalf("extended=%v: %v", extended, err)
				}
			}
		})
	}
}

func TestShortCOFFImportHasNoSections(t *testing.T) {
	header := make([]byte, 20)
	copy(header, []byte{0, 0, 0xff, 0xff})
	binary.LittleEndian.PutUint32(header[12:], 4)
	data := append(header, []byte{'x', 0, 'y', 0}...)
	if err := checkObject(io.NewSectionReader(bytes.NewReader(data), 0, int64(len(data)))); err != nil {
		t.Fatal(err)
	}
}
