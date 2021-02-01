package biz

import (
	"adserver/biz/common"
	"adserver/biz/company"
	"adserver/biz/sales"
)

func StartHandlers() {
	common.StartHandlers()
	company.StartHandlers()
	sales.StartHandlers()
}
