package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	query := `INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)`
	res, err := s.db.Exec(query, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("Add: ошибка вставки: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Add: не удалось получить ID: %w", err)
	}
	// верните идентификатор последней добавленной записи
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	query := `SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`
	row := s.db.QueryRow(query, number)

	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return p, fmt.Errorf("Get: посылка с номером %d не найдена", number)
	}
	if err != nil {
		return p, fmt.Errorf("Get: ошибка сканирования строки: %w", err)
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	query := `SELECT number, client, status, address, created_at FROM parcel WHERE client = ?`
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, fmt.Errorf("GetByClient: ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	// заполните срез Parcel данными из таблицы
	var res []Parcel
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetByClient: ошибка сканирования строки: %w", err)
		}
		res = append(res, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetByClient: ошибка при чтении строк: %w", err)
	}
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	query := `UPDATE parcel SET status = ? WHERE number = ?`
	res, err := s.db.Exec(query, status, number)
	if err != nil {
		return fmt.Errorf("SetStatus: ошибка обновления: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("SetStatus: ошибка проверки затронутых строк: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("SetStatus: посылка с номером %d не найдена", number)
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	var currentStatus string
	err := s.db.QueryRow(
		`SELECT status FROM parcel WHERE number = ?`,
		number,
	).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		return fmt.Errorf("SetAddress: посылка с номером %d не найдена", number)
	}
	if err != nil {
		return fmt.Errorf("SetAddress: ошибка получения статуса: %w", err)
	}
	if currentStatus != "registered" {
		return fmt.Errorf(
			"SetAddress: нельзя менять адрес для посылки со статусом %q (разрешён только для registered)",
			currentStatus,
		)
	}
	res, err := s.db.Exec(
		`UPDATE parcel SET address = ? WHERE number = ?`,
		address,
		number,
	)
	if err != nil {
		return fmt.Errorf("SetAddress: ошибка обновления адреса: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("SetAddress: ошибка проверки затронутых строк: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("SetAddress: не удалось обновить адрес, строка не затронута")
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	var currentStatus string
	err := s.db.QueryRow(
		`SELECT status FROM parcel WHERE number = ?`,
		number,
	).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		return fmt.Errorf("Delete: посылка с номером %d не найдена", number)
	}
	if err != nil {
		return fmt.Errorf("Delete: ошибка получения статуса: %w", err)
	}
	if currentStatus != "registered" {
		return fmt.Errorf(
			"Delete: нельзя удалить посылку со статусом %q (разрешено только для registered)",
			currentStatus,
		)
	}
	res, err := s.db.Exec(
		`DELETE FROM parcel WHERE number = ?`,
		number,
	)
	if err != nil {
		return fmt.Errorf("Delete: ошибка удаления: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("Delete: ошибка проверки затронутых строк: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("Delete: не удалось удалить посылку, строка не затронута")
	}
	return nil
}
