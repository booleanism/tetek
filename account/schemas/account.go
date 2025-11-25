package schemas

import (
	_ "embed"
)

//go:embed account.sql
var AccountSQL string
