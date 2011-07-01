package mpq

import (
	"./file"
	"./encryptedfile"
	"encoding/binary"
//	"bytes"
	"os"
	"fmt"
//	"strings"
)

// derived from: http://wiki.devklog.net/index.php?title=The_MoPaQ_Archive_Format
// with some help from: https://github.com/arkx/mpyq/blob/master/mpyq.py

type MPQHeader struct {
	Magic [4]byte
	HeaderSize int32
	ArchiveSize int32
	FormatVersion int16
	SectorSizeShift int16 // MoPaQ wiki lists this as int8, but this fails
	HashTableOffset int32
	BlockTableOffset int32
	HashTableEntries int32
	BlockTableEntries int32
}

type MPQUserDataHeader struct {
	Magic [4]byte
	UserDataSize int32
	ArchiveHeaderOffset int32
	UserData []byte
}

type MPQHeaderExt struct {
	ExtendedBlockTableOffset int64
	HashTableOffsetHigh int16
	BlockTableOffsetHigh int16
}

type BlockTableEntry struct {
	BlockOffset int32
	BlockSize int32
	FileSize int32
	Flags int32
}

type HashTableEntry struct {
	FilePathHashA int32
	FilePathHashB int32
	Language int16
	Platform int16 // another case where int8 is listed
	FileBlockIndex int32
}

type MPQFile struct {
	file *file.File
	UserDataHeader *MPQUserDataHeader
	Header *MPQHeader
	HeaderExt *MPQHeaderExt
	BlockTable []BlockTableEntry
	HashTable []HashTableEntry
}

func (f *MPQFile) readUserDataHeader() (e os.Error) {
	f.UserDataHeader = new(MPQUserDataHeader)
	e = binary.Read(f.file, binary.LittleEndian,
		&f.UserDataHeader.Magic)
	if e != nil {
		return e
	}
	e = binary.Read(f.file, binary.LittleEndian, 
		&f.UserDataHeader.UserDataSize)
	if e != nil {
		return e
	}
	e = binary.Read(f.file, binary.LittleEndian, 
		&f.UserDataHeader.ArchiveHeaderOffset)
	if e != nil {
		return e
	}
	f.UserDataHeader.UserData = make([]byte, f.UserDataHeader.UserDataSize)
	_, e = f.file.Read(f.UserDataHeader.UserData)
	return e
}

func (f *MPQFile) readHeader() (e os.Error) {
	f.Header = new(MPQHeader)
	e = binary.Read(f.file, binary.LittleEndian, f.Header)
	return e
}

func (f *MPQFile) readHeaderExt() (e os.Error) {
	f.HeaderExt = new(MPQHeaderExt)
	e = binary.Read(f.file, binary.LittleEndian, f.HeaderExt)
	return e
}

func (f *MPQFile) readHashTable() (e os.Error) {
	var key uint32 = encryptedfile.HashString("(hash table)", 
		encryptedfile.MPQ_HASH_TABLE_OFFSET)
	f.file.Seek(int64(f.Header.HashTableOffset), 0)
	f.HashTable = make([]HashTableEntry, f.Header.HashTableEntries)
	fmt.Printf("trying to read %d hash table entries", f.Header.HashTableEntries)
	ef := encryptedfile.NewEncryptedFile(f.file, key)
	e = binary.Read(ef, binary.LittleEndian, f.HashTable)
	return e
}

func readFile(f *file.File) (mpqFile *MPQFile, err os.Error) {

	mpqFile = new(MPQFile)
	mpqFile.file = f

	var magic [4]byte
	r, e := mpqFile.file.Read(magic[:])
	mpqFile.file.Seek(0, 0)
	if r < 1 { // 0 is EOF, TODO: support this case too
		fmt.Printf("Problem reading header: %s\n", e.String())
		return nil,e
	}

	if magic[0] != 'M' || magic[1] != 'P' || magic[2] != 'Q' {
		fmt.Printf("not a valid mpq file: %d", magic)
		return nil,e
	}

	if magic[3] == '\x1B' { // user data header is first
		e = mpqFile.readUserDataHeader()
		if e != nil {
			return nil,e
		}
		mpqFile.file.Seek(int64(mpqFile.UserDataHeader.ArchiveHeaderOffset), 0)
		r, e := mpqFile.file.Read(magic[:])
		mpqFile.file.Seek(int64(mpqFile.UserDataHeader.ArchiveHeaderOffset), 0)
		switch {
		case r < 1:
			//fmt.Printf("Problem reading header: %s\n", e.String())
			return nil,e
		case r == 0: // EOF
			return mpqFile,e
		default:
		}
	} else {
		mpqFile.UserDataHeader = new(MPQUserDataHeader)
		mpqFile.UserDataHeader.ArchiveHeaderOffset = 0
	}
		
	if magic[3] != '\x1A' {
		fmt.Printf("Unexpected magic value: %d\n", magic)
		return nil,os.NewError("Unexpected magic value")
	}
	
	e = mpqFile.readHeader()
	if e != nil {
		return nil,e
	}
	if mpqFile.Header.FormatVersion == 1 {
		e = mpqFile.readHeaderExt()
		if e != nil {
			return nil,e
		}
	}

	e = mpqFile.readHashTable()
	
	return mpqFile, e
}

func Open(filename string) (mpqFile *MPQFile, err os.Error) {
	f, e := file.Open(filename)
	if f == nil {
		return nil,e
	}
	return readFile(f)
}