# CU - Central University CLI Tool

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**CU** - это инструмент командной строки для взаимодействия с API Central University. Утилита позволяет получать информацию о курсах, управлять аутентификацией и синхронизировать данные.

## Установка

### Одной командой (macOS, Linux: Ubuntu/Fedora, amd64/arm64)

```bash
curl -fsSL https://raw.githubusercontent.com/EgorTarasov/cu/main/install.sh | sh
```

Одна команда делает всё: скачивает последний релиз с GitHub, проверяет sha256,
ставит бинарь в `~/.local/bin/cuni`, снимает quarantine-атрибут на macOS **и
прописывает PATH** в rc-файл вашей оболочки (`~/.zshrc`, `~/.bash_profile`,
`~/.bashrc` или `~/.config/fish/config.fish` — определяется по `$SHELL`).

После установки перезапустите терминал (или выполните `. ~/.zshrc`) — и всё.

PATH дописывается **в начало**, чтобы установленный бинарь выигрывал у
системных утилит с тем же именем. Блок помечен комментарием
`# added by cuni installer`, поэтому повторный запуск не плодит дубликаты.

Опции:

```bash
# Конкретная версия
curl -fsSL https://raw.githubusercontent.com/EgorTarasov/cu/main/install.sh | CU_VERSION=v0.1.5 sh

# Системная установка
curl -fsSL https://raw.githubusercontent.com/EgorTarasov/cu/main/install.sh | CU_INSTALL_DIR=/usr/local/bin sh

# Не трогать rc-файлы — скрипт просто напечатает нужную строку
curl -fsSL https://raw.githubusercontent.com/EgorTarasov/cu/main/install.sh | CU_NO_MODIFY_PATH=1 sh
```

| Переменная | Описание |
|------------|----------|
| `CU_VERSION` | Поставить конкретный тег (по умолчанию — последний релиз) |
| `CU_INSTALL_DIR` | Куда ставить (по умолчанию `~/.local/bin`) |
| `CU_NO_MODIFY_PATH` | `1` — не редактировать rc-файл, только напечатать строку |

### Бинарь вручную

Скачайте подходящий архив со страницы [Releases](https://github.com/EgorTarasov/cu/releases),
сверьте sha256 по `checksums.txt`, распакуйте `cuni` куда-то в PATH (`~/.local/bin` или `/usr/local/bin`),
сделайте исполняемым.

### Через `go install`

Если в системе уже есть Go ≥ 1.26:

```bash
go install github.com/EgorTarasov/cu/cmd/cuni@latest
```

Бинарь окажется в `$(go env GOPATH)/bin/cuni`.

### Из исходного кода

```bash
git clone https://github.com/EgorTarasov/cu.git
cd cu
make install         # собирает в bin/cuni и копирует в $GOPATH/bin
# или просто:
make build           # bin/cuni
```

### Проверка

```bash
cuni --version
cuni --help
```

## Аутентификация

Для работы утилиты нужен cookie доступа к LMS. При первом запуске `cuni login`
спросит, каким способом вы хотите его получать, и запомнит выбор в
`~/.cu-cli/config.json`. Изменить — `cuni login --setup`.

### Способ 1 — вручную (по умолчанию)

Ничего дополнительно ставить не нужно. Подходит для серверов без графики.

```bash
cuni login --manual
```

1. Откройте [Central University](https://my.centraluniversity.ru) и войдите.
2. Откройте DevTools (<kbd>F12</kbd> или <kbd>Cmd</kbd>+<kbd>Option</kbd>+<kbd>I</kbd>).
3. Вкладка **Application** (Chrome/Edge) или **Storage** (Firefox).
4. Слева **Cookies** → `https://my.centraluniversity.ru`.
5. Скопируйте значение `bff.cookie` из колонки **Value** и вставьте в запрос cu.

Ввод скрыт, чтобы cookie не остался в истории терминала. cuni проверит его
через API и сохранит в `~/.cu-cli/cookie` (права `0600`).

### Способ 2 — через браузер

```bash
cuni login --browser
```

cuni откроет Chrome, вы войдёте через Keycloak SSO, cookie будет сохранён
автоматически.

| Флаг        | По умолчанию | Описание                          |
| ----------- | ------------ | --------------------------------- |
| `--timeout` | `5m`         | Таймаут ожидания завершения входа |

> **Требование:** нужен Google Chrome или Chromium. cuni ищет его в `CHROME_PATH`,
> стандартных путях установки и `PATH`. Если Chrome не найден — cuni предложит
> скачать Chrome for Testing (~170 МБ) в `~/.cu-cli/chrome`; эта сборка
> используется только для входа и не трогает ваш обычный браузер.

### Переменная окружения

```bash
export CU_BFF_COOKIE="ваше-значение-cookie"
```

> `CU_BFF_COOKIE` имеет приоритет над сохранённым файлом. В CI это позволяет
> обойтись без `cuni login` вовсе.

Если cookie истёк — выполните `cuni login` ещё раз.
Подробности: [docs/login.md](docs/login.md).

## Команды

| Команда | Описание | Документация |
|---------|----------|--------------|
| `cuni login` | Аутентификация (вручную или через браузер) | [docs/login.md](docs/login.md) |
| `cuni courses` | Список курсов с ID | [docs/courses.md](docs/courses.md) |
| `cuni deadlines [course]` | Ближайшие дедлайны | [docs/deadlines.md](docs/deadlines.md) |
| `cuni grades [course]` | Оценки и ведомость | [docs/grades.md](docs/grades.md) |
| `cuni materials <course>` | Скачать PDF и ссылки на материалы | [docs/materials.md](docs/materials.md) |
| `cuni task <id>` | Детали задания | [docs/task.md](docs/task.md) |

Курс можно указать по **ID** (`901`) или по **названию** (`go`, `sql`, `алгоритмы`) — поиск регистронезависимый.

### Дополнительные команды fetch

`fetch` помогает получить данные напрямую из API:

- `cuni fetch student` — профиль текущего студента
- `cuni fetch course-summary <course-id>` — краткая информация о курсе
- `cuni fetch theme <theme-id>` — краткая информация о теме
- `cuni fetch longread <longread-id>` — краткая информация о лонгриде

### Обновление MCP-сервера

После `make install` запущенные хосты (Claude Code, IDE) продолжают держать
старый процесс — новые тулы и фиксы не подхватятся. После переустановки
перезапустите сервер в хосте:

```bash
claude mcp restart cuni   # или /mcp reconnect в Claude Code
```

### Быстрый пример

```bash
# Что горит?
cuni deadlines

# Оценки по Go
cuni grades go

# Скачать все лекции и семинары
cuni materials алгоритмы --path ./downloads

# Подробности по заданию
cuni task 1536681
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

**Решение:** Выполните `cuni login` для получения нового cookie.

### Аутентификация не найдена

```
No authentication found.

Option 1 — login via browser:
  cuni login

Option 2 — set cookie manually:
  export CU_BFF_COOKIE='your-cookie-value-here'
```

**Решение:** Выполните `cuni login` или установите переменную окружения `CU_BFF_COOKIE`.

### Chrome не найден

```
Chrome not found. Install Google Chrome or set CHROME_PATH environment variable
```

**Решение:** Установите Google Chrome, укажите путь через `CHROME_PATH`,
согласитесь на загрузку Chrome for Testing или используйте `cuni login --manual`.

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
├── cmd/cuni/         # Основное приложение CLI
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
go build -o cuni ./cmd/cuni

# Кросс-компиляция
GOOS=windows GOARCH=amd64 go build -o cuni.exe ./cmd/cuni
GOOS=linux GOARCH=amd64 go build -o cuni-linux ./cmd/cuni
GOOS=darwin GOARCH=amd64 go build -o cuni-macos ./cmd/cuni
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

- [x] Авторизация через браузер (chromedp)
- [x] Получение списка курсов
- [x] Скачивание PDF-материалов
- [x] Просмотр дедлайнов
- [x] Просмотр оценок и ведомости
- [x] Поиск курса по названию
- [ ] Скачивание лонгридов с GitLab
- [ ] Интерактивный режим
- [ ] Уведомления о приближающихся дедлайнах
