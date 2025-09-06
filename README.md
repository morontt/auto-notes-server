# Сервер

Сервер для автоблокнота. Проект для изучения gRPC

## Настройки приложения

```sh
cp config.dist.toml config.toml
```

## Создание ключа для подписи JWT-токена

```shell
openssl rand -base64 32
```

После сохранить в config.toml

## Генерация исходных файлов по .proto

```sh
go generate ./...
```
