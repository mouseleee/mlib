package msql_test

import (
	"testing"
	"time"

	"github.com/mouseleee/mlib/msql"
)

func TestRegisterTable(t *testing.T) {
	s := msql.Student{
		Name:  "test",
		Age:   15,
		Birth: genBirth(1995, 6, 16),
		Base: msql.Base{
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		},
	}
	msql.RegisterTable(s, "test")
}

func genBirth(year, month, day uint) time.Time {
	return time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.Local)
}

func TestExtractTableTags(t *testing.T) {
	t.FailNow()
	s := msql.Student{
		Name:  "test",
		Age:   15,
		Birth: genBirth(1995, 6, 16),
		Base: msql.Base{
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		},
	}

	_, err := msql.ExtractColFromTableType(s)
	if err != nil {
		t.Error(err)
	}
}
