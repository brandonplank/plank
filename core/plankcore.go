package plankcore

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
)

/*
Static defines here
*/

type Data []byte

const (
	FilenameFlag   = 1 << 0 // 1
	EncryptedFlag  = 1 << 1 // 2
	CompressedFlag = 1 << 2 // 4
)

type EndSection struct {
	Filenames  []string
	FileHashes []string
}

type StaticDefines_ struct {
	AdresserSize int64
	MetaSections int64
}

// End

type PlankDecoded_ struct {
	Data      []Data
	Filenames []string
	Hashes    []string
	Keybag    string
}

type SectionDefines_ struct {
	DataStartOffset  int64
	DataSections     int64
	DataSectionSizes []int64

	DataCombined      []byte
	FilenamesCombined []byte
	FormedData        []byte
	StartOffsets      []int64
	EndOffsets        []int64

	SectionSizes []int64

	CombinedOffsets []int64

	Header    int64
	One       int64
	Two       int64
	FileNames []string
}

func chunk(in Data, chunkSize int) []Data {
	if len(in) == 0 {
		return nil
	}
	divided := make([]Data, (len(in)+chunkSize-1)/chunkSize)
	prev := 0
	i := 0
	till := len(in) - chunkSize
	for prev < till {
		next := prev + chunkSize
		divided[i] = in[prev:next]
		prev = next
		i++
	}
	divided[i] = in[prev:]
	return divided
}

func GZipCompress(s []byte) []byte {
	buf := bytes.Buffer{}
	compressed := gzip.NewWriter(&buf)
	compressed.Write(s)
	compressed.Close()
	return buf.Bytes()
}

func GZipDecompress(s []byte) []byte {
	read, _ := gzip.NewReader(bytes.NewReader(s))
	data, err := ioutil.ReadAll(read)
	if err != nil {
		panic(err)
	}
	read.Close()
	return data
}

func encodeEnd(filenames []string, hashes []string) []byte {
	var end EndSection

	end.Filenames = filenames
	end.FileHashes = hashes

	var buf bytes.Buffer

	encoder := gob.NewEncoder(&buf)
	encoder.Encode(end)
	return GZipCompress(buf.Bytes())
}

func decodeEnd(data []byte) ([]string, []string) {
	var end EndSection

	decompress := GZipDecompress(data)
	buf := bytes.NewBuffer(decompress)
	decoder := gob.NewDecoder(buf)

	decoder.Decode(&end)

	return end.Filenames, end.FileHashes
}

var StaticDefines StaticDefines_
var SectionDefines SectionDefines_

// Basic init for defaults
func init() {
	StaticDefines.AdresserSize = 16 // Bytes in a full adresser
	StaticDefines.MetaSections = 3

	SectionDefines.Header = 8
	SectionDefines.One = 8
}

func PlankEncode(data []Data, filenames []string, encrypt bool, compress bool, verbose bool) []byte {
	if verbose && compress {
		fmt.Println("Compressing data")
	}

	// Gen universal file key for use in decode

	key := make([]byte, 32)
	var aesGCM cipher.AEAD
	var nonce []byte

	if encrypt {
		if _, err := rand.Read(key); err != nil {
			panic(err.Error())
		}
		keybag := hex.EncodeToString(key)
		fmt.Printf("Key:\t%s\n", keybag)

		block, err := aes.NewCipher(key)
		if err != nil {
			panic(err.Error())
		}

		aesGCM, err = cipher.NewGCM(block)
		if err != nil {
			panic(err.Error())
		}

		nonce = make([]byte, aesGCM.NonceSize())
		if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
			panic(err.Error())
		}
		if verbose {
			fmt.Printf("Nonce:\t0x%s\n", hex.EncodeToString(nonce))
		}
	}

	var hashes []string

	for index, item := range data {
		dataHash := sha256.Sum256(item)
		hashes = append(hashes, hex.EncodeToString(dataHash[:]))
		if verbose {
			fmt.Printf("File %d sha256:\t%s\n", index+1, hex.EncodeToString(dataHash[:]))
		}

		if encrypt {
			item = aesGCM.Seal(nonce, nonce, item, nil)
		}

		if filenames != nil {
			if len(data) != len(filenames) {
				panic("Must have an equal number of data and filenames")
			}
			SectionDefines.FileNames = append(SectionDefines.FileNames, filenames[index])
		}

		if compress {
			compressed := GZipCompress(item)
			SectionDefines.DataSectionSizes = append(SectionDefines.DataSectionSizes, int64(len(compressed)))
			SectionDefines.DataCombined = append(SectionDefines.DataCombined, compressed...)
		} else {
			SectionDefines.DataSectionSizes = append(SectionDefines.DataSectionSizes, int64(len(item)))
			SectionDefines.DataCombined = append(SectionDefines.DataCombined, item...)
		}

		SectionDefines.DataSections++
	}

	SectionDefines.Two = int64(SectionDefines.DataSections) * StaticDefines.AdresserSize

	if verbose {
		fmt.Printf("File count: %d\n", SectionDefines.DataSections)
	}

	SectionDefines.SectionSizes = append(SectionDefines.SectionSizes, SectionDefines.Header, SectionDefines.One, SectionDefines.Two)

	SectionDefines.StartOffsets = append(SectionDefines.StartOffsets, 0x00, 0x08, 0x10) // Static and never change
	SectionDefines.EndOffsets = append(SectionDefines.EndOffsets, 0x07, 0xf, SectionDefines.Header+SectionDefines.One+SectionDefines.Two-1)

	// Fill in section sizes
	for i := 0; i < int(SectionDefines.DataSections); i++ {
		SectionDefines.SectionSizes = append(SectionDefines.SectionSizes, SectionDefines.DataSectionSizes[i])
	}

	// Fill in offsets
	for i := 0; i < int(len(SectionDefines.SectionSizes)); i++ {
		if i > int(StaticDefines.MetaSections)-1 {
			array := SectionDefines.SectionSizes[0 : i+1] // Go is weird, miss swift :(
			var total int64
			for _, size := range array {
				total += size
			}
			SectionDefines.EndOffsets = append(SectionDefines.EndOffsets, total-1)
			SectionDefines.StartOffsets = append(SectionDefines.StartOffsets, total-SectionDefines.DataSectionSizes[int64(i)-StaticDefines.MetaSections])
		}
	}

	Header := []byte{0x70, 0x6c, 0x61, 0x6e, 0x6b, 0x00, 0x00} // p l a n k . .

	var flags byte
	if filenames != nil {
		flags |= FilenameFlag
	}

	if compress {
		flags |= CompressedFlag
	}

	if encrypt {
		flags |= EncryptedFlag
	}

	// Append flags
	Header = append(Header, flags)

	if verbose {
		fmt.Printf("Writing plank header\n")
		fmt.Printf("%s", hex.Dump(Header))
	}

	if verbose {
		var start string
		var end string
		for _, off := range SectionDefines.StartOffsets {
			start += fmt.Sprintf("0x%x ", off)
		}

		for _, off := range SectionDefines.EndOffsets {
			end += fmt.Sprintf("0x%x ", off)
		}

		fmt.Printf("Start Offsets:\t%s\n", start)
		fmt.Printf("End Offsets:\t%s\n", end)
	}

	SectionOne := make([]byte, 8) // Should be 8 bytes ?

	binary.LittleEndian.PutUint64(SectionOne, uint64(SectionDefines.Two))

	SectionDefines.FormedData = append(SectionDefines.FormedData, Header...)
	SectionDefines.FormedData = append(SectionDefines.FormedData, SectionOne...)

	tempAddr := make([]byte, 8)

	for i := 0; i < int(SectionDefines.DataSections); i++ {
		startaddr := SectionDefines.StartOffsets[i+int(StaticDefines.MetaSections)]
		binary.LittleEndian.PutUint64(tempAddr, uint64(startaddr))
		SectionDefines.FormedData = append(SectionDefines.FormedData, tempAddr...)

		endaddr := SectionDefines.EndOffsets[i+int(StaticDefines.MetaSections)]
		binary.LittleEndian.PutUint64(tempAddr, uint64(endaddr))
		SectionDefines.FormedData = append(SectionDefines.FormedData, tempAddr...)
	}

	SectionDefines.FormedData = append(SectionDefines.FormedData, SectionDefines.DataCombined...)

	SectionDefines.FormedData = append(SectionDefines.FormedData, encodeEnd(SectionDefines.FileNames, hashes)...)

	return SectionDefines.FormedData
}

func PlankDecode(data Data, verbose bool, verify bool, keybag string) PlankDecoded_ {
	var PlankDecoded PlankDecoded_

	HeaderSection := data[0x0:0x08]
	if verbose {
		fmt.Println("Read header")
		fmt.Printf("%s", hex.Dump(HeaderSection))
	}

	// Check some static points

	flags := HeaderSection[0x7]

	hasFilenames := flags&FilenameFlag > 0
	isEncrypted := flags&EncryptedFlag > 0
	isCompressed := flags&CompressedFlag > 0

	if isEncrypted && keybag == "" {
		panic("The file is encrypted, you need a key!")
	}

	if verbose {
		fmt.Printf("hasFilenames:\t%t\nisEncrypted:\t%t\nisCompressed:\t%t\n", hasFilenames, isEncrypted, isCompressed)
	}

	SectionOne := make([]byte, 8)
	copy(SectionOne, data[0x8:SectionDefines.One+1])

	if verbose {
		fmt.Println("Section one")
		fmt.Printf("%s", hex.Dump(SectionOne))
	}

	SectionDefines.Two = int64(binary.LittleEndian.Uint64(SectionOne))

	if verbose {
		fmt.Printf("Got section 2 size: %d\n", SectionDefines.Two)
	}

	SectionTwo := make([]byte, SectionDefines.Two)
	copy(SectionTwo, data[16:SectionDefines.Header+SectionDefines.Two+SectionDefines.One])

	SectionDefines.DataSections = SectionDefines.Two / StaticDefines.AdresserSize

	if verbose {
		fmt.Printf("Data sections: %d\n", SectionDefines.DataSections)
	}

	SectionTwoRead := chunk(SectionTwo, 8)

	trackArray := make([]byte, 8)

	for i := 0; i < len(SectionTwoRead); i++ {
		for j := 0; j < len(SectionTwoRead[i]); j++ {
			trackArray[j] = SectionTwoRead[i][j]
			if j == len(SectionTwoRead[i])-1 {
				off := int64(binary.LittleEndian.Uint64(trackArray))
				SectionDefines.CombinedOffsets = append(SectionDefines.CombinedOffsets, off)
			}
		}
	}

	// Seperate offsets
	for index, item := range SectionDefines.CombinedOffsets {
		if index%2 == 0 {
			SectionDefines.StartOffsets = append(SectionDefines.StartOffsets, item)
		} else {
			SectionDefines.EndOffsets = append(SectionDefines.EndOffsets, item)
		}
	}

	if verbose {
		var start string
		var end string
		for _, off := range SectionDefines.StartOffsets {
			start += fmt.Sprintf("0x%x ", off)
		}

		for _, off := range SectionDefines.EndOffsets {
			end += fmt.Sprintf("0x%x ", off)
		}

		fmt.Printf("Start Offsets:\t%s\n", start)
		fmt.Printf("End Offsets:\t%s\n", end)
	}

	var key []byte
	var nonceSize int
	var aesGCM cipher.AEAD

	if keybag != "" {
		key, _ = hex.DecodeString(keybag)
		block, err := aes.NewCipher(key)
		if err != nil {
			panic(err.Error())
		}
		aesGCM, err = cipher.NewGCM(block)
		if err != nil {
			panic(err.Error())
		}
		nonceSize = aesGCM.NonceSize()
	}

	for i := 0; i < int(SectionDefines.DataSections); i++ {
		if isCompressed {
			PlankDecoded.Data = append(PlankDecoded.Data, GZipDecompress(data[SectionDefines.StartOffsets[i]:SectionDefines.EndOffsets[i]+1]))
		} else {
			PlankDecoded.Data = append(PlankDecoded.Data, data[SectionDefines.StartOffsets[i]:SectionDefines.EndOffsets[i]+1])
		}

		if keybag != "" {
			if i < 1 {
				if verbose {
					fmt.Println("Extracting nonce")
				}
			}
			nonce, cipher := PlankDecoded.Data[i][:nonceSize], PlankDecoded.Data[i][nonceSize:]
			if i < 1 {
				if verbose {
					fmt.Printf("Nonce:\t0x%s\n", hex.EncodeToString(nonce))
				}
			}
			PlankDecoded.Data[i], _ = aesGCM.Open(nil, nonce, cipher, nil)
		}
	}

	var totalSize int64

	for range data {
		totalSize++
	}

	startEndSectionOffset := SectionDefines.EndOffsets[len(SectionDefines.EndOffsets)-1]
	sizeOfEndSectionSection := totalSize - startEndSectionOffset
	endOfEndSectionOffset := startEndSectionOffset + sizeOfEndSectionSection - 1
	EndSection := data[startEndSectionOffset+1 : endOfEndSectionOffset+1]

	PlankDecoded.Filenames, PlankDecoded.Hashes = decodeEnd(EndSection)

	if verify {
		for i := 0; i < int(SectionDefines.DataSections); i++ {
			//Verify hash is told to
			if verify {
				hash := sha256.Sum256(PlankDecoded.Data[i])
				dataHash := hex.EncodeToString(hash[:])
				if dataHash != PlankDecoded.Hashes[i] {
					fmt.Printf("Mismatch: %s\t%s\n", dataHash, PlankDecoded.Hashes[i])
					panic("Failed to verify the hash for " + PlankDecoded.Filenames[i])
				}
				fmt.Printf("%s passed checksum\n", PlankDecoded.Filenames[i])
			}
		}
	}
	return PlankDecoded
}

// Ok here are my notes, going to keep it
/*
 Hader
 Byte 1- 5 will contain plank (the file type)
 Byte 8 contains the flags

 00000001 Is the hasFilename flag
 00000010 Is the encryption flag
 00000100 Is the made in go flag

You can or bits for flags

    |--> Encryption 1 << 1
	    |
	    | |-> Filename 1 << 0
 0  0 0 1 1 -> filenames ad encrypted
 16 8 4 2 1


Regular encoding/decoding header
 | 0x70 0x6c 0x61 0x6e 0x6b 0x00 x00 0x00 |

Section one
 Byte 1 - 8 ill contain the size of the size mapping section
 Will hold a 64 bit number
 | 0x00 0x00 0x00 0x00 0x0 0x00 0x00 0x00 |

Example
 | 0x39 x39 0x39 0x39 0x39 0x39 0x39 0x39 | Int


Section two, size will be based off the number in section one
 Needs to be in 16 byte intervals as that is how big a sectionaddresser is

Section 2 will contain the amount of sections in section 3-max
 | 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Beginning of sec3 ddress
 | 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | End of sec3 address

| 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Beginning of sec4 address
 | 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | End of sec4 address

etc

Section 3
 Some data
 | 0x70 0xf 0x67 0x67 0x65 0x72 0x73 0x00 | p o g g e r s

Section 4
 Some data
 | 0x73 0x7 0x69 0x66 0x74 0x73 0x77 0x69 | s w i f t s w i

end of sizeof(sec1) + sizeof(sec2) is Beginning of sec3 address(the beginning of sec3)

sizeof(sec1) + sizeof(sec2) + sizeof(sec3) is the end of sec3

so the current setup would look like
 | 0x20 0x00 0x00 0x00 0x00 0x00 0x000x00 | 32 bytes in section 2, as defined here take this and devide by 16 and we have 2 data sections
 | 0x29 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 3 start address
 | 0x30 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 3 end address
 | 0x31 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 4 start addres
 | 0x38 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 4 end address
 | 0x70 0x6f 0x67 0x67 0x65 0x72 0x73 0x00 | Data 1
 | 0x73 0x77 0x69 0x66 0x74 0x73 0x77 0x69 | Data 2

Note data size can differ along with section numbers

When decoding it, we first look at the first 8 bytes
 | 0x20 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | <- This hows us how many bytes is in section 2

From here, we cal calculate how many files/data are stored in our filetype
 There are 16 bytes in each section definition, the start and end addresses8 bytes for the start, 8 bytes for the end

So, if section 2 has 32 bytes, we know there are 2 files in the data section
 Formula

files = bytes / 16

Now, all we have to do is read all the values in
 | 0x29 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Secton 3 start address
 | 0x30 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 3 end address
 | 0x31 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 4 start addres
 | 0x38 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 4 end address

They are 8 bytes each (64 bit)

Once we have that, we have the offsets for the actual files data sections

| 0x70 0x6f 0x67 0x67 0x65 0x72 0x73 0x00 | Data 1
 | 0x73 0x77 0x69 0x66 0x74 0x73 0x77 0x69 | Data 2

These can have different data, i am just doing a simple demo, here is a more complex one with different lens

| 0x54 0x68 0x69 0x73 0x20 0x69 0x73 0x20 | This is
 | 0x70 0x72 0x65 0x74 0x74 0x79 0x20 0x69 | pretty
 | 0x6e 0x74 0x65 0x72 0x65 0x73 0x74 0x69 | nteresti
 | 0x6e 0x67 0x2c 0x20 0x69 0x74 0x73 0x20 | ng, its
 | 0x73 0x6f 0x20 0x63 0x6f 0x6f 0x6c 0x20 | so cool
 | 0x77 0x68 0x61 0x74 0x20 0x63 0x6f 0x64 | what co
 | 0x65 0x20 0x63 0x61 0x6e 0x20 0x64 0x6f | e can do
 | 0x53 0x77 0x69 0x66 0x74 0x20 0x69 0x73 | Swift is
 | 0x20 0x67 0x6f 0x6f 0x64 0x2c 0x20 0x66 |  good, f
 | 0x69 0x72 0x65 0x20 0x69 0x73 0x20 0x62 | ire is b
 | 0x61 0x64 0x20 0x6c 0x6d 0x61 0x6f 0x00 | ad lmao

And it has no problems distinguising between any data!

After this, we store all the filenames in a structure
 that looks like this

pseudocode

EndSection
	Filenames[
 end


this gets appended to the data section, the size of this
 section is found by taking the total size of the plank
 file and substracting the final data sections offset

dataSize - finalDataOff = EndSectionSize

then just convert the bytes back into the struct with gob
*/
