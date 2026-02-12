package app

import "time"

type Clock interface {
	Now() time.Time
}

type RNG interface {
	Int63() int64
}
