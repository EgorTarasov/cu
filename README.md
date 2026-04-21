# CU - Central University CLI Tool

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**CU** - это инструмент командной строки для взаимодействия с API Central University. Утилита позволяет получать информацию о курсах, управлять аутентификацией и синхронизировать данные.

## Установка

### Из исходного кода

```bash
git clone <repository-url>
cd cu_sync
go build -o cu ./cmd/cli
```

### Быстрый старт

```bash
# Переместите исполняемый файл в PATH (опционально)
sudo mv cu /usr/local/bin/

# Проверьте установку
cu --help
```

## Аутентификация

Для работы утилиты необходимо пройти аутентификацию. Есть два способа:

### Способ 1 — Авторизация через браузер (рекомендуется)

```bash
cu login
```

Команда откроет Chrome, перенаправит на страницу входа Keycloak. После успешного логина cookie будет автоматически сохранён в `~/.cu-cli/cookie` и подхвачен всеми последующими командами.

Если cookie истёк — просто выполните `cu login` ещё раз.

Дополнительные флаги:

| Флаг        | По умолчанию | Описание                          |
| ----------- | ------------ | --------------------------------- |
| `--timeout` | `5m`         | Таймаут ожидания завершения входа |

> **Требование:** необходим установленный Google Chrome или Chromium. Если Chrome установлен в нестандартном месте, укажите путь через переменную `CHROME_PATH`.

### Способ 2 — Ручная установка cookie

1. Откройте браузер и войдите в [Central University](https://my.centraluniversity.ru)
2. Откройте Developer Tools (F12) -> вкладка **Network**
3. Обновите страницу и найдите любой запрос к API
4. Скопируйте значение `bff.cookie` из заголовков

```bash
export CU_BFF_COOKIE="ваше-значение-cookie"
```

> Переменная окружения `CU_BFF_COOKIE` имеет приоритет над сохранённым файлом.

## Использование

### Основные команды

```bash
# Показать справку
cu --help

# Авторизация через браузер
cu login

# Список всех доступных курсов
cu fetch courses

# Получить детальную информацию о курсе
cu fetch course 519

# Скачать материалы курса
cu fetch course 519 --dump --path ./materials
```

### Примеры вывода

#### Список курсов

```bash
$ cu fetch courses

Fetching Student Courses
===========================

Successfully fetched 5 courses!
Total available: 5 courses

1. Case Evenings (ID: 519)
   State: published | Archived: false
   Published: 2025-09-01 07:00:18

2. Java Core (ID: 526)
   State: published | Archived: false
   Published: 2025-09-01 07:00:19

...
```

#### Детальная информация о курсе

```bash
$ cu fetch course 519

Fetching Course Overview
===========================

Course fetched successfully!

Course: Case Evenings (ID: 519)
State: published
Archived: false
Publish Date: 2025-09-01 07:00:00
Themes: 4
  1. Силлабус (ID: 4399)
     Longreads: 1
       - Ссылка на силлабус (common)
  ...
```

## Структура команд

```
cu
├── login             # Авторизация через браузер
├── fetch             # Получение данных
│   ├── courses       # Список всех курсов студента
│   └── course [ID]   # Детальная информация о курсе
├── storage           # Управление хранилищем (в разработке)
│   ├── init         # Инициализация хранилища
│   ├── unseal       # Разблокировка хранилища
│   ├── seal         # Блокировка хранилища
│   ├── status       # Статус хранилища
│   └── clear        # Очистка хранилища
└── courses           # Синхронизация курсов (в разработке)
```

## Переменные окружения

| Переменная      | Описание                          | Обязательная |
| --------------- | --------------------------------- | ------------ |
| `CU_BFF_COOKIE` | Cookie аутентификации (приоритет над файлом) | Нет |
| `CHROME_PATH`   | Путь к Chrome/Chromium (если не стандартный)  | Нет |

## Устранение неполадок

### Cookie истёк (403 Forbidden)

```
Cookie validation failed: bff.cookie is invalid or expired: 403
```

**Решение:** Выполните `cu login` для получения нового cookie.

### Аутентификация не найдена

```
No authentication found.

Option 1 — login via browser:
  cu login

Option 2 — set cookie manually:
  export CU_BFF_COOKIE='your-cookie-value-here'
```

**Решение:** Выполните `cu login` или установите переменную окружения `CU_BFF_COOKIE`.

### Chrome не найден

```
Chrome not found. Install Google Chrome or set CHROME_PATH environment variable
```

**Решение:** Установите Google Chrome или укажите путь через `CHROME_PATH`.

### Ошибки сети

Убедитесь, что у вас есть доступ к интернету и серверы Central University доступны.

## Тестирование

```bash
# Запуск всех тестов
go test ./...

# Запуск тестов с подробным выводом
go test -v ./internal/cu
```

## Разработка

### Структура проекта

```
.
├── .github/workflows/ # GitHub Actions для CI/CD
├── cmd/cli/           # Основное приложение CLI
├── internal/
│   ├── cli/          # Команды CLI (Cobra)
│   └── cu/           # Клиент API Central University
├── integration_tests/ # Интеграционные тесты
├── build/            # Собранные бинарные файлы
├── Makefile          # Задачи для разработки
├── go.mod
├── go.sum
└── README.md
```

### Разработка с Makefile

```bash
# Показать все доступные команды
make help

# Запустить тесты
make test

# Собрать для текущей платформы
make build

# Собрать для всех платформ
make build-all

# Запустить приложение
make run ARGS='fetch courses'
```

### Ручная сборка

```bash
# Сборка для текущей платформы
go build -o cu ./cmd/cli

# Кросс-компиляция
GOOS=windows GOARCH=amd64 go build -o cu.exe ./cmd/cli
GOOS=linux GOARCH=amd64 go build -o cu-linux ./cmd/cli
GOOS=darwin GOARCH=amd64 go build -o cu-macos ./cmd/cli
```

### CI/CD

Проект использует GitHub Actions:

- **test.yml** — тесты на Go 1.25, покрытие кода, `go vet`
- **build.yml** — сборка для 6 платформ, контрольные суммы SHA256, релизы по тегам
- **pr.yml** — форматирование, тесты, Gosec, golangci-lint

## Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для фичи (`git checkout -b feature/amazing-feature`)
3. Сделайте коммит изменений (`git commit -m 'Add amazing feature'`)
4. Запушьте ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## Лицензия

Этот проект распространяется под лицензией MIT. См. файл [LICENSE](LICENSE) для деталей.

## Roadmap

- [x] Получение списка курсов
- [x] Детальная информация о курсе
- [x] Валидация аутентификации
- [x] Авторизация через браузер (chromedp)
- [ ] Синхронизация курсов
- [ ] Экспорт данных в различные форматы
- [ ] Интерактивный режим
- [ ] Конфигурационные файлы
