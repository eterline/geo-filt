package forbidden

import (
	"net/http"
	"os"
)

type PlainWriter struct {
	Text []byte
}

func NewPlainWriter(text string) *PlainWriter {
	if text == "" {
		text = "Forbidden - ip not allowed"
	}
	return &PlainWriter{Text: []byte(text)}
}

func (p *PlainWriter) ResponseForbidden(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
	w.Write(p.Text)
}

type HTMLWriter struct {
	FilePath string
	content  []byte
}

func NewHTMLWriter(filePath string) *HTMLWriter {
	return &HTMLWriter{
		FilePath: filePath,
	}
}

func (h *HTMLWriter) loadContent() error {
	if h.content != nil {
		return nil
	}
	data, err := os.ReadFile(h.FilePath)
	if err != nil {
		return err
	}
	h.content = data
	return nil
}

func (h *HTMLWriter) ResponseForbidden(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)

	if err := h.loadContent(); err != nil {
		w.Write([]byte("Forbidden - ip not allowed"))
		return
	}

	_, _ = w.Write(h.content)
}
