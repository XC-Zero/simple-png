package main

import (
	"encoding/binary"
	"errors"
	"strings"
	"time"
)

var b binary.ByteOrder = binary.BigEndian

type ChunkName string

const (
	IHDRChunk ChunkName = "IHDR"
	PLTEChunk ChunkName = "PLTE"
	IDATChunk ChunkName = "IDAT"
	IENDChunk ChunkName = "IEND"

	TRNSChunk ChunkName = "tRNS"
	PHYSChunk ChunkName = "pHYs"
	TEXTChunk ChunkName = "tEXt"
	ZTXTChunk ChunkName = "zTXT"
	TIMEChunk ChunkName = "tIME"
)

// ISO_3309_CRC x32+x26+x23+x22+x16+x12+x11+x10+x8+x7+x5+x4+x2+x+1
// TODO implement crc
var ISO_3309_CRC = []uint{1, 1, 0, 1, 1, 0, 1, 1, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 1, 0, 0, 0, 0, 0, 1}

type ChunkParse interface {
	ChunkName() ChunkName
	Parse(chunk *chunk) error
}

type chunk struct {
	len  [4]byte
	code [4]byte
	data []byte
	crc  [4]byte
}

/*

--------------------------------------------------------------------------------------

*/

type IHDR struct {
	Width             uint32
	Height            uint32
	BitDepth          uint8
	ColorType         uint8
	ComdivssionMethod uint8
	FilterMethod      uint8
	InterlaceMethod   uint8
}

func (c *IHDR) Parse(chunk *chunk) error {
	code := ChunkName(chunk.code[:])
	if code != IHDRChunk {
		return errors.New("invalid chunk code")
	}
	if chunk.data == nil || len(chunk.data) < 13 {
		return errors.New("invalid chunk data")
	}
	c.Width = b.Uint32(chunk.data[:4])
	c.Height = b.Uint32(chunk.data[4:8])
	c.BitDepth = chunk.data[8]
	c.ColorType = chunk.data[9]
	c.ComdivssionMethod = chunk.data[10]
	c.FilterMethod = chunk.data[11]
	c.InterlaceMethod = chunk.data[12]
	return nil
}

func (c *IHDR) ChunkName() ChunkName {
	return IHDRChunk
}

/*

--------------------------------------------------------------------------------------

*/

type IEND struct {
	IDAT
}

func (i *IEND) ChunkName() ChunkName {
	return IENDChunk
}

/*

--------------------------------------------------------------------------------------

*/

type PLTE struct {
	IDAT
}

func (P *PLTE) ChunkName() ChunkName {
	return PLTEChunk
}

/*

--------------------------------------------------------------------------------------

*/

type IDAT struct {
	Length        uint32
	ChunkTypeCode string
	Data          []byte
}

func (i *IDAT) ChunkName() ChunkName {
	return IDATChunk
}

func (i *IDAT) Parse(chunk *chunk) error {
	i.Length = b.Uint32(chunk.len[:])
	i.ChunkTypeCode = string(chunk.code[:])
	i.Data = chunk.data[:]
	return nil
}

/*

--------------------------------------------------------------------------------------

*/

type TRNS struct {
}

func (T *TRNS) ChunkName() ChunkName {
	return TRNSChunk
}

func (T *TRNS) Parse(chunk *chunk) error {
	//TODO implement me
	panic("implement me")
}

/*

--------------------------------------------------------------------------------------

*/

type PHYS struct {
}

func (P *PHYS) ChunkName() ChunkName {
	return PHYSChunk
}

func (P *PHYS) Parse(chunk *chunk) error {
	//TODO implement me
	panic("implement me")
}

/*

--------------------------------------------------------------------------------------

*/

// TEXT
//
// Textual information that the encoder wishes to record with the image can be stored in tEXt chunks. Each tEXt chunk contains a keyword and a text string, in the format:
//
//	Keyword:        1-79 bytes (character string)
//	Null separator: 1 byte
//	Text:           n bytes (character string)
//
// The keyword and text string are separated by a zero byte (null character). Neither the keyword nor the text string can contain a null character. Note that the text string is not null-terminated (the length of the chunk is sufficient information to locate the ending). The keyword must be at least one character and less than 80 characters long. The text string can be of any length from zero bytes up to the maximum permissible chunk size less the length of the keyword and separator.
type TEXT struct {
	Keyword   string
	Separator string
	Text      string
}

func (t *TEXT) ChunkName() ChunkName {
	return TEXTChunk
}

const nullSep = string(byte(0x00))

func (t *TEXT) Parse(chunk *chunk) error {
	str := strings.TrimSpace(string(chunk.data[:]))
	strs := strings.Split(str, nullSep)
	if len(strs) != 2 {
		return errors.New("invalid text")
	}
	t.Keyword = strs[0]
	t.Separator = " "
	t.Text = strs[1]

	return nil
}

/*

--------------------------------------------------------------------------------------

*/

// ZTXT
// The zTXt chunk contains textual data, just as tEXt does; however, zTXt takes advantage of compression. zTXt and tEXt chunks are semantically equivalent, but zTXt is recommended for storing large blocks of text.
// A zTXt chunk contains:
//
//	Keyword:            1-79 bytes (character string)
//	Null separator:     1 byte
//	Compression method: 1 byte
//	Compressed text:    n bytes
//
// The keyword and null separator are exactly the same as in the tEXt chunk. Note that the keyword is not compressed. The compression method byte identifies the compression method used in this zTXt chunk. The only value presently defined for it is 0 (deflate/inflate compression). The compression method byte is followed by a compressed datastream that makes up the remainder of the chunk. For compression method 0, this datastream adheres to the zlib datastream format (see Deflate/Inflate Compression). Decompression of this datastream yields Latin-1 text that is identical to the text that would be stored in an equivalent tEXt chunk.
// Any number of zTXt and tEXt chunks can appear in the same file. See the preceding definition of the tEXt chunk for the predefined keywords and the recommended format of the text.
//
// See Recommendations for Encoders: Text chunk processing, and Recommendations for Decoders: Text chunk processing.
type ZTXT struct {
	Keyword           string
	Separator         string
	CompressionMethod uint8
	Text              string
}

func (Z *ZTXT) ChunkName() ChunkName {
	return ZTXTChunk
}

func (Z *ZTXT) Parse(chunk *chunk) error {
	return nil
}

/*

--------------------------------------------------------------------------------------

*/

// TIME
//
// The tIME chunk gives the time of the last image modification (not the time of initial image creation). It contains:
//
//	Year:   2 bytes (complete; for example, 1995, not 95)
//	Month:  1 byte (1-12)
//	Day:    1 byte (1-31)
//	Hour:   1 byte (0-23)
//	Minute: 1 byte (0-59)
//	Second: 1 byte (0-60)    (yes, 60, for leap seconds; not 61,
//	                          a common error)
//
// Universal Time (UTC, also called GMT) should be specified rather than local time.
// The tIME chunk is intended for use as an automatically-applied time stamp that is updated whenever the image data is changed. It is recommended that tIME not be changed by PNG editors that do not change the image data. See also the Creation Time tEXt keyword, which can be used for a user-supplied time.
type TIME struct {
	Year   uint16
	Month  uint8
	Day    uint8
	Hour   uint8
	Minute uint8
	Second uint8
}

func (t *TIME) ChunkName() ChunkName {
	return TIMEChunk
}

func (t *TIME) Parse(chunk *chunk) error {
	t.Year = b.Uint16(chunk.data[:2])
	t.Month = chunk.data[2]
	t.Day = chunk.data[3]
	t.Hour = chunk.data[4]
	t.Minute = chunk.data[5]
	t.Second = chunk.data[6]
	return nil
}

func (t *TIME) ToTime() time.Time {
	return time.Date(int(t.Year), time.Month(t.Month), int(t.Day), int(t.Hour), int(t.Minute), int(t.Second), 0, time.UTC)
}
