package flagsfiller_test

import (
	"fmt"
	"github.com/itzg/go-flagsfiller"
)

func ExampleCompositeRenamer() {
	renamer := flagsfiller.CompositeRenamer(
		flagsfiller.PrefixRenamer("App"),
		flagsfiller.ScreamingSnakeRenamer())

	renamed := renamer("SomeFieldName")
	fmt.Println(renamed)
	// Output:
	// APP_SOME_FIELD_NAME
}
