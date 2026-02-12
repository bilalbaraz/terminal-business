package random

import (
	"math/rand"
	"time"
)

type Source struct {
	r *rand.Rand
}

func New() *Source {
	return &Source{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (s *Source) Int63() int64 {
	return s.r.Int63()
}
