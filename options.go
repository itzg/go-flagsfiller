package flagsfiller

import "github.com/iancoleman/strcase"

// Renamer takes a field's name and returns the flag name to be used
type Renamer func(name string) string

// DefaultFieldRenamer is used when no WithFieldRenamer option is passed to the FlagSetFiller
// constructor.
var DefaultFieldRenamer = strcase.ToKebab

// FillerOption instances are passed to the FlagSetFiller constructor.
type FillerOption func(opt *fillerOptions)

type fillerOptions struct {
	FieldRenamer Renamer
}

// WithFieldRenamer declares an option to customize the Renamer used to convert field names
// to flag names.
func WithFieldRenamer(renamer Renamer) FillerOption {
	return func(opt *fillerOptions) {
		opt.FieldRenamer = renamer
	}
}

func (o *fillerOptions) renameLongName(name string) string {
	if o.FieldRenamer == nil {
		return DefaultFieldRenamer(name)
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
