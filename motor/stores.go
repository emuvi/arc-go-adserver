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

type Fetcher struct {
	As     string
	Column string
	Form   *Formatter
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
		fmt.Println("You've tried to get the rows whitout make a query on:")
		debug.PrintStack()

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

func (transit *Convey) checkFetchers(fetchers ...Fetcher) bool {
	for idx, fetcher := range fetchers {
		if fetcher.Column == "" {
			if len(fetchers) > 1 {
				transit.PutError("the fetcher number", idx+1, "has a empty column")
			} else {
				transit.PutError("the fetcher has a empty column")
			}
			return false
		}
		if fetcher.As == "" {
			fetcher.As = fetcher.Column
		}
	}
	return true
}

func (transit *Convey) Put(fetcher Fetcher) bool {
	if !transit.checkFetchers(fetcher) {
		transit.PutError("can't put column", fetcher.Column)
		return false
	}
	if !transit.takeValues() {
		transit.PutError("can't put column", fetcher.Column)
		return false
	}	
	value, found := transit.values[fetcher.Column]
	if !found {
		transit.PutError("there's no column with name", fetcher.Column)
		transit.PutError("can't put column", fetcher.Column)
		return false
	}
	if fetcher.Form == nil {
		transit.Set(fetcher.As, value)
	} else {
		transit.Set(fetcher.As, fetcher.Form.Format(transit, value))
	}
	return true
}

func (transit *Convey) PutAll(fetchers ...Fetcher) bool {
	if !transit.checkFetchers(fetchers) {
		transit.PutError("can't put all values from the row")
		return false
	}
	if transit.Next() {
		if !transit.takeValues() {
			transit.PutError("can't put all values from the row")
			return false
		}
		for _, fetcher := range fetchers {
			value, found := transit.values[fetcher.Column]
			if !found {
				transit.PutError("there's no column with name", fetcher.Column)
				transit.PutError("can't put all values from the row")
				return false
			}
			if fetcher.Form == nil {
				transit.Set(fetcher.As, value)
			} else {
				transit.Set(fetcher.As, fetcher.Form.Format(transit, value))
			}
		}
	}
	if transit.HasError() {
		transit.PutError("can't put all values from the row")
		return false
	}
	return true
}

func (transit *Convey) PutRows(as string, fetchers ...Fetcher) bool {
	if !transit.checkFetchers(fetchers) {
		goto BadError
	}
	results := []interface{}{}
	for transit.Next() {
		if !transit.takeValues() {
			goto BadError
		}
		result := map[string]interface{}{}
		for _, fetcher := range fetchers {
			value, found := transit.values[fetcher.Column]
			if !found {
				transit.PutError("there's no column with name", fetcher.Column)
				goto BadError
			}
			if fetcher.Form == nil {
				result[fetcher.As] = value
			} else {
				result[fetcher.As] = fetcher.Form.Format(transit, value)
			}
		}
		results = append(results, result)
	}
	if transit.HasError() {
		goto BadError
	}
	transit.Set(as, results)
	return true
BadError:
	transit.PutError("can't put all values from all rows")
	return false
}
