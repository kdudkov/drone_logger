package main

import (
	"fmt"
	"sync"
	"time"
)

type DroneInfo struct {
	info    sync.Map
	updated time.Time
}

func (d *DroneInfo) getString(key string) string {
	if v, ok := d.info.Load(key); ok {
		return fmt.Sprintf("%v", v)
	}

	return ""
}

func (d *DroneInfo) getFloat(key string) float64 {
	if v, ok := d.info.Load(key); ok {
		if vv, ok2 := v.(float64); ok2 {
			return vv
		}
	}

	return 0
}

func (d *DroneInfo) getByte(key string) byte {
	if v, ok := d.info.Load(key); ok {
		if vv, ok2 := v.(byte); ok2 {
			return vv
		}
	}

	return 0
}

func (d *DroneInfo) getInt(key string) int {
	if v, ok := d.info.Load(key); ok {
		if vv, ok2 := v.(int); ok2 {
			return vv
		}
	}

	return 0
}

func (d *DroneInfo) getBool(key string) bool {
	if v, ok := d.info.Load(key); ok {
		if vv, ok2 := v.(bool); ok2 {
			return vv
		}
	}

	return false
}

func (d *DroneInfo) put(key string, v any) {
	d.info.Store(key, v)
}
