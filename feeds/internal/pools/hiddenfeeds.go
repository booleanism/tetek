package pools

import (
	"sync"

	"github.com/booleanism/tetek/feeds/internal/model"
)

type HiddenFeeds struct {
	Value []model.HiddenFeed
}

func (f *HiddenFeeds) Reset() {
	f.Value = f.Value[:0]
}

var HiddenFeedsPool = sync.Pool{
	New: func() any {
		return &HiddenFeeds{
			Value: make([]model.HiddenFeed, 0, 512),
		}
	},
}
