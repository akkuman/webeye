package cache

import (
	"errors"
	"reflect"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type MemoryCache struct {
	c *gocache.Cache
}

var _ Cache = &MemoryCache{}

// 新建缓存实例
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		c: gocache.New(1*time.Minute, 1*time.Minute),
	}
}

// 获取
func (c *MemoryCache) Get(key string, ptrOut any) error {
	v, found := c.c.Get(key)
	if !found {
		return ErrKeyNotFound
	}
	value := reflect.ValueOf(ptrOut)
	// If e represents a value as opposed to a pointer, the answer won't
	// get back to the caller. Make sure it's a pointer.
	if value.Type().Kind() != reflect.Pointer {
		return errors.New("gob: attempt to decode into a non-pointer")
	}
	if value.IsValid() {
		if value.Kind() == reflect.Pointer && !value.IsNil() {
			// That's okay, we'll store through the pointer.
		} else if !value.CanSet() {
			return errors.New("gob: DecodeValue of unassignable value")
		}
	}
	reflect.Indirect(value).Set(reflect.ValueOf(v))
	return nil
}

// 删除
func (c *MemoryCache) Del(key string) error {
	c.c.Delete(key)
	return nil
}

// 设置
func (c *MemoryCache) Set(key string, v any, d time.Duration) error {
	c.c.Set(key, v, d)
	return nil
}
