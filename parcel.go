package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {

	query := `INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)`
	res, err := s.db.Exec(query, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("Add: ошибка вставки: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Add: не удалось получить ID: %w", err)
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {

	query := `SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`
	row := s.db.QueryRow(query, number)

	p := Parcel{}
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Parcel{}, fmt.Errorf("Get: посылка с номером %d не найдена: %w", number, err)
	}
	if err != nil {
		return Parcel{}, fmt.Errorf("Get: ошибка сканирования строки: %w", err)
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {

	query := `SELECT number, client, status, address, created_at FROM parcel WHERE client = ?`
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, fmt.Errorf("GetByClient: ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

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

	res, err := s.db.Exec(
		`UPDATE parcel SET address = ? WHERE number = ? AND status = ?`,
		address,
		number,
		ParcelStatusRegistered,
	)
	if err != nil {
		return fmt.Errorf("SetAddress: ошибка обновления адреса: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("SetAddress: ошибка проверки затронутых строк: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("SetAddress: не удалось обновить адрес. Посылка с номером %d либо не найдена, либо её статус не %s", number, ParcelStatusRegistered)
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {

	res, err := s.db.Exec(
		`DELETE FROM parcel WHERE number = ? AND status = ?`,
		number,
		ParcelStatusRegistered,
	)
	if err != nil {
		return fmt.Errorf("Delete: ошибка удаления: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("Delete: ошибка проверки затронутых строк: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("Delete: не удалось удалить посылку. Посылка с номером %d либо не найдена, либо её статус не %s", number, ParcelStatusRegistered)
	}
	return nil
}
