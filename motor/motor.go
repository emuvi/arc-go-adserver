package motor

import (
	"log"
	"net/http"
	"strconv"
)

func Start(port int, handlersStarters ...func()) {
	for _, handlerStarter := range handlersStarters {
		handlerStarter()
	}
	go maintainSessions()
	log.Panic(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
