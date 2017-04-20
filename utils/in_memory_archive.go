package utils

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"log"
)

type InMemoryArchive struct {
	content *bytes.Buffer
	writer  *tar.Writer
}

func NewInMemoryArchive() *InMemoryArchive {
	buf := new(bytes.Buffer)
	return &InMemoryArchive{
		writer:  tar.NewWriter(buf),
		content: buf,
	}
}

func (bc *InMemoryArchive) Add(name, data string) {
	bc.AddBytes(name, []byte(data))
}

func (bc *InMemoryArchive) AddBytes(name string, data []byte) {
	bc.writer.WriteHeader(&tar.Header{Name: name, Size: int64(len(data))})
	bc.writer.Write(data)
}

func (bc *InMemoryArchive) AddReader(name string, data io.Reader) {
	bytes, err := ioutil.ReadAll(data)
	if err != nil {
		log.Fatal(err)
	}
	bc.AddBytes(name, bytes)
}

func (bc *InMemoryArchive) Close() *bytes.Buffer {
	bc.writer.Close()
	return bc.content
}
