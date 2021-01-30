package motor

import (
	"adserver/guide"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
)

type aStore struct {
	DB     *pgx.Conn
	Client string
	User   string
}

func openStore(session *aSession, client, user, pass string) error {
	closeStore(session)
	storesHost := guide.Configs.GetString("StoreHost", "pointel.pointto.us")
	storesPort := guide.Configs.GetInt("StorePort", 5432)
	conn, err := pgx.Connect(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", user, pass, storesHost, storesPort, client))
	if err != nil {
		return err
	}
	session.mutex.Lock()
	defer session.mutex.Unlock()
	session.store = &aStore{DB: conn, Client: client, User: user}
	return nil
}

func closeStore(session *aSession) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	if session.store != nil && session.store.DB != nil {
		session.store.DB.Close(context.Background())
	}
	session.store = nil
}

func (transit *Convey) Open(client, user, pass string) bool {
	err := openStore(transit.getSession(), client, user, pass)
	if err != nil {
		transit.PutError(err.Error())
		return false
	}
	return true
}

func (transit *Convey) Close() *Convey {
	transit.getSession().close()
	return transit
}

func (transit *Convey) Store() *aStore {
	result := transit.getSession().store
	if result == nil {
		transit.PutError("can't find the store of the session")
	}
	return result
}

func (transit *Convey) Query(sql string, args ...interface{}) bool {
	transit.rows = nil
	transit.values = nil
	store := transit.Store()
	if store == nil {
		return false
	}
	rows, err := store.DB.Query(context.Background(), sql, args...)
	if err != nil {
		transit.PutError(err.Error())
		return false
	}
	transit.rows = rows
	return true
}

func (transit *Convey) Next() bool {
	if transit.rows == nil {
		transit.PutError("there's no rows from a query")
		transit.PutError("can't get the next row")
		return false
	}
	transit.values = nil
	return transit.rows.Next()
}

func (transit *Convey) tryTakeValues() error {
	if transit.values == nil {
		if transit.rows == nil {
			return errors.New("there's no rows from a query")
		}
		values, err := transit.rows.Values()
		if err != nil {
			return err
		}
		descriptions := transit.rows.FieldDescriptions()
		if descriptions == nil {
			return errors.New("can't get the descriptions of the rows")
		}
		if len(values) != len(descriptions) {
			return errors.New("different number of columns between values and descriptions")
		}
		transit.values = map[string]interface{}{}
		for idx, desc := range descriptions {
			transit.values[string(desc.Name)] = values[idx]
		}
	}
	return nil
}

func (transit *Convey) takeValues() bool {
	err := transit.tryTakeValues()
	if err != nil {
		transit.PutError(err)
		return false
	}
	return true
}

func (transit *Convey) Take(column string) (interface{}, error) {
	err := transit.tryTakeValues()
	if err != nil {
		return nil, err
	}
	return transit.values[column], nil
}

func (transit *Convey) PutAll() bool {
	if !transit.takeValues() {
		transit.PutError("can't put all values from row")
		return false
	}
	for name, value := range transit.values {
		transit.Set(name, value)
	}
	return true
}

func (transit *Convey) PutAs(column, as string) bool {
	if !transit.takeValues() {
		transit.PutError("can't put column", column, "as", as)
		return false
	}
	transit.Set(as, transit.values[column])
	return true
}

func (transit *Convey) Put(column string) bool {
	return transit.PutAs(column, column)
}
