package schemas

import (
	_ "embed"
)

//go:embed feeds.sql
var FeedsSQL string
