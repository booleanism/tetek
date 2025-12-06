package pools

import "github.com/booleanism/tetek/comments/internal/internal/domain/entities"

type Comments struct {
	Value []entities.Comment
}

func (f *Comments) Reset() {
	f.Value = f.Value[:0]
}
