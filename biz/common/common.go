package common

import (
	"adserver/motor"
	"net/http"
)

func StartMotor() {
	http.HandleFunc("/biz/ping", handPing)
	http.HandleFunc("/biz/enter", handEnter)
	http.HandleFunc("/biz/exit", handExit)
}

func CheckLogged(ofTransit *motor.Convey) bool {
	if ofTransit.GetMap("user_logged") != "yes" {
		ofTransit.PutError("there's no user logged")
		return false
	}
	return true
}

func handPing(w http.ResponseWriter, r *http.Request) {
	transit := motor.Transit(w, r)
	transit.Set("ping", "pong")
	transit.Send()
}

func handEnter(w http.ResponseWriter, r *http.Request) {
	transit := motor.Transit(w, r)
	client := r.FormValue("client")
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	if transit.Open(client, user, pass) {
		transit.SetMap("user_logged", "yes")
		transit.SetMap("user_logged_name", user)
		transit.SetMap("user_logged_client", client)
		transit.Set("enter", "success")
	} else {
		transit.SetMap("user_logged", "no")
		transit.SetMap("user_logged_name", "")
		transit.SetMap("user_logged_client", "")
		transit.PutError("can't hand the entrance")
	}
	transit.Send()
}

func handExit(w http.ResponseWriter, r *http.Request) {
	transit := motor.Transit(w, r)
	transit.Close()
	transit.ClearMap()
	transit.Set("exit", "success")
	transit.Send()
}
