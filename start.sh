#!/bin/bash

set -e # если где-то будет ошибка — скрипт остановится

echo "🚀 Запуск Collector..."

echo "🚀 Проверка Docker..."
if ! docker info > /dev/null 2>&1; then
  echo "❌ Docker не запущен. Попробуй запустить его вручную."
  exit 1
fi

echo "🐳 Запуск Docker-контейнеров..."
docker-compose up -d  # или docker run ... (если без compose)

echo "🟢 Контейнеры запущены"

echo "🧠 Запуск Go-сервера..."
go run main.go
