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

// добавление данных по долгу у клиентв
func (r *DataClientsPostgres) AddDebtByClient(client *app.User) error {

	query := fmt.Sprintf("ON CONFLICT (t2.first_name, t2.last_name, t2.middle_name) DO UPDATE SET debt = EXCLUDED.debt", DBauth, DBusers) // отчест во необязательно!!! OR first_name=$2 AND last_name=$3
	_, err := r.db.Exec(query, client.Debt, client.FirstName, client.LastName, client.MiddleName)
	if err != nil {
		return errors.New("DataClientsPostgres/AddDebtByClient()/ошибка при добавлении данных по клиенту в БД: " + err.Error())
	}
	return nil
}

// UPDATE %s AS t1 SET t1.debt = $1 FROM %s AS t2 WHERE t2.id=t1.id_user AND t2.first_name=$2 AND t2.last_name=$3 AND t2.middle_name=$4

// SELECT EXISTS (SELECT (id) FROM users WHERE first_name='Алексей' AND last_name='Комиссаров' AND middle_name='');

// SELECT (id) FROM users WHERE EXISTS (SELECT (id) FROM users WHERE first_name='Светлана' AND last_name='Симакова' AND middle_name='Владимировна');

// INSERT INTO debts (id_user, debt)
// VALUES (1, 100)
// ON CONFLICT (id_user) DO UPDATE SET debt = EXCLUDED.debt;
