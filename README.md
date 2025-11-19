# Order Service

Микросервис для обработки заказов с использованием Go, PostgreSQL, Kafka и in-memory кэша.

## 🏗️ Архитектура
[Producer] → [Kafka] → [Consumer] → [PostgreSQL] → [Cache] → [HTTP API] → [Web Interface]

## 🚀 Быстрый старт

### 1. Запуск 
```bash
#PostgreSQL
brew services start postgresql

#Kafka + Zookeeper
brew services start zookeeper
brew services start kafka
sleep 10

#Создание топика Kafka
kafka-topics --create --topic orders --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1

#Создание БД и пользователя
psql postgres -c "CREATE DATABASE order_service;"
psql postgres -c "CREATE USER order_user WITH PASSWORD 'your_password_here';"
psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE order_service TO order_user;"

#Применение миграций
psql -U order_user -d order_service -f migrations/001_create_tables.sql

#Установка зависимостей
go mod download

#Запуск сервера
export LOG_PRETTY=true
go run cmd/server/main.go

#Отправка тестового заказа
go run cmd/producer/main.go

#Доступ к веб-интерфейсу
open http://localhost:8081
```
### 2. Структура проекта
L0/
├── cmd/
│   ├── server/          #Основное приложение
│   └── producer/        #Эмулятор отправки заказов
├── internal/
│   ├── cache/           #In-memory кэш
│   ├── config/          #Конфигурация
│   ├── kafka/           #Kafka consumer
│   ├── models/          #Модели данных
│   ├── repository/      #Работа с БД
│   ├── validation/      #Валидация данных
│   └── interfaces/      #Интерфейсы
├── migrations/          #SQL миграции
├── web/static/          #Веб-интерфейс
└── .env.example         #Пример конфигурации



### 3. Стек
Go 
PostgreSQL
Kafka 
HTML/CSS/JS 