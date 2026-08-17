package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println("Ошибка открытия БД:", err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	if err != nil {
		fmt.Println("Ошибка при добавлении посылки:", err)
		return
	}
	if id <= 0 {
		fmt.Println("Неверный ID после добавления:", id)
		return
	}
	parcel.Number = id
	fmt.Printf("Добавлено: №%d\n", id)

	// get
	getParcel, err := store.Get(id)
	if err != nil {
		fmt.Println("Ошибка при получении посылки:", err)
		return
	}

	if getParcel.Number != parcel.Number ||
		getParcel.Client != parcel.Client ||
		getParcel.Status != parcel.Status ||
		getParcel.Address != parcel.Address {
		fmt.Println("Поля не совпадают!")
		fmt.Printf("Ожидалось: %+v\nПолучено: %+v\n", parcel, getParcel)
		return
	}
	fmt.Printf("Получено: №%d, адрес %s, статус %s\n", getParcel.Number, getParcel.Address, getParcel.Status)

	// delete
	err = store.Delete(id)
	if err != nil {
		fmt.Println("Ошибка при удалении посылки:", err)
		return
	}
	fmt.Printf("Удалено: №%d\n", id)

	_, err = store.Get(id)
	if err == nil {
		fmt.Println("ОШИБКА: после удаления посылка всё ещё находится!")
		return
	}
	fmt.Printf("Проверка удаления: корректно вернула ошибку: %v\n", err)

	fmt.Println("TestAddGetDelete: OK")
}
func TestSetAddress(t *testing.T) {
	// prepare

	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println("Ошибка открытия БД:", err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)

	// add
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	if err != nil {
		fmt.Println("Ошибка при добавлении посылки:", err)
		return
	}
	if id <= 0 {
		fmt.Println("Неверный ID после добавления:", id)
		return
	}
	parcel.Number = id
	fmt.Printf("Добавлено: №%d\n", id)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	if err != nil {
		fmt.Println("Ошибка при обновлении адреса:", err)
		return
	}
	fmt.Printf("Адрес обновлён на: %s\n", newAddress)

	// check
	updatedParcel, err := store.Get(id)
	if err != nil {
		fmt.Println("Ошибка при получении посылки после обновления:", err)
		return
	}

	if updatedParcel.Address != newAddress {
		fmt.Println("ОШИБКА: адрес не обновился!")
		fmt.Printf("Ожидалось: %s\nПолучено: %s\n", newAddress, updatedParcel.Address)
		return
	}

	fmt.Println("TestSetAddress: OK")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println("Ошибка открытия БД:", err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)

	// add
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	if err != nil {
		fmt.Println("Ошибка при добавлении посылки:", err)
		return
	}
	if id <= 0 {
		fmt.Println("Неверный ID после добавления:", id)
		return
	}
	parcel.Number = id
	fmt.Printf("Добавлено: №%d\n", id)

	// set status
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	if err != nil {
		fmt.Println("Ошибка при обновлении статуса:", err)
		return
	}
	fmt.Printf("Статус обновлён на: %v\n", newStatus)

	// check
	updateParcel, err := store.Get(id)
	if err != nil {
		fmt.Println("Ошибка при получении посылки после обновления:", err)
		return
	}
	if updateParcel.Status != newStatus {
		fmt.Println("ОШИБКА: статус не обновился!")
		fmt.Printf("Ожидалось: %v\nПолучено: %v\n", newStatus, updateParcel.Status)
		return
	}
	fmt.Println("TestSetStatus: OK (статус успешно обновлён)")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println("Ошибка открытия БД:", err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		if err != nil {
			fmt.Printf("Ошибка при добавлении посылки №%d: %v\n", i, err)
			return
		}
		if id <= 0 {
			fmt.Printf("Неверный ID для посылки №%d: %d\n", i, id)
			return
		}

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}
	fmt.Printf("Добавлено %d посылок для клиента %d\n", len(parcels), client)

	// get by client
	storedParcels, err := store.GetByClient(client)
	if err != nil {
		fmt.Println("Ошибка при получении посылок по клиенту:", err)
		return
	}
	if len(storedParcels) != len(parcels) {
		fmt.Printf("Количество не совпадает! Ожидалось: %d, получено: %d\n",
			len(parcels), len(storedParcels))
		return
	}
	fmt.Printf("Получено %d посылок — количество верное\n", len(storedParcels))

	// check
	for _, parcel := range storedParcels {
		original, ok := parcelMap[parcel.Number]
		if !ok {
			fmt.Printf("Посылка с ID=%d не найдена в карте добавленных\n", p.Number)
			return
		}

		if parcel.Client != original.Client ||
			parcel.Status != original.Status ||
			parcel.Address != original.Address ||
			parcel.CreatedAt != original.CreatedAt {
			fmt.Printf("Поля не совпадают для посылки ID=%d\n", parcel.Number)
			fmt.Printf("  Ожидалось: Client=%d, Status=%v, Address=%s, CreatedAt=%s\n",
				original.Client, original.Status, original.Address, original.CreatedAt)
			fmt.Printf("  Получено:  Client=%d, Status=%v, Address=%s, CreatedAt=%s\n",
				parcel.Client, parcel.Status, parcel.Address, parcel.CreatedAt)
			return
			// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
			// убедитесь, что все посылки из storedParcels есть в parcelMap
			// убедитесь, что значения полей полученных посылок заполнены верно
		}
	}
	fmt.Println("TestGetByClient: OK (все посылки найдены и поля совпадают)")
}
