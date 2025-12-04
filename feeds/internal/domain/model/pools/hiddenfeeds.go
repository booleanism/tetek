package pools

import (
	"sync"

	"github.com/booleanism/tetek/feeds/internal/domain/entities"
)

type HiddenFeeds struct {
	Value []entities.HiddenFeed
}

func (f *HiddenFeeds) Reset() {
	f.Value = f.Value[:0]
}

var HiddenFeedsPool = sync.Pool{
	New: func() any {
		return &HiddenFeeds{
			Value: make([]entities.HiddenFeed, 0, 512),
		}
	},
}
