package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // Подключаем драйвер для SQLite
)

// Статусы посылок
const (
	ParcelStatusRegistered = "registered" // Статус "Зарегистрирована"
	ParcelStatusSent       = "sent"       // Статус "Отправлена"
	ParcelStatusDelivered  = "delivered"  // Статус "Доставлена"
)

// Структура, представляющая посылку
type Parcel struct {
	Number    int    // Номер посылки
	Client    int    // Идентификатор клиента
	Status    string // Статус посылки
	Address   string // Адрес доставки
	CreatedAt string // Время создания посылки
}

// Сервис для работы с посылками
type ParcelService struct {
	store ParcelStore // Хранилище, которое взаимодействует с базой данных
}

// Функция-конструктор для создания нового сервиса ParcelService
func NewParcelService(store ParcelStore) ParcelService {
	return ParcelService{store: store} // Инициализация сервиса с переданным хранилищем
}

// Регистрация новой посылки
func (s ParcelService) Register(client int, address string) (Parcel, error) {
	// Создание новой посылки
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered, // Статус "Зарегистрирована"
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339), // Текущее время в формате RFC3339
	}

	// Добавляем посылку в хранилище (базу данных)
	id, err := s.store.Add(parcel)
	if err != nil {
		return parcel, err // Если ошибка при добавлении, возвращаем ошибку
	}

	// Присваиваем посылке полученный номер (ID)
	parcel.Number = id

	// Логируем информацию о зарегистрированной посылке
	fmt.Printf("Новая посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt)

	// Возвращаем успешно зарегистрированную посылку
	return parcel, nil
}

// Вывод всех посылок клиента
func (s ParcelService) PrintClientParcels(client int) error {
	// Получаем все посылки клиента из хранилища
	parcels, err := s.store.GetByClient(client)
	if err != nil {
		return err // Возвращаем ошибку, если не удалось получить посылки
	}

	// Выводим информацию о посылках клиента
	fmt.Printf("Посылки клиента %d:\n", client)
	for _, parcel := range parcels {
		// Печатаем информацию о каждой посылке
		fmt.Printf("Посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt, parcel.Status)
	}
	fmt.Println()

	return nil
}

// Обновление статуса посылки
func (s ParcelService) NextStatus(number int) error {
	// Получаем посылку по её номеру
	parcel, err := s.store.Get(number)
	if err != nil {
		return err // Возвращаем ошибку, если посылка не найдена
	}

	// Определяем следующий статус в зависимости от текущего
	var nextStatus string
	switch parcel.Status {
	case ParcelStatusRegistered:
		nextStatus = ParcelStatusSent // Если зарегистрирована, статус меняем на "Отправлена"
	case ParcelStatusSent:
		nextStatus = ParcelStatusDelivered // Если отправлена, статус меняем на "Доставлена"
	case ParcelStatusDelivered:
		return nil // Если уже доставлена, то статус менять не нужно
	}

	// Логируем изменение статуса посылки
	fmt.Printf("У посылки № %d новый статус: %s\n", number, nextStatus)

	// Обновляем статус в хранилище
	return s.store.SetStatus(number, nextStatus)
}

// Изменение адреса доставки посылки
func (s ParcelService) ChangeAddress(number int, address string) error {
	// Обновляем адрес доставки посылки в хранилище
	return s.store.SetAddress(number, address)
}

// Удаление посылки
func (s ParcelService) Delete(number int) error {
	// Удаляем посылку из хранилища
	return s.store.Delete(number)
}

func main() {
	// Настройка подключения к SQLite базе данных
	db, err := sql.Open("sqlite", "tracker.db") // Открываем соединение с базой данных SQLite
	if err != nil {
		fmt.Println("Ошибка при подключении к базе данных:", err)
		return
	}
	defer db.Close() // Обеспечиваем закрытие соединения с базой данных при завершении программы

	// Проверяем соединение с базой данных
	err = db.Ping()
	if err != nil {
		fmt.Println("Ошибка при проверке соединения с базой данных:", err)
		return
	}

	// Создаем таблицу parcel, если она не существует
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS parcel (
            number INTEGER PRIMARY KEY,
            client INTEGER,                           
            status TEXT,                              
            address TEXT,                             
            created_at TEXT                           
        );
    `)
	if err != nil {
		fmt.Println("Ошибка при создании таблицы:", err)
		return
	}

	// Создаем объект ParcelStore, который будет взаимодействовать с базой данных
	store := NewParcelStore(db)
	service := NewParcelService(store) // Создаем сервис для работы с посылками

	// Регистрация посылки
	client := 1
	address := "Псков, д. Пушкина, ул. Колотушкина, д. 5"
	p, err := service.Register(client, address)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Изменение адреса посылки
	newAddress := "Саратов, д. Верхние Зори, ул. Козлова, д. 25"
	err = service.ChangeAddress(p.Number, newAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Обновление статуса посылки
	err = service.NextStatus(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Вывод всех посылок клиента
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Попытка удаления отправленной посылки
	err = service.Delete(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Повторный вывод всех посылок клиента
	// Примечание: предыдущая посылка не должна быть удалена, т.к. её статус НЕ "зарегистрирована"
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Регистрация новой посылки
	p, err = service.Register(client, address)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Удаление новой посылки
	err = service.Delete(p.Number)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Вывод всех посылок клиента
	// Новая посылка должна быть удалена, и её не будет в списке
	err = service.PrintClientParcels(client)
	if err != nil {
		fmt.Println(err)
		return
	}
}
