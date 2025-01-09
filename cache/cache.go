package cache

import "time"

type Cache interface {
	Set(key string, value any, timeout time.Duration) error
	Get(key string, ptrOut any) error
	Del(key string) (error)
}