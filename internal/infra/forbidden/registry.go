package forbidden

import (
	"fmt"

	"github.com/eterline/geo-filt/internal/model"
)

type ForbiddenWriterFactory func(content string) (model.ResponseForbiddenWriter, error)

type ForbiddenWriterRegistry struct {
	reg map[string]ForbiddenWriterFactory
}

func NewForbiddenWriterRegistry() *ForbiddenWriterRegistry {
	return &ForbiddenWriterRegistry{
		reg: make(map[string]ForbiddenWriterFactory),
	}
}

func (r *ForbiddenWriterRegistry) Register(name string, fac ForbiddenWriterFactory) {
	if _, ok := r.reg[name]; ok {
		panic(fmt.Sprintf("forbidden writer '%s' already registered", name))
	}
	r.reg[name] = fac
}

func (r *ForbiddenWriterRegistry) Build(writerType, content string) (model.ResponseForbiddenWriter, error) {
	factory, ok := r.reg[writerType]
	if !ok {
		return nil, fmt.Errorf("unknown forbidden writer type: %s", writerType)
	}

	w, err := factory(content)
	if err != nil {
		return nil, fmt.Errorf("failed to build forbidden writer '%s': %w", writerType, err)
	}

	return w, nil
}

func InitForbiddenWriter(writerType, content string) (model.ResponseForbiddenWriter, error) {
	if writerType == "" {
		writerType = "default"
	}
	return ForbiddenWriters.Build(writerType, content)
}
