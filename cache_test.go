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
	_, err := testCache.Fetch(10*time.Millisecond, func() (int,error) { return 1,nil })

	if err == nil {
		t.Error("method should be error")
	}
}

func TestSingleCache(t *testing.T) {
	testCache.Reset()

	shouldBe := 11
	shouldBe2 := 13
	content, _ := testCache.Fetch(3*time.Second, func() (int,error) { return shouldBe, nil })
	t.Logf("content now : %d", content)

	if content != shouldBe {
		t.Errorf("wrong! got %d should be %d", content, shouldBe)
	}

	content, _ = testCache.Fetch(3*time.Second, func() (int,error) { return shouldBe2, nil })
	t.Logf("content now : %d", content)

	if content != shouldBe {
		t.Errorf("wrong! got %d should be %d", content, shouldBe)
	}

	time.Sleep(4 * time.Second)

	content, _ = testCache.Fetch(3*time.Second, func() (int,error) { return shouldBe2, nil })
	t.Logf("content sekarang : %d", content)

	if content != shouldBe2 {
		t.Errorf("wrong! got %d should be %d", content, shouldBe2)
	}

	testCache.Reset()

	content, _ = testCache.Fetch(3*time.Second, func() (int,error) { return shouldBe, nil })
	t.Logf("content now : %d", content)

	if content != shouldBe {
		t.Errorf("wrong! got %d should be %d", content, shouldBe)
	}
}

func TestParallelCache(t *testing.T) {
	testCache.Reset()

	callCounter := 0

	task := func(wg *sync.WaitGroup, val int) {
		defer wg.Done()

		loader := func() (int,error) {
			callCounter++
			return val, nil
		}

		_, err := testCache.Fetch(3*time.Second, loader)
		if err != nil {
			t.Errorf("Error during fetch")
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go task(&wg, i)
	}

	wg.Wait()

	if callCounter != 1 {
		t.Errorf("task called %d times, shouldBe 1", callCounter)
	}
}

func TestExpireRaceCache(t *testing.T) {
	testCache.Reset()

	callCounter := 0

	task := func(val int) {
		loader := func() (int,error) {
			callCounter++

      time.Sleep(2 * time.Second) // simulate long running process
			return val, nil
		}

		_, err := testCache.Fetch(5*time.Second, loader)
		if err != nil {
			t.Errorf("Error during fetch")
		}
	}

  task(23)
  time.Sleep(7 * time.Second) // simulate expiring cache
  go task(31)
  time.Sleep(time.Second) // simulate small discrepencies
  go task(37)
  time.Sleep(time.Second) // make sure cache is filled

	if callCounter != 2 {
		t.Errorf("task called %d times, shouldBe 2", callCounter)
	}

  res, err := testCache.Fetch(5*time.Second, func() (int,error) { return 41,nil })
  if err != nil {
    t.Errorf("Error during fetch")
  }
  if res != 31 {
		t.Errorf("cache content is %d should be 31", res)
  }

  time.Sleep(7 * time.Second) // simulate expiring cache
  res, err = testCache.Fetch(5*time.Second, func() (int,error) { return 43,nil })
  if err != nil {
    t.Errorf("Error during fetch")
  }
  if res != 43 {
		t.Errorf("cache content is %d should be 43", res)
  }
}
