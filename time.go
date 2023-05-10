package flagsfiller

import (
	"reflect"
	"time"
)

func init() {
	RegisterSimpleType(timeConverter)
}

// DefaultTimeLayout is the default layout string to parse time, following golang time.Parse() format,
// can be overridden per field by field tag "layout". Default value is "2006-01-02 15:04:05", which is
// the same as time.DateTime in Go 1.20
var DefaultTimeLayout = "2006-01-02 15:04:05"

func timeConverter(s string, tag reflect.StructTag) (time.Time, error) {
	layout, _ := tag.Lookup("layout")
	if layout == "" {
		layout = DefaultTimeLayout
	}
	return time.Parse(layout, s)
}
