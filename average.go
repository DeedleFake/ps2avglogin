package main

import (
	"time"
)

type RollingAverage struct {
	cur time.Duration
	num time.Duration
}

func (r *RollingAverage) Update(new time.Duration) time.Duration {
	r.cur = ((r.cur * r.num) + new) / (r.num + 1)
	r.num++

	return r.cur
}

func (r *RollingAverage) Get() time.Duration {
	return r.cur
}
