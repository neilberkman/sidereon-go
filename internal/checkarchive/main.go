// Command checkarchive rejects LLVM bitcode and LTO sections in static archives.
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
	"strconv"
	"strings"
)

func main() {
	for _, path := range os.Args[1:] {
		if err := checkArchive(path); err != nil {
			fmt.Fprintf(os.Stderr, "native archive error: %s: %v\n", path, err)
			os.Exit(1)
		}
	}
}

func checkArchive(path string) (result error) {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil && result == nil {
			result = err
		}
	}()
	magic := make([]byte, 8)
	if _, err := io.ReadFull(f, magic); err != nil {
		return err
	}
	if string(magic) != "!<arch>\n" {
		return fmt.Errorf("invalid archive signature")
	}
	for {
		header := make([]byte, 60)
		if _, err := io.ReadFull(f, header); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if string(header[58:]) != "`\n" {
			return fmt.Errorf("invalid archive member header")
		}
		size, err := strconv.ParseInt(strings.TrimSpace(string(header[48:58])), 10, 64)
		if err != nil || size < 0 {
			return fmt.Errorf("invalid archive member size %q", header[48:58])
		}
		offset, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		name := strings.TrimSpace(string(header[:16]))
		objectOffset, objectSize := offset, size
		if strings.HasPrefix(name, "#1/") {
			length, err := strconv.ParseInt(name[3:], 10, 32)
			if err != nil || length < 0 || length > size {
				return fmt.Errorf("invalid extended member name %q", name)
			}
			text := make([]byte, length)
			if _, err := io.ReadFull(f, text); err != nil {
				return err
			}
			name = strings.TrimRight(string(text), "\x00")
			objectOffset += length
			objectSize -= length
		}
		if name != "/" && name != "//" && name != "/SYM64/" && !strings.HasPrefix(name, "__.SYMDEF") {
			if err := checkObject(io.NewSectionReader(f, objectOffset, objectSize)); err != nil {
				return fmt.Errorf("member %s: %w", name, err)
			}
		}
		if _, err := f.Seek(offset+size+(size&1), io.SeekStart); err != nil {
			return err
		}
	}
}

func ltoSection(name string) bool {
	return strings.HasPrefix(name, ".gnu.lto_") || strings.HasPrefix(name, ".gnu.debuglto_") ||
		strings.HasPrefix(name, ".llvm.lto") || name == ".llvmbc" || name == ".llvmcmd" ||
		name == "__bitcode" || name == "__bundle"
}

func checkObject(r *io.SectionReader) error {
	magic := make([]byte, 4)
	if _, err := r.ReadAt(magic, 0); err != nil {
		return err
	}
	if bytes.Equal(magic, []byte{'B', 'C', 0xc0, 0xde}) || bytes.Equal(magic, []byte{0xde, 0xc0, 0x17, 0x0b}) {
		return fmt.Errorf("disallowed LLVM bitcode object")
	}
	if bytes.Equal(magic, []byte{0x7f, 'E', 'L', 'F'}) {
		f, err := elf.NewFile(r)
		if err != nil {
			return err
		}
		for _, s := range f.Sections {
			if ltoSection(s.Name) || s.Type == 0x6fff4c0c {
				return fmt.Errorf("disallowed LLVM bitcode/LTO section %q", s.Name)
			}
		}
		return nil
	}
	m := binary.LittleEndian.Uint32(magic)
	if m == macho.Magic64 || m == macho.Magic32 || binary.BigEndian.Uint32(magic) == macho.Magic64 || binary.BigEndian.Uint32(magic) == macho.Magic32 {
		f, err := macho.NewFile(r)
		if err != nil {
			return err
		}
		for _, load := range f.Loads {
			if s, ok := load.(*macho.Segment); ok && (s.Name == "__LLVM" || s.Name == "__GNU_LTO") {
				return fmt.Errorf("disallowed LLVM bitcode/LTO segment %q", s.Name)
			}
		}
		for _, s := range f.Sections {
			if s.Seg == "__LLVM" || s.Seg == "__GNU_LTO" || ltoSection(s.Name) {
				return fmt.Errorf("disallowed LLVM bitcode/LTO section %s,%s", s.Seg, s.Name)
			}
		}
		return nil
	}
	// Short COFF import members contain symbol/DLL names, not sections.
	if bytes.Equal(magic, []byte{0, 0, 0xff, 0xff}) {
		header := make([]byte, 20)
		if _, err := r.ReadAt(header, 0); err != nil {
			return err
		}
		if binary.LittleEndian.Uint16(header[4:6]) == 0 && int64(binary.LittleEndian.Uint32(header[12:16]))+20 == r.Size() {
			return nil
		}
	}
	var coff io.ReaderAt = r
	// debug/pe reads a 96-byte DOS probe even for smaller COFF objects.
	if r.Size() < 96 {
		if r.Size() < 20 {
			return fmt.Errorf("truncated COFF header")
		}
		probe := make([]byte, 96)
		if _, err := r.ReadAt(probe[:r.Size()], 0); err != nil {
			return err
		}
		coff = bytes.NewReader(probe)
	}
	f, err := pe.NewFile(coff)
	if err != nil {
		return fmt.Errorf("unrecognized object: %w", err)
	}
	for _, s := range f.Sections {
		if ltoSection(s.Name) {
			return fmt.Errorf("disallowed LLVM bitcode/LTO section %q", s.Name)
		}
	}
	return nil
}
