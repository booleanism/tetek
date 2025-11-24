package helper

import (
	"sort"

	"github.com/booleanism/tetek/comments/amqp"
	"github.com/google/uuid"
)

func BuildCommentTree(tables []amqp.Comment, head uuid.UUID) []amqp.Comment {
	tree := []amqp.Comment{}
	for _, v := range tables {
		if v.Parent.String() == head.String() {
			v.Child = BuildCommentTree(tables, v.ID)
			tree = append(tree, v)
		}
	}

	sort.Slice(tree, func(i, j int) bool {
		return tree[i].Points > tree[j].Points
	})

	return tree
}
