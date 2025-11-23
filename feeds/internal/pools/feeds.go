package pools

import (
	"sync"

	"github.com/booleanism/tetek/feeds/internal/model"
)

type Feeds struct {
	Value []model.Feed
}

func (f *Feeds) Reset() {
	f.Value = f.Value[:0]
}

var FeedsPool = sync.Pool{
	New: func() any {
		return &Feeds{
			Value: make([]model.Feed, 0, 1024),
		}
	},
}
