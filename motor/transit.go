package motor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Convey struct {
	response http.ResponseWriter
	request  *http.Request
	session  *aSession
	link     *pgxpool.Conn
	rows     pgx.Rows
	values   map[string]interface{}
	err      *aError
	carry    map[string]interface{}
	sent     bool
}

func Transit(w http.ResponseWriter, r *http.Request) *Convey {
	return &Convey{
		response: w,
		request:  r,
		session:  nil,
		link:     nil,
		rows:     nil,
		values:   nil,
		err:      nil,
		carry:    nil,
		sent:     false,
	}
}

func (transit *Convey) Session() *aSession {
	if transit.session == nil {
		transit.session = popSession(transit.response, transit.request)
	}
	return transit.session
}

func (transit *Convey) Get(name string) interface{} {
	if transit.carry == nil {
		return nil
	}
	return transit.carry[name]
}

func (transit *Convey) Set(name string, value interface{}) *Convey {
	if transit.carry == nil {
		transit.carry = map[string]interface{}{}
	}
	transit.carry[name] = value
	return transit
}

func (transit *Convey) Clear(name string, value interface{}) *Convey {
	transit.carry = nil
	return transit
}

func (transit *Convey) GetMapped(key string) string {
	return transit.Session().getMapped(key)
}

func (transit *Convey) SetMapped(key, value string) *Convey {
	transit.Session().setMapped(key, value)
	return transit
}

func (transit *Convey) ClearMapped() *Convey {
	transit.Session().clearMapped()
	return transit
}

func (transit *Convey) GetCookie(key string) string {
	cookie, err := transit.request.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (transit *Convey) Send() {
	if transit.sent {
		fmt.Println("You've tried to send a transit twice on:")
		debug.PrintStack()
		return
	}
	transit.Done()
	transit.release()
	transit.response.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(transit.response)
	if transit.err == nil {
		transit.response.WriteHeader(http.StatusOK)
		encoder.Encode(transit.carry)
	} else {
		transit.err.prepareToSend()
		transit.response.WriteHeader(http.StatusInternalServerError)
		encoder.Encode(transit.err)
	}
	transit.sent = true
}
