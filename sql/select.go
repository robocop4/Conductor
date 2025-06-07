package sql

import "database/sql"

type App struct {
	DB *sql.DB
}

func (a *App) getPods(hash string) (GetPodsStruct, error) {
	return SQLgetPods(a.DB, hash)
}
