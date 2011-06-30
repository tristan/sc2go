package mpq

import (
	"./file"
	"encoding/binary"
//	"bytes"
	"os"
	"fmt"
)

// derived from: http://wiki.devklog.net/index.php?title=The_MoPaQ_Archive_Format
// with some help from: https://github.com/arkx/mpyq/blob/master/mpyq.py

type MPQHeader struct {
	HeaderSize int32
	ArchiveSize int32
	FormatVersion int16
	SectorSizeShift int8
	HashTableOffset int32
	BlockTableOffset int32
	HashTableEntries int32
	BlockTableEntries int32
}

type MPQUserDataHeader struct {
	UserDataSize int32
	ArchiveHeaderOffset int32
	UserData []byte
}

type MPQHeaderExt struct {
	ExtendedBlockTableOffset int64
	HashTableOffsetHigh int16
	BlockTableOffsetHigh int16
}

type BlockEntry struct {
	BlockOffset int32
	BlockSize int32
	FileSize int32
	Flags int32
}

/* WORK IN PROGRESS
func (be *BlockEntry) Decrypt(dst, src []byte) {
	seed1 := []byte("(block entry)")
	var seed2 uint64 = 0xEEEEEEEE
	for i := 0 ; i < len(dst)/4; i++ {
		// TODO: need encryption table
		// can we have static sections to initialise stuff?
	}
}*/

type MPQFile struct {
	Shunt *MPQUserDataHeader
	Header *MPQHeader
	HeaderExt *MPQHeaderExt
	Blocks []BlockEntry
}

func readUserDataHeader(f *file.File) (udh *MPQUserDataHeader, err os.Error) {
	udh = new(MPQUserDataHeader)
	e := binary.Read(f, binary.LittleEndian, &udh.UserDataSize)
	if e != nil {
		return nil, err
	}
	e = binary.Read(f, binary.LittleEndian, &udh.ArchiveHeaderOffset)
	if e != nil {
		return nil, err
	}
	udh.UserData = make([]byte, udh.UserDataSize)
	r, e := f.Read(udh.UserData)
	if r < 1 { // 0 is EOF, TODO: support this case too
		fmt.Printf("Problem reading header: %s\n", e.String())
		return nil,e
	}
	return udh,e
}

func readHeader(f *file.File) (h *MPQHeader, err os.Error) {
	h = new(MPQHeader)
	e := binary.Read(f, binary.LittleEndian, h)
	if e != nil {
		return nil, err
	}
	return h,e
}

func readHeaderExt(f *file.File) (h *MPQHeaderExt, err os.Error) {
	h = new(MPQHeaderExt)
	e := binary.Read(f, binary.LittleEndian, h)
	if e != nil {
		return nil, err
	}
	return h,e
}

func readFile(f *file.File) (mpqFile *MPQFile, err os.Error) {

	mpqFile = new(MPQFile)

	var magic [4]byte
	r, e := f.Read(magic[:])
	if r < 1 { // 0 is EOF, TODO: support this case too
		fmt.Printf("Problem reading header: %s\n", e.String())
		return nil,e
	}

	if magic[0] != 'M' || magic[1] != 'P' || magic[2] != 'Q' {
		fmt.Printf("not a valid mpq file: %d", magic)
		return nil,e
	}

	if magic[3] == '\x1B' { // user data shunt is first
		mpqFile.Shunt, e = readUserDataHeader(f)
		//fmt.Printf("magic: %d, %d, %d, %s\n", magic, mpqFile.Shunt.UserDataSize, mpqFile.Shunt.ArchiveHeaderOffset, mpqFile.Shunt.UserData)
		if e != nil {
			return nil,e
		}
		f.Seek(int64(mpqFile.Shunt.ArchiveHeaderOffset), 0)
		r, e := f.Read(magic[:])
		switch {
		case r < 1:
			//fmt.Printf("Problem reading header: %s\n", e.String())
			return nil,e
		case r == 0: // EOF
			return mpqFile,e
		default:
		}
	}
		
	if magic[3] != '\x1A' {
		fmt.Printf("Unexpected magic value: %d\n", magic)
		return nil,os.NewError("Unexpected magic value")
	}
	
	mpqFile.Header, e = readHeader(f)
	if e != nil {
		return nil,e
	}
	if mpqFile.Header.FormatVersion == 1 {
		mpqFile.HeaderExt,e = readHeaderExt(f)
		if e != nil {
			return nil,e
		}
	}

	mpqFile.Blocks = make([]BlockEntry, mpqFile.Header.BlockTableEntries)
	//f.Seek(MPQFile.Header.BlockTableOffset

	return mpqFile, e
}

func Open(filename string) (mpqFile *MPQFile, err os.Error) {
	f, e := file.Open(filename)
	if f == nil {
		return nil,e
	}
	return readFile(f)
}