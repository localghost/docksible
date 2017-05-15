package utils

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
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

func (bc *InMemoryArchive) Add(name, data string) error {
	return bc.AddBytes(name, []byte(data))
}

func (bc *InMemoryArchive) AddBytes(name string, data []byte) error {
	if err := bc.writer.WriteHeader(&tar.Header{Name: name, Size: int64(len(data))}); err != nil {
		return err
	}
	if _, err := bc.writer.Write(data); err != nil {
		return err
	}
	return nil
}

func (bc *InMemoryArchive) AddReader(name string, data io.Reader) error {
	bytes, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}
	return bc.AddBytes(name, bytes)
}

func (bc *InMemoryArchive) Close() (*bytes.Buffer, error) {
	if err := bc.writer.Close(); err != nil {
		return nil, err
	}
	return bc.content, nil
}
