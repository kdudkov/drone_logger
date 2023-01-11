package main

import (
	"sync"
	"time"
)

type DroneInfo struct {
	info    sync.Map
	updated time.Time
}

func (d *DroneInfo) getFloat(key string) float64 {
	if v, ok := d.info.Load(key); ok {
		if vv, ok2 := v.(float64); ok2 {
			return vv
		}
	}

	return 0
}

func (d *DroneInfo) put(key string, v any) {
	d.info.Store(key, v)
}
