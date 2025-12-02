package pools

import (
	"sync"

	"github.com/booleanism/tetek/feeds/internal/entities"
)

type Feeds struct {
	Value []entities.Feed
}

func (f *Feeds) Reset() {
	f.Value = f.Value[:0]
}

var FeedsPool = sync.Pool{
	New: func() any {
		return &Feeds{
			Value: make([]entities.Feed, 0, 1024),
		}
	},
}
