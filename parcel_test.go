package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // используем in-memory БД для тестов
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// get
	storedParcel, err := store.Get(id)
	require.NoError(t, err)

	// Устанавливаем номер посылки перед сравнением, так как он задается БД
	parcel.Number = id
	require.Equal(t, parcel, storedParcel)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)

	// try getting deleted parcel
	_, err = store.Get(id)
	require.Error(t, err)                  // ожидаем ошибку, так как посылка должна быть удалена
	require.ErrorIs(t, err, sql.ErrNoRows) // корректный способ проверки ошибки
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // использование in-memory базы для тестов
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	parcel.Number = id

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, storedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // использование in-memory базы для тестов
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	parcel.Number = id

	// set status
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// check
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, storedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // используем in-memory БД
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	// задаем всем посылкам один и тот же client ID
	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// добавляем в базу
	for i := range parcels {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)

		// обновляем номер посылки перед сравнением
		parcels[i].Number = id
	}

	// получаем из базы
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)

	// проверяем, что количество совпадает
	require.Len(t, storedParcels, len(parcels))

	// проверяем, что все посылки из БД совпадают с оригинальными
	require.ElementsMatch(t, parcels, storedParcels)
}
