package motor

import (
	"fmt"
	"runtime/debug"
	"strings"
)

type aError struct {
	Message string
	Stack   string
}

func (err *aError) getTail() string {
	if err.Message == "" {
		return ""
	}
	return " because " + err.Message
}

func (err *aError) stackMessage(message string) {
	err.Message = message + err.getTail()
}

func (err *aError) prepareToSend() {
	err.Stack = string(debug.Stack()[:])
	err.Message = strings.Title(strings.ToLower(err.Message)) + "."
}

func (transit *Convey) HasError() bool {
	return transit.err != nil
}

func (transit *Convey) IfHasErrorPut(messageParts ...interface{}) *Convey {
	if transit.HasError() {
		transit.PutError(messageParts...)
	}
	return transit
}

func (transit *Convey) PutError(messageParts ...interface{}) *Convey {
	if transit.err == nil {
		transit.err = &aError{}
	}
	message := fmt.Sprint(messageParts...)
	if message == "" {
		message = "strange empty error message"
	}
	transit.err.stackMessage(message)
	return transit
}
