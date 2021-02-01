package motor

import (
	"log"
	"net/http"
	"strconv"
)

func StartListen(onPort int) {
	go maintainSessions()
	log.Panic(http.ListenAndServe(":"+strconv.Itoa(onPort), nil))
}
