package mredis_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mouseleee/mlib/mredis"
)

func TestRedisStrSet(t *testing.T) {
	c := mredis.NewRedisClient("localhost:6379", 0)

	ks := []string{"k1", "k2", "k3", "k4"}
	vs := []string{"v1", "v2", "v3", "v4"}

	for i := range ks {
		err := mredis.StrSet(c, ks[i], vs[i])
		if err != nil {
			t.Error(err)
		}
	}

	for i := range ks {
		v := mredis.StrGet(c, ks[i])
		if v != vs[i] {
			t.Fail()
		}
	}

	for _, v := range ks {
		mredis.Del(c, v)
	}
}

func TestRedisLock(t *testing.T) {
	c := mredis.NewRedisClient("localhost:6379", 0)

	// A获取锁，B无法修改，A释放锁，B修改
	lock := "l"

	wg := sync.WaitGroup{}
	wg.Add(5)

	for i := 0; i < 5; i++ {
		go func(i int) {
			defer wg.Done()
			if ok := mredis.Lock(c, lock, 500*time.Millisecond); ok {
				k := fmt.Sprint(i)
				mredis.StrSet(c, k, "mewo")
			}
		}(i)
	}

	wg.Wait()
}

func TestRedisListSet(t *testing.T) {
	c := mredis.NewRedisClient("localhost:6379", 0)

	err := mredis.ListSetAndLpush(c, "ls", "1", "2", "3")
	if err != nil {
		t.Error(err)
	}
}
