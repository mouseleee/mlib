package mutil_test

import (
	"os"
	"testing"

	"github.com/mouseleee/mlib/mutil"
)

func TestWriteFile(t *testing.T) {
	f := "./test.txt"
	err := mutil.WriteFile(f, []byte("test write"))
	if err != nil {
		t.Error(err)
	}

	os.Remove(f)
}
