package helper

import (
	"sort"

	messaging "github.com/booleanism/tetek/comments/infra/messaging/rabbitmq"
	"github.com/google/uuid"
)

func BuildCommentTree(tables []messaging.Comment, head uuid.UUID) []messaging.Comment {
	tree := []messaging.Comment{}
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
