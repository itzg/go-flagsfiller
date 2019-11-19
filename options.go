package flagsfiller

import "github.com/iancoleman/strcase"

// Renamer takes a field's name and returns the flag name to be used
type Renamer func(name string) string

// DefaultFieldRenamer is used when no WithFieldRenamer option is passed to the FlagSetFiller
// constructor.
var DefaultFieldRenamer = KebabRenamer()

// FillerOption instances are passed to the FlagSetFiller constructor.
type FillerOption func(opt *fillerOptions)

type fillerOptions struct {
	fieldRenamer Renamer
	envRenamer   Renamer
}

// WithFieldRenamer declares an option to customize the Renamer used to convert field names
// to flag names.
func WithFieldRenamer(renamer Renamer) FillerOption {
	return func(opt *fillerOptions) {
		opt.fieldRenamer = renamer
	}
}

// WithEnv activates pre-setting the flag values from environment variables.
// Fields are mapped to environment variables names by prepending the given prefix and
// converting word-wise to SCREAMING_SNAKE_CASE. The given prefix can be empty.
func WithEnv(prefix string) FillerOption {
	return WithEnvRenamer(
		CompositeRenamer(PrefixRenamer(prefix), ScreamingSnakeRenamer()))
}

// WithEnvRenamer activates pre-setting the flag values from environment variables where fields
// are mapped to environment variable names by applying the given Renamer
func WithEnvRenamer(renamer Renamer) FillerOption {
	return func(opt *fillerOptions) {
		opt.envRenamer = renamer
	}
}

func (o *fillerOptions) renameLongName(name string) string {
	if o.fieldRenamer == nil {
		return DefaultFieldRenamer(name)
	} else {
		return o.fieldRenamer(name)
	}
}

func newFillerOptions(options ...FillerOption) *fillerOptions {
	v := &fillerOptions{}
	for _, opt := range options {
		opt(v)
	}
	return v
}

// PrefixRenamer prepends the given prefix to a name
func PrefixRenamer(prefix string) Renamer {
	return func(name string) string {
		return prefix + name
	}
}

// KebabRenamer converts a given name into kebab-case
func KebabRenamer() Renamer {
	return strcase.ToKebab
}

// ScreamingSnakeRenamer converts a given name into SCREAMING_SNAKE_CASE
func ScreamingSnakeRenamer() Renamer {
	return strcase.ToScreamingSnake
}

// CompositeRenamer applies all of the given Renamers to a name
func CompositeRenamer(renamers ...Renamer) Renamer {
	return func(name string) string {
		for _, r := range renamers {
			name = r(name)
		}
		return name
	}
}
