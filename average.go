package main

import (
	"time"
)

type RollingAverage struct {
	Cur time.Duration
	Num time.Duration
}

func (r *RollingAverage) Update(new time.Duration) time.Duration {
	r.Cur = ((r.Cur * r.Num) + new) / (r.Num + 1)
	r.Num++

	return r.Cur
}
