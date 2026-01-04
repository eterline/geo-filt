package forbidden

import (
	"errors"
	"net/http"
	"os"

	"github.com/eterline/geo-filt/internal/model"
)

var (
	ForbiddenWriters = NewForbiddenWriterRegistry()
)

func init() {
	ForbiddenWriters.Register("default", NewPlainWriter)
	ForbiddenWriters.Register("text", NewPlainWriter)
	ForbiddenWriters.Register("html", NewHTMLWriter)
}

// =============

type forbiddenWriter struct {
	content []byte
}

func newForbiddenWriter(p []byte) *forbiddenWriter {
	if len(p) == 0 {
		p = []byte("Forbidden: ip not allowed!")
	}
	return &forbiddenWriter{content: p}
}

func (fb *forbiddenWriter) ResponseForbidden(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusForbidden)
	_, err := w.Write(fb.content)
	return err
}

// =============

type PlainWriter struct {
	*forbiddenWriter
}

func NewPlainWriter(text string) (model.ResponseForbiddenWriter, error) {
	if text == "" {
		return nil, errors.New("text content can't be empty")
	}

	w := &PlainWriter{
		forbiddenWriter: newForbiddenWriter([]byte(text)),
	}

	return w, nil
}

type HTMLWriter struct {
	*forbiddenWriter
}

// =============

func NewHTMLWriter(file string) (model.ResponseForbiddenWriter, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	w := &HTMLWriter{
		forbiddenWriter: newForbiddenWriter(data),
	}

	return w, err
}
