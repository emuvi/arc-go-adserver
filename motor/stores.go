package motor

import (
	"adserver/guide"
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/jackc/pgx/v4/pgxpool"
)

type aStore struct {
	pool   *pgxpool.Pool
	Client string
	User   string
}

type Fetcher struct {
	As     string
	Column string
	Form   *Style
}

func openStore(session *aSession, client, user, pass string) error {
	closeStore(session)
	storesHost := guide.Configs.GetString("StoreHost", "pointel.pointto.us")
	storesPort := guide.Configs.GetInt("StorePort", 5432)
	pool, err := pgxpool.Connect(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", user, pass, storesHost, storesPort, client))
	if err != nil {
		return err
	}
	session.mutex.Lock()
	defer session.mutex.Unlock()
	session.store = &aStore{pool: pool, Client: client, User: user}
	return nil
}

func closeStore(session *aSession) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	if session.store != nil && session.store.pool != nil {
		go session.store.pool.Close()
	}
	session.store = nil
}

func (transit *Convey) Open(client, user, pass string) bool {
	err := openStore(transit.Session(), client, user, pass)
	if err != nil {
		transit.PutError(err.Error())
		return false
	}
	return true
}

func (transit *Convey) Close() *Convey {
	transit.Session().closeStore()
	return transit
}

func (transit *Convey) Store() *aStore {
	result := transit.Session().store
	if result == nil {
		transit.PutError("can't find the store of the session")
		return nil
	}
	return result
}

func (transit *Convey) aquire() *pgxpool.Conn {
	if transit.link == nil {
		store := transit.Store()
		if store == nil {
			transit.PutError("can't aquire a link")
			return nil
		}
		result, err := store.pool.Acquire(context.Background())
		if err != nil {
			transit.PutError(err)
			transit.PutError("can't aquire a link")
			return nil
		}
		transit.link = result
	}
	return transit.link
}

func (transit *Convey) release() {
	if transit.link != nil {
		transit.link.Release()
	}
}

func (transit *Convey) Query(sql string, args ...interface{}) bool {
	transit.rows = nil
	transit.values = nil
	link := transit.aquire()
	if link == nil {
		transit.PutError("can't query the database")
		return false
	}
	transit.Done()
	rows, err := link.Query(context.Background(), sql, args...)
	if err != nil {
		transit.PutError(err.Error())
		transit.PutError("can't query the database")
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

func (transit *Convey) Done() {
	if transit.rows != nil {
		transit.rows.Close()
		transit.rows = nil
	}
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

func (transit *Convey) checkFetchers(fetchers ...*Fetcher) bool {
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

func (transit *Convey) Put(fetcher *Fetcher) bool {
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

func (transit *Convey) PutAll(fetchers ...*Fetcher) bool {
	if !transit.checkFetchers(fetchers...) {
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

func (transit *Convey) PutRows(as string, fetchers ...*Fetcher) bool {
	if !transit.checkFetchers(fetchers...) {
		transit.PutError("can't put all values from all rows")
		return false
	}
	results := []interface{}{}
	for transit.Next() {
		if !transit.takeValues() {
			transit.PutError("can't put all values from all rows")
			return false
		}
		result := map[string]interface{}{}
		for _, fetcher := range fetchers {
			value, found := transit.values[fetcher.Column]
			if !found {
				transit.PutError("there's no column with name", fetcher.Column)
				transit.PutError("can't put all values from all rows")
				return false
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
		transit.PutError("can't put all values from all rows")
		return false
	}
	transit.Set(as, results)
	return true
}
