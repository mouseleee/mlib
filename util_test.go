package mouselib_test

import (
	"testing"

	"github.com/mouseleee/mouselib"
)

// func TestWriteFile(t *testing.T) {
// 	f := "./test.txt"
// 	err := mouselib.WriteFile(f, []byte("test write"))
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	os.Remove(f)
// }

func TestCamelToUnderline(t *testing.T) {
	c := []string{
		"", "mewo", "Mewo", "FunFair", "BiLiBiLi",
	}
	e := []string{
		"", "mewo", "mewo", "fun_fair", "bi_li_bi_li",
	}

	for i, tc := range c {
		if v := mouselib.CamelToUnderline(tc); v != e[i] {
			t.Logf("c: %s v: %s e: %s\n", c[i], v, e[i])
			t.Fail()
		}
	}
}
