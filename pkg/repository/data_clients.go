package repository

import (
	"errors"
	"fmt"

	app "github.com/AlexKomzzz/collectivity"
	"github.com/jmoiron/sqlx"
)

type DataClientsPostgres struct {
	db *sqlx.DB
}

func NewDataClientsPostgres(db *sqlx.DB) *DataClientsPostgres {
	return &DataClientsPostgres{
		db: db,
	}
}

// добавление данных по долгу у клиента
func (r *DataClientsPostgres) AddDebtByClient(client *app.User) error {

	query := fmt.Sprintf("INSERT INTO %s (username, first_name, last_name, middle_name, debt) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (username) DO UPDATE SET debt = EXCLUDED.debt ", DBusers)
	_, err := r.db.Exec(query, client.Username, client.FirstName, client.LastName, client.MiddleName, client.Debt)
	if err != nil {
		return errors.New("DataClientsPostgres/AddDebtByClient()/ошибка при добавлении данных по клиенту в БД: " + err.Error())
	}
	return nil
}
