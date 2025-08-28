# Lootor comments service

# Локальная установка

## Установка зависимостей
```shell
go mod download
```

## Установка компилятора protoс

```shell
brew install protobuf  ## Для macOS
sudo apt install protobuf-compiler ## Для Linux
choco install protoc ## Для Windows
```

## Генерация protobuf для связи с lootor_webapp

```shell
make gen
```

## Запуск

```shell
go run cmd/api/main.go
```

## Contributing
Хуютинг

## Лицензия

[MIT](https://choosealicense.com/licenses/mit/)