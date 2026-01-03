package mmapreader

import (
	"io"
	"os"
)

type FileReadCloser struct {
	f      *os.File
	offset int64
	size   int64
}

func NewFileReadCloser(path string) (*FileReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	return &FileReadCloser{
		f:      f,
		offset: 0,
		size:   info.Size(),
	}, nil
}

func (r *FileReadCloser) Read(p []byte) (int, error) {
	if r.offset >= r.size {
		return 0, io.EOF
	}

	n, err := r.f.ReadAt(p, r.offset)
	r.offset += int64(n)
	if err == io.EOF && r.offset < r.size {
		err = nil
	}

	return n, err
}

func (r *FileReadCloser) Close() error {
	if r.f != nil {
		err := r.f.Close()
		r.f = nil
		return err
	}
	return nil
}
