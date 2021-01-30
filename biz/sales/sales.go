package sales

import (
	"adserver/biz/common"
	"adserver/motor"
	"net/http"
)

func StartMotor() {
	http.HandleFunc("/biz/sales/desk", handSalesDesk)
}

func handSalesDesk(w http.ResponseWriter, r *http.Request) {
	transit := motor.Transit(w, r)
	if !common.CheckLogged(transit) {
		transit.PutError("can't hand the sales desk").Send()
		return
	}
	if !GetLastPreOrders(transit) {
		transit.PutError("can't hand the sales desk").Send()
		return
	}
	transit.Send()
}
