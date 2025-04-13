package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // Для подключения SQLite драйвера
)

// Инициализация случайного генератора для тестов
var (
	randSource = rand.NewSource(time.Now().UnixNano()) // Источник случайных чисел
	randRange  = rand.New(randSource)                  // Генератор случайных чисел
)

// Функция для создания тестовой посылки
// Возвращает посылку с тестовыми значениями для клиента, статуса, адреса и времени создания
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,                                  // Тестовый клиент
		Status:    ParcelStatusRegistered,                // Статус "зарегистрирован"
		Address:   "test",                                // Тестовый адрес
		CreatedAt: time.Now().UTC().Format(time.RFC3339), // Время создания в формате RFC3339
	}
}

// Функция для настройки тестовой базы данных
// Открывает SQLite базу в памяти, создает таблицу и возвращает соединение
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:") // Открытие базы данных в памяти
	require.NoError(t, err)                   // Проверка на ошибку при открытии

	// Создание таблицы "parcel", если она еще не существует
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS parcel (
			number INTEGER PRIMARY KEY, 
			client INTEGER,
			status TEXT, 
			address TEXT, 
			created_at TEXT
		);
	`)
	require.NoError(t, err) // Проверка на ошибку при создании таблицы

	return db // Возвращаем соединение с базой данных
}

// Тест на добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	db := setupTestDB(t)        // Настройка тестовой базы данных
	store := NewParcelStore(db) // Создание нового хранилища посылок
	parcel := getTestParcel()   // Получение тестовой посылки

	// Добавление посылки в базу данных
	id, err := store.Add(parcel)
	require.NoError(t, err) // Проверка на ошибку
	require.NotZero(t, id)  // Проверка, что ID не равен нулю

	// Получение посылки из базы данных
	storedParcel, err := store.Get(id)
	require.NoError(t, err) // Проверка на ошибку
	// Проверка, что данные посылки совпадают с добавленной посылкой
	require.Equal(t, parcel.Client, storedParcel.Client)
	require.Equal(t, parcel.Status, storedParcel.Status)
	require.Equal(t, parcel.Address, storedParcel.Address)
	require.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt)

	// Удаление посылки
	err = store.Delete(id)
	require.NoError(t, err) // Проверка на ошибку при удалении

	// Проверка, что посылка была удалена
	_, err = store.Get(id)
	require.Error(t, err) // Ожидаем ошибку при попытке получить удаленную посылку
}

// Тест на изменение адреса посылки
func TestSetAddress(t *testing.T) {
	db := setupTestDB(t)
	store := NewParcelStore(db)

	// Добавление посылки в базу данных
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Изменение адреса посылки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err) // Проверка на ошибку при изменении адреса

	// Проверка, что адрес был изменен
	updatedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, updatedParcel.Address) // Проверка, что новый адрес совпадает
}

// Тест на изменение статуса посылки
func TestSetStatus(t *testing.T) {
	db := setupTestDB(t)
	store := NewParcelStore(db)

	// Добавление посылки в базу данных
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Изменение статуса посылки
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err) // Проверка на ошибку при изменении статуса

	// Проверка, что статус был изменен
	updatedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, updatedParcel.Status) // Проверка, что новый статус совпадает
}

// Тест на получение посылок по клиенту
func TestGetByClient(t *testing.T) {
	db := setupTestDB(t)
	store := NewParcelStore(db)

	// Создание нескольких тестовых посылок
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{} // Словарь для сопоставления ID и посылок

	// Генерация случайного идентификатора клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// Добавление посылок в базу данных
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)    // Проверка на ошибку при добавлении
		parcels[i].Number = id     // Сохранение ID посылки
		parcelMap[id] = parcels[i] // Добавление в словарь для проверки
	}

	// Получение посылок клиента из базы данных
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)                            // Проверка на ошибку при получении посылок
	require.Equal(t, len(parcels), len(storedParcels)) // Проверка, что количество полученных посылок совпадает с добавленными

	// Проверка, что все полученные посылки соответствуют добавленным
	for _, parcel := range storedParcels {
		expected, exists := parcelMap[parcel.Number]
		require.True(t, exists) // Проверка, что посылка была добавлена
		// Проверка, что данные посылки совпадают с добавленными
		require.Equal(t, expected.Client, parcel.Client)
		require.Equal(t, expected.Status, parcel.Status)
		require.Equal(t, expected.Address, parcel.Address)
		require.Equal(t, expected.CreatedAt, parcel.CreatedAt)
	}
}
