# Algorithm Synchronization Service

Algorithm Synchronization Service — это специализированный сервис, предназначенный для управления и синхронизации Kubernetes pod, назначенных пользователям. Этот проект предоставляет API для добавления, обновления и удаления клиентов, обеспечивая эффективное взаимодействие и управление ресурсами в Kubernetes инфраструктуре.

## Доступ к Swagger

Документация API доступна по адресу:

http://HOST:PORT/docs

## Пример .env для запуска

Для запуска проекта убедитесь, что ваш файл `.env` содержит следующие переменные:

```.env
ENV= # local/dev/prod

POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_HOST=
POSTGRES_PORT=
DATABASE_NAME=

SERVER_HOST=
SERVER_PORT=
SERVER_TIMEOUT=

KUBECONFIG=
CONTAINER_IMAGE=
```
