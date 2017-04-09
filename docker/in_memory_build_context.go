package docker

import (
	"archive/tar"
	"bytes"
)

type InMemoryBuildContext struct {
	content *bytes.Buffer
	writer  *tar.Writer
}

func NewInMemoryBuildContext() *InMemoryBuildContext {
	buf := new(bytes.Buffer)
	return &InMemoryBuildContext{
		writer:  tar.NewWriter(buf),
		content: buf,
	}
}

func (bc *InMemoryBuildContext) Add(name, data string) {
	bc.writer.WriteHeader(&tar.Header{Name: name, Size: int64(len(data))})
	bc.writer.Write([]byte(data))
}

func (bc *InMemoryBuildContext) Close() *bytes.Buffer {
	bc.writer.Close()
	return bc.content
}
