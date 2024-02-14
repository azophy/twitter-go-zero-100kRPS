package main

import (
	"errors"
	"sync"
	"time"
)

const minExpireDurationDiff = 100 * time.Millisecond

type Cache[ValType any] struct {
	expireTime time.Time
	val        ValType
	mtx        sync.Mutex
}

func (c *Cache[ValType]) Reset() {
	c.expireTime = *new(time.Time)
	c.val = *new(ValType)
}

func (c *Cache[ValType]) Fetch(expireDuration time.Duration, loader func() (ValType, error)) (ValType, error) {
	var emptyVal ValType
	if expireDuration < minExpireDurationDiff {
		return emptyVal, errors.New("expireDuration is too small")
	}

	now := time.Now()
	newExpireTime := now.Add(expireDuration)
	if c.expireTime.Before(now) {
		c.mtx.Lock()
		defer c.mtx.Unlock()
		// double check to make sure it haven't been updated "recently". here we
		// avoid using `Equal()` check in case several `Fetch()` are being called
		// almost at the same time. so some may have slightly different `now`
		// values (although still called after expiry time). if we use `Equal()`
		// check here, those different `now`'s would be not 'equal' and each would
		// acquire a new lock, thus defeating our lock mechanism here.
		if c.expireTime.Sub(newExpireTime).Abs() > minExpireDurationDiff {
			c.expireTime = newExpireTime
			val, err := loader()
			if err != nil {
				return emptyVal, err
			}
			c.val = val
		}
	}

	return c.val, nil
}
