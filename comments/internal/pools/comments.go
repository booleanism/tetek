package pools

import (
	"sync"

	"github.com/booleanism/tetek/comments/internal/model"
)

type Comments struct {
	Value []model.Comment
}

func (f *Comments) Reset() {
	f.Value = f.Value[:0]
}

var CommentsPool = sync.Pool{
	New: func() any {
		return &Comments{
			Value: make([]model.Comment, 0, 1024),
		}
	},
}
