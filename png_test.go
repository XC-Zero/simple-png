package main

import (
	"log"
	"os"
	"testing"
)

func TestParsePng(t *testing.T) {
	open, err := os.Open("./demo.png")
	if err != nil {
		panic(err)
	}
	p, err := ParsePng(open)
	if err != nil {
		panic(err)
	}
	log.Println(*p.IDATs[0])
	log.Println(*p.TEXTs[0])
	for i := range p.chunks {
		log.Println(string(p.chunks[i].code[:]))
	}
}

type CustomChunkParse struct {
}

func (c *CustomChunkParse) ChunkName() ChunkName {
	return ChunkName("cust")
}

func (c *CustomChunkParse) Parse(chunk *chunk) error {
	// your custom parse
	return nil
}

func TestCustomChunk(t *testing.T) {
	open, err := os.Open("./demo.png")
	if err != nil {
		panic(err)
	}
	p, err := ParsePng(open)
	if err != nil {
		panic(err)
	}
	c := &CustomChunkParse{}
	err = p.ParseChunk(c)
	if err != nil {
		panic(err)
	}
}
