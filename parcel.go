package main

import (
	"database/sql"
)

// Структура для работы с хранилищем посылок
// Хранилище использует базу данных для хранения и получения данных о посылках
type ParcelStore struct {
	db *sql.DB // Соединение с базой данных
}

// Функция-конструктор для создания нового хранилища посылок
func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db} // Инициализация структуры с переданным соединением с базой данных
}

// Добавление новой посылки в базу данных
func (s ParcelStore) Add(p Parcel) (int, error) {
	// Выполняем SQL запрос для добавления новой посылки в таблицу
	res, err := s.db.Exec(
		"INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt,
	)
	if err != nil {
		return 0, err // Если ошибка при выполнении запроса, возвращаем ошибку
	}

	// Получаем ID только что добавленной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err // Если ошибка при получении ID, возвращаем ошибку
	}

	// Возвращаем ID добавленной посылки
	return int(id), nil
}

// Получение посылки по номеру
func (s ParcelStore) Get(number int) (Parcel, error) {
	// Выполняем SQL запрос для получения посылки по номеру
	row := s.db.QueryRow(
		"SELECT number, client, status, address, created_at FROM parcel WHERE number = ?",
		number,
	)

	var p Parcel
	// Сканируем результат в структуру Parcel
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return p, err // Если ошибка при получении данных, возвращаем ошибку
	}

	// Возвращаем найденную посылку
	return p, nil
}

// Получение всех посылок клиента
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// Выполняем SQL запрос для получения всех посылок данного клиента
	rows, err := s.db.Query(
		"SELECT number, client, status, address, created_at FROM parcel WHERE client = ?",
		client,
	)
	if err != nil {
		return nil, err // Если ошибка при выполнении запроса, возвращаем ошибку
	}
	defer rows.Close() // Закрываем rows после завершения работы с ними

	var parcels []Parcel
	// Читаем каждую строку результата
	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err // Если ошибка при сканировании строки, возвращаем ошибку
		}
		parcels = append(parcels, p) // Добавляем посылку в список
	}

	// Проверка на ошибки после завершения чтения всех строк
	if err = rows.Err(); err != nil {
		return nil, err // Если ошибка в процессе итерации, возвращаем ошибку
	}

	// Возвращаем список всех посылок клиента
	return parcels, nil
}

// Обновление статуса посылки
func (s ParcelStore) SetStatus(number int, status string) error {
	// Выполняем SQL запрос для обновления статуса посылки
	_, err := s.db.Exec(
		"UPDATE parcel SET status = ? WHERE number = ?",
		status, number,
	)
	return err // Возвращаем ошибку, если она возникла
}

// Изменение адреса доставки посылки
func (s ParcelStore) SetAddress(number int, address string) error {
	// Выполняем SQL запрос для изменения адреса, если статус посылки "зарегистрирован"
	_, err := s.db.Exec(
		"UPDATE parcel SET address = ? WHERE number = ? AND status = ?",
		address, number, ParcelStatusRegistered,
	)
	return err // Возвращаем ошибку, если она возникла
}

// Удаление посылки
func (s ParcelStore) Delete(number int) error {
	// Выполняем SQL запрос для удаления посылки, если статус "зарегистрирован"
	_, err := s.db.Exec(
		"DELETE FROM parcel WHERE number = ? AND status = ?",
		number, ParcelStatusRegistered,
	)
	return err // Возвращаем ошибку, если она возникла
}
