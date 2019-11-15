package flagsfiller

import "github.com/iancoleman/strcase"

type Renamer func(name string) string

var defaultLongRenamer = strcase.ToKebab

type FillerOption func(opt *fillerOptions)

type fillerOptions struct {
	FieldRenamer Renamer
}

func WithFieldRenamer(renamer Renamer) FillerOption {
	return func(opt *fillerOptions) {
		opt.FieldRenamer = renamer
	}
}

func (o *fillerOptions) RenameLongName(name string) string {
	if o.FieldRenamer == nil {
		return defaultLongRenamer(name)
	} else {
		return o.FieldRenamer(name)
	}
}

func newFillerOptions(options []FillerOption) *fillerOptions {
	v := &fillerOptions{}
	for _, opt := range options {
		opt(v)
	}
	return v
}
