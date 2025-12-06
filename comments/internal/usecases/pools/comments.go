package pools

import (
	"sync"

	"github.com/booleanism/tetek/comments/internal/internal/domain/entities"
	"github.com/booleanism/tetek/comments/internal/internal/domain/model/pools"
)

type Comments = pools.Comments

var CommentsPool = sync.Pool{
	New: func() any {
		return &Comments{
			Value: make([]entities.Comment, 0, 1024),
		}
	},
}
