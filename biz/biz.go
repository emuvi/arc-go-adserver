package biz

import (
	"adserver/biz/common"
	"adserver/biz/company"
	"adserver/biz/sales"
)

// StartMotor sets up all the biz handlers
func StartMotor() {
	common.StartMotor()
	company.StartMotor()
	sales.StartMotor()
}
