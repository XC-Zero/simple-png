package simple_png

import (
	"io"
	"slices"
	"sync"

	"github.com/pkg/errors"
)

var pngHeaderBytes = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
var pngHeader = string(pngHeaderBytes)

type Png struct {
	sync.RWMutex
	IHDR  *IHDR
	IDATs []*IDAT
	PLTE  *PLTE
	BKGD  *BKGD
	CHRM  *CHRM
	GAMA  *GAMA
	HIST  *HIST
	PHYS  *PHYS
	SBIT  *SBIT

	TEXTs []*TEXT
	TRNS  *TRNS
	TIME  *TIME
	ZTXTs []*ZTXT

	IEND       *IEND
	OtherChunk map[ChunkName][]ChunkParse
	chunks     []*chunk
	bs         []byte
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
	length := by.Uint32(l)
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

func (p *Png) ParseChunk(c ChunkParse, notSave ...bool) error {
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
	if !(len(notSave) > 0 && notSave[0]) {
		p.RLock()
		x, ok := p.OtherChunk[c.ChunkName()]
		p.RUnlock()
		if ok {
			x = append(x, c)
			p.Lock()
			p.OtherChunk[c.ChunkName()] = x
			p.Unlock()
		} else {
			p.Lock()
			p.OtherChunk[c.ChunkName()] = []ChunkParse{c}
			p.Unlock()
		}
	}
	return chunkNotFoundErr

}

func (p *Png) parseBaseChunk() error {
	p.Lock()
	defer p.Unlock()
	var IHDR = &IHDR{}
	err := p.ParseChunk(IHDR, true)
	if err != nil {
		return errors.WithStack(err)
	}
	p.IHDR = IHDR

	var IDATs []*IDAT
	for {
		var idat = &IDAT{}
		err := p.ParseChunk(idat, true)
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

	var PLTE = &PLTE{}
	err = p.ParseChunk(PLTE, true)
	if err == nil {
		p.PLTE = PLTE
	}

	var BKGD = &BKGD{}
	err = p.ParseChunk(BKGD, true)
	if err == nil {
		p.BKGD = BKGD
	}

	var CHRM = &CHRM{}
	err = p.ParseChunk(CHRM, true)
	if err == nil {
		p.CHRM = CHRM
	}
	var GAMA = &GAMA{}
	err = p.ParseChunk(GAMA, true)
	if err == nil {
		p.GAMA = GAMA
	}
	var HIST = &HIST{}
	err = p.ParseChunk(HIST, true)
	if err == nil {
		p.HIST = HIST
	}
	var PHYS = &PHYS{}
	err = p.ParseChunk(PHYS, true)
	if err == nil {
		p.PHYS = PHYS
	}

	var SBIT = &SBIT{}
	err = p.ParseChunk(SBIT, true)
	if err == nil {
		p.SBIT = SBIT
	}
	var TEXTs []*TEXT
	for {
		var text = &TEXT{}
		err := p.ParseChunk(text, true)
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

	var TRNS = &TRNS{}
	err = p.ParseChunk(TRNS, true)
	if err == nil {
		p.TRNS = TRNS
	}

	var TIME = &TIME{}
	err = p.ParseChunk(TIME, true)
	if err == nil {
		p.TIME = TIME
	}

	var ZTXTs []*ZTXT
	for {
		var text = &ZTXT{}
		err := p.ParseChunk(text, true)
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

	var IEND = &IEND{}
	err = p.ParseChunk(IEND, true)
	if err != nil {
		return errors.WithStack(err)
	}
	p.IEND = IEND
	return nil
}

func (p *Png) GetOtherChunkByName(name ChunkName) ([]ChunkParse, error) {
	p.RLock()
	defer p.RUnlock()
	if list, ok := p.OtherChunk[name]; ok {
		return list, nil
	}
	return nil, chunkNotFoundErr
}
