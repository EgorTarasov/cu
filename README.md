# CU - Central University CLI Tool

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**CU** - это инструмент командной строки для взаимодействия с API Central University. Утилита позволяет получать информацию о курсах, управлять аутентификацией и синхронизировать данные.

## 🚀 Установка

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

## 🔧 Настройка

Для работы утилиты необходимо получить и установить cookie аутентификации:

### 1. Получение bff.cookie

1. Откройте браузер и войдите в [Central University](https://my.centraluniversity.ru)
2. Откройте Developer Tools (F12)
3. Перейдите на вкладку **Network**
4. Обновите страницу или перейдите к курсам
5. Найдите любой запрос к API и скопируйте значение `bff.cookie` из заголовков

### 2. Установка переменной окружения

```bash
# Linux/macOS
export CU_BFF_COOKIE="ваше-значение-cookie"

# Windows (PowerShell)
$env:CU_BFF_COOKIE="ваше-значение-cookie"

# Windows (CMD)
set CU_BFF_COOKIE=ваше-значение-cookie
```

Для постоянного сохранения добавьте в ваш shell profile:

```bash
# ~/.bashrc, ~/.zshrc, или ~/.profile
echo 'export CU_BFF_COOKIE="ваше-значение-cookie"' >> ~/.zshrc
source ~/.zshrc
```

## 📖 Использование

### Основные команды

```bash
# Показать справку
cu --help

# Список всех доступных курсов
cu fetch courses

# Получить детальную информацию о курсе
cu fetch course 519

# Показать справку по команде fetch
cu fetch --help
```

### Примеры использования

#### Получение списка курсов

```bash
$ cu fetch courses
📚 Fetching Student Courses
===========================

✅ Successfully fetched 5 courses!
📊 Total available: 5 courses

1. 📖 Case Evenings (Кейс-вечера) (ID: 519)
   📊 State: published | 📁 Archived: false
   📅 Published: 2025-09-01 07:00:18
   🎯 Skill Level: none (Enabled: false)

2. 📖 Java Core (ID: 526)
   📊 State: published | 📁 Archived: false
   📅 Published: 2025-09-01 07:00:19
   🎯 Skill Level: none (Enabled: false)

...
```

#### Детальная информация о курсе

```bash
$ cu fetch course 519
📚 Fetching Course Overview
===========================

✅ Course fetched successfully!

📖 Course: Case Evenings (Кейс-вечера) (ID: 519)
📊 State: published
📁 Archived: false
📅 Publish Date: 2025-09-01 07:00:00
🎯 Skill Level: none (Enabled: false)
📚 Themes: 4
  1. Силлабус (ID: 4399)
     📖 Longreads: 1
       - Ссылка на силлабус (common)
  ...
```

## 🗂️ Структура команд

```
cu
├── fetch              # Получение данных
│   ├── courses       # Список всех курсов студента
│   └── course [ID]   # Детальная информация о курсе
├── login             # Аутентификация (в разработке)
├── storage           # Управление хранилищем (в разработке)
│   ├── init         # Инициализация хранилища
│   ├── unseal       # Разблокировка хранилища
│   ├── seal         # Блокировка хранилища
│   ├── status       # Статус хранилища
│   └── clear        # Очистка хранилища
└── courses           # Синхронизация курсов (в разработке)
```

## ⚙️ Переменные окружения

| Переменная      | Описание                          | Обязательная |
| --------------- | --------------------------------- | ------------ |
| `CU_BFF_COOKIE` | Cookie аутентификации из браузера | ✅ Да        |

## 🔍 Устранение неполадок

### Ошибка 403 (Forbidden)

```
⚠️  Cookie validation failed: bff.cookie is invalid or expired: 403
The CU_BFF_COOKIE might be expired. Please update it.
```

**Решение:** Cookie истек. Получите новый cookie из браузера и обновите переменную окружения.

### Cookie не найден

```
⚠️  No CU_BFF_COOKIE environment variable found.
Please set the CU_BFF_COOKIE environment variable with your bff.cookie value:

Example:
  export CU_BFF_COOKIE='your-cookie-value-here'
  cu fetch courses
```

**Решение:** Установите переменную окружения `CU_BFF_COOKIE` с актуальным значением cookie.

### Ошибки сети

Убедитесь, что у вас есть доступ к интернету и серверы Central University доступны.

## 🧪 Тестирование

```bash
# Запуск всех тестов
go test ./...

# Запуск тестов с подробным выводом
go test -v ./internal/cu

# Запуск тестов для конкретного модуля
go test ./internal/auth
```

## 🏗️ Разработка

### Структура проекта

```
.
├── cmd/cli/           # Основное приложение CLI
├── internal/
│   ├── auth/         # Модуль аутентификации и хранения
│   └── cu/           # Клиент API Central University
├── requests/         # Примеры curl запросов
├── go.mod
├── go.sum
└── README.md
```

### Сборка

```bash
# Сборка для текущей платформы
go build -o cu ./cmd/cli

# Кросс-компиляция
GOOS=windows GOARCH=amd64 go build -o cu.exe ./cmd/cli
GOOS=linux GOARCH=amd64 go build -o cu-linux ./cmd/cli
GOOS=darwin GOARCH=amd64 go build -o cu-macos ./cmd/cli
```

## 🤝 Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для фичи (`git checkout -b feature/amazing-feature`)
3. Сделайте коммит изменений (`git commit -m 'Add amazing feature'`)
4. Запушьте ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 📄 Лицензия

Этот проект распространяется под лицензией MIT. См. файл [LICENSE](LICENSE) для деталей.

## 🆘 Поддержка

Если у вас есть вопросы или проблемы:

1. Проверьте раздел [Устранение неполадок](#-устранение-неполадок)
2. Создайте [Issue](../../issues) в репозитории
3. Проверьте актуальность cookie аутентификации

## 🗺️ Roadmap

-   [x] Получение списка курсов
-   [x] Детальная информация о курсе
-   [x] Валидация аутентификации
-   [ ] OIDC аутентификация
-   [ ] Синхронизация курсов
-   [ ] Экспорт данных в различные форматы
-   [ ] Интерактивный режим
-   [ ] Конфигурационные файлы

---

Made with ❤️ for Central University students
