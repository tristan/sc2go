package encryptedfile

import (
	"./file"
	"strings"
	"os"
	"fmt"
	)

func initCryptTable() (ct []uint32) {
	var seed uint32 = 0x00100001

	ct = make([]uint32, 0x500)

	for index1 := 0; index1 < 0x100; index1++ {
		for index2, i := index1, 0; i < 5; i, index2 = i+1, index2+0x100 {
			seed = (seed * 125 + 3) % 0x2AAAAB
			temp1 := (seed & 0xFFFF) << 0x10
			seed = (seed * 125 + 3) % 0x2AAAAB
			temp2 := (seed & 0xFFFF)
			ct[index2] = (temp1 | temp2)
		}
	}
	return ct
}

var cryptTable []uint32

func init() {
	cryptTable = initCryptTable()
}

const (
	MPQ_HASH_TABLE_OFFSET = 0
	MPQ_HASH_NAME_A = 1
	MPQ_HASH_NAME_B = 2
	MPQ_HASH_FILE_KEY = 3
)

func HashString(s string, hashType uint32) (hash uint32) {
	var seed1 uint32 = 0x7FED7FED
	var seed2 uint32 = 0xEEEEEEEE
	s = strings.ToUpper(s)
	for _, ch := range s {
		chi := uint32(ch)
		seed1 = cryptTable[(hashType * 0x100) + chi] ^ (seed1 + seed2)
		seed2 = chi + seed1 + seed2 + (seed2 << 5) + 3
	}
	return seed1
}

type EncryptedFile struct {
	file *file.File
	key uint32
}

func bytesToUint32(pos int, b []byte) (ret uint32) {
	ret = 0
	ret += uint32(b[pos])
	ret += uint32(b[pos+1]) << 8
	ret += uint32(b[pos+2]) << 16
	ret += uint32(b[pos+3]) << 24
	return ret
}

func uint32toBytes(val uint32, pos int, b []byte) {
	b[pos+0] = byte(val)
	b[pos+1] = byte(val >> 8)
	b[pos+2] = byte(val >> 16)
	b[pos+3] = byte(val >> 24)
}


func (f *EncryptedFile) Read(b []byte) (ret int, err os.Error) {

	var seed uint32 = 0xEEEEEEEE
	var ch uint32
	key := f.key

	r,e := f.file.Read(b)
	//fmt.Printf("EF.Read(%s) read: %d bytes\n", key, r)
	fmt.Printf("encrypted: ")
	for _, c := range b {
		fmt.Printf("%x ", c)
	}
	fmt.Printf("\n")
	r /= 4
	for i := 0; i < r; i++ {
		seed += cryptTable[0x400 + (key & 0xFF)]
		ch = bytesToUint32(i*4, b) ^ (key + seed)
		key = ((^key << 0x15) + 0x11111111) | (key >> 0x0B)
		seed = ch + seed + (seed << 5) + 3
		uint32toBytes(ch, i*4, b)
	}
	fmt.Printf("decrypted: ")
	for _, c := range b {
		fmt.Printf("%x ", c)
	}
	fmt.Printf("\n")
	return r,e
}


func NewEncryptedFile(file *file.File, key uint32) (f *EncryptedFile) {
	f = new(EncryptedFile)
	f.file = file
	f.key = key
	return f
}