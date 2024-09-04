package main

import (
	"io"
	"slices"

	"github.com/pkg/errors"
)

var pngHeaderBytes = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
var pngHeader = string(pngHeaderBytes)

type Png struct {
	IHDR   *IHDR
	IDATs  []*IDAT
	TEXTs  []*TEXT
	ZTXTs  []*ZTXT
	IEND   *IEND
	TIME   *TIME
	chunks []*chunk
	bs     []byte
}

func ParsePng(r io.Reader) (*Png, error) {
	var p = &Png{}
	var hex = make([]byte, 8)
	read, err := r.Read(hex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if read != 8 || string(hex) != pngHeader {
		return nil, errors.WithStack(errors.New("invalid png"))
	}
	for {
		chunk, err := readChunk(r)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		p.chunks = append(p.chunks, chunk)
		if ChunkName(chunk.code[:]) == IENDChunk {
			break
		}
	}
	err = p.parseBaseChunk()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return p, nil
}

func readChunk(r io.Reader) (*chunk, error) {
	var l = make([]byte, 4)
	var name = make([]byte, 4)
	var crc = make([]byte, 4)

	_, err := r.Read(l)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	_, err = r.Read(name)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	length := b.Uint32(l)
	var content = make([]byte, length)
	_, err = r.Read(content)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	_, err = r.Read(crc)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &chunk{
		len:  [4]byte(l),
		code: [4]byte(name),
		data: content,
		crc:  [4]byte(crc),
	}, nil
}

var chunkNotFoundErr = errors.New("chunk not found")

func (p *Png) ParseChunk(c ChunkParse) error {
	var nChunks = slices.Clone(p.chunks)
	for i := range p.chunks {
		if p.chunks[i] == nil {
			continue
		}
		cc := *p.chunks[i]
		if ChunkName(cc.code[:]) != c.ChunkName() {
			continue
		}
		err := c.Parse(p.chunks[i])
		if err != nil {
			return errors.WithStack(err)
		}
		if i != len(nChunks)-1 {
			nChunks = append(nChunks[:i], nChunks[i+1:]...)
		} else {
			nChunks = nChunks[:i]
		}
		p.chunks = nChunks
		return nil
	}
	return chunkNotFoundErr

}

func (p *Png) parseBaseChunk() error {
	var IHDR = &IHDR{}
	err := p.ParseChunk(IHDR)
	if err != nil {
		return errors.WithStack(err)
	}
	p.IHDR = IHDR

	var IDATs []*IDAT
	for {
		var idat = &IDAT{}
		err := p.ParseChunk(idat)
		if err != nil {
			if errors.Is(err, chunkNotFoundErr) {
				break
			} else {
				return errors.WithStack(err)
			}
		}
		IDATs = append(IDATs, idat)
	}
	if len(IDATs) == 0 {
		return errors.New("no IDAT found")
	}
	p.IDATs = IDATs

	var TEXTs []*TEXT
	for {
		var text = &TEXT{}
		err := p.ParseChunk(text)
		if err != nil {
			if errors.Is(err, chunkNotFoundErr) {
				break
			} else {
				return errors.WithStack(err)
			}
		}
		TEXTs = append(TEXTs, text)
	}
	p.TEXTs = TEXTs

	var ZTXTs []*ZTXT
	for {
		var text = &ZTXT{}
		err := p.ParseChunk(text)
		if err != nil {
			if errors.Is(err, chunkNotFoundErr) {
				break
			} else {
				return errors.WithStack(err)
			}
		}
		ZTXTs = append(ZTXTs, text)
	}
	p.ZTXTs = ZTXTs

	var TIME = &TIME{}
	err = p.ParseChunk(TIME)
	if err == nil {
		p.TIME = TIME
	}

	var IEND = &IEND{}
	err = p.ParseChunk(IEND)
	if err != nil {
		return errors.WithStack(err)
	}
	p.IEND = IEND
	return nil
}
