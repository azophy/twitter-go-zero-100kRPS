package main

import (
	"sync"
	"testing"
	"time"
)

var (
	testCache = Cache[int]{}
)

func TestValidExpireTime(t *testing.T) {
	_, err := testCache.Fetch(10*time.Millisecond, func() int { return 1 })

	if err == nil {
		t.Error("SALAH! harusnya error")
	}
}

func TestSingleCache(t *testing.T) {
	testCache.Reset()

	seharusnya := 5
	seharusnya2 := 7
	isi, _ := testCache.Fetch(3*time.Second, func() int { return seharusnya })
	t.Logf("isi sekarang : %d", isi)

	if isi != seharusnya {
		t.Errorf("SALAH! dapat %d seharusnya %d", isi, seharusnya)
	}

	isi, _ = testCache.Fetch(3*time.Second, func() int { return seharusnya2 })
	t.Logf("isi sekarang : %d", isi)

	if isi != seharusnya {
		t.Errorf("SALAH! dapat %d seharusnya %d", isi, seharusnya)
	}

	time.Sleep(4 * time.Second)

	isi, _ = testCache.Fetch(3*time.Second, func() int { return seharusnya2 })
	t.Logf("isi sekarang : %d", isi)

	if isi != seharusnya2 {
		t.Errorf("SALAH! dapat %d seharusnya %d", isi, seharusnya2)
	}

	testCache.Reset()

	isi, _ = testCache.Fetch(3*time.Second, func() int { return seharusnya })
	t.Logf("isi sekarang : %d", isi)

	if isi != seharusnya {
		t.Errorf("SALAH! dapat %d seharusnya %d", isi, seharusnya)
	}
}

func TestParallelCache(t *testing.T) {
	testCache.Reset()

	callCounter := 0

	task := func(wg *sync.WaitGroup, val int) {
		defer wg.Done()

		loader := func() int {
			callCounter++
			return val
		}

		_, err := testCache.Fetch(3*time.Second, loader)
		if err != nil {
			t.Errorf("Error ketika fetch")
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go task(&wg, i)
	}

	wg.Wait()

	if callCounter != 1 {
		t.Errorf("task dipanggil sebanyak %d seharusnya 1", callCounter)
	}
}
