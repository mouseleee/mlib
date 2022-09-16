package mouselib_test

import (
	"os"
	"testing"

	"github.com/mouseleee/mouselib"
)

func TestWriteFile(t *testing.T) {
	f := "./test.txt"
	err := mouselib.WriteFile(f, []byte("test write"))
	if err != nil {
		t.Error(err)
	}

	os.Remove(f)
}
