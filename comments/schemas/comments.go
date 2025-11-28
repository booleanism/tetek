package schemas

import (
	_ "embed"
)

//go:embed comments.sql
var CommentsSQL string
