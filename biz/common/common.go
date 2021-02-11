package common

import (
	"adserver/motor"
	"encoding/json"
	"net/http"
)

// StartHandlers starts all the handlers on common.
func StartHandlers() {
	http.HandleFunc("/biz/ping", handPing)
	http.HandleFunc("/biz/enter", handEnter)
	http.HandleFunc("/biz/connect", handConnect)
	http.HandleFunc("/biz/exit", handExit)
}

// CheckConnected checks if sessions is connected to a database,
// puts an error and returns false if not.
func CheckConnected(transit *motor.Convey) bool {
	if transit.GetMapped("user_connected") != "yes" {
		transit.PutError("there's no user logged")
		return false
	}
	return true
}

func handPing(w http.ResponseWriter, r *http.Request) {
	transit := motor.Transit(w, r)
	if r.Method == "GET" {
		transit.Set("ping", "pong")
	}
	transit.Send()
}

func handEnter(w http.ResponseWriter, r *http.Request) {
	transit := motor.Transit(w, r)
	if r.Method == "GET" {
		transit.Set("uid", transit.Session().GetUID())
	}
	transit.Send()
}

func handConnect(w http.ResponseWriter, r *http.Request) {
	transit := motor.Transit(w, r)
	if r.Method == "POST" {
		received := struct {
			Client string
			User   string
			Pass   string
		}{}
		json.NewDecoder(r.Body).Decode(&received)
		if transit.Open(received.Client, received.User, received.Pass) {
			transit.SetMapped("user_connected", "yes")
			transit.SetMapped("user_connected_name", received.User)
			transit.SetMapped("user_connected_client", received.Client)
			transit.Set("enter", "success")
		} else {
			transit.SetMapped("user_connected", "no")
			transit.SetMapped("user_connected_name", "")
			transit.SetMapped("user_connected_client", "")
			transit.PutError("can't hand the entrance")
		}
	}
	transit.Send()
}

func handExit(w http.ResponseWriter, r *http.Request) {
	transit := motor.Transit(w, r)
	transit.Close()
	transit.ClearMapped()
	transit.Set("exit", "success")
	transit.Send()
}
