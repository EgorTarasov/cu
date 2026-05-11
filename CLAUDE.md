# CLAUDE.md

Принципы и практики работы с этим репозиторием. Документ собран по итогам ревью
ветки `copilot/integrate-missing-methods` — он отражает реальные грабли, а не
теоретический идеал.

## О проекте

`cu-cli` (модуль `cu-sync`) — CLI и MCP-сервер для LMS Центрального
Университета. `cu` умеет:
- авторизоваться через браузер и хранить cookie,
- ходить в публичный HTTP API LMS (`https://my.centraluniversity.ru/api/...`),
- отдавать те же данные хостам через MCP (Claude Code, IDE).

Go **1.26**. Зависимости управляются через `go mod`. Линтер — `golangci-lint`
v2.7.2 (конфиг в `.golangci.yml`). Сборка/линт/тесты — через `Makefile`.

## Архитектура

Слои (от внешнего к внутреннему):

```
cmd/cli/main.go               — точка входа, ставит version, вызывает RootCmd
internal/cli/command/         — cobra-команды CLI
internal/mcp/                 — MCP-сервер и тулы
  ├── server.go               — регистрирует все тулы, держит интерфейс LMSClient
  ├── format/                 — markdown-форматирование результатов для LLM
  └── tool/<name>/tool.go     — по одной папке на тул
internal/usecase/<feature>/   — оркестрация: gateway + бизнес-логика
internal/gateway/cu/          — HTTP-клиент LMS, DTO, авторизация
internal/model/               — общие input/output структуры между слоями
internal/version/             — версия бинаря (LDFLAGS)
```

**Правило зависимостей:** `cli` и `mcp` зависят от `usecase`/`gateway`, но не
наоборот. `gateway` не знает про `cobra`/`mcp`. `format` зависит только от
DTO/модели.

### Когда заводить usecase, а когда нет

- Если тул/команда делает 1–2 простых вызова и просто форматирует ответ —
  ходим напрямую в `gateway` (примеры: `coursestructure`, `coursesummary`,
  `listcourses`, `studentprofile`).
- Если есть нетривиальная логика (фильтрация, агрегация, события, маппинг
  нескольких ручек) — оформляем `internal/usecase/<feature>` с интерфейсом
  `LMSClient`, и CLI/MCP вызывают usecase (примеры: `grades`, `deadlines`,
  `task`, `materials`).

Не плодим usecase ради единообразия — это создаёт лишний слой без логики.

## Gateway (HTTP-клиент LMS)

### Все эндпоинты — константы в одном месте

`internal/gateway/cu/client.go` держит **все** пути API как константы:

```go
const (
    CourseEndpoint                = "/api/micro-lms/courses/%d"
    CourseOverviewEndpoint        = "/api/micro-lms/courses/%d/overview"
    StudentCoursesEndpoint        = "/api/micro-lms/courses/student"
    ThemeEndpoint                 = "/api/micro-lms/themes/%d"
    LongreadEndpoint              = "/api/micro-lms/longreads/%d"
    // ...
)
```

Никаких inline-строк `"/api/micro-lms/..."` в `course.go`/`longread.go`/etc.
Если добавляете эндпоинт — сначала константа, потом метод.

### Все GET-ы — через `doJSON[T]`

В `request.go` есть generic-хелпер:

```go
func doJSON[T any](ctx context.Context, c *Client, endpoint string) (*T, error)
```

Он сам проверяет `bff.cookie`, формирует запрос, декодирует тело или `APIError`.
Новый метод gateway — одна строка:

```go
func (c *Client) GetTheme(ctx context.Context, themeID int) (*Theme, error) {
    return doJSON[Theme](ctx, c, fmt.Sprintf(ThemeEndpoint, themeID))
}
```

Не копируем 30 строк обвязки `prepareRequest → executeRequest → decode`. Если
для нового эндпоинта `doJSON` не подходит (например, не GET или нестандартное
тело — см. `DownloadFile`) — пишем кастомно, но это исключение.

### DTO

- Все DTO живут в `internal/gateway/cu/dto.go`, теги JSON обязательны.
- Nullable-поля API → `*time.Time` / `*int` / `*Progress`. Пустую строку
  считаем «нет значения», `nil` указатель — тоже.
- Если API эволюционирует — синхронизируем DTO и `api-spec.yaml` в одном
  коммите.

## MCP-тулы

### Структура папки тула

`internal/mcp/tool/<name>/tool.go`:

```go
package <name>

type LMSClient interface { /* только нужные методы */ }

type Input struct {
    Field T `json:"field" jsonschema:"human-readable description"`
}

var Definition = &mcp.Tool{
    Name:        "snake_case_name",
    Description: "что делает + как получить параметры (см. ниже)",
}

func NewHandler(lms LMSClient) func(context.Context, *mcp.CallToolRequest, Input) (*mcp.CallToolResult, any, error) {
    return func(ctx context.Context, _ *mcp.CallToolRequest, in Input) (*mcp.CallToolResult, any, error) {
        // ВАЖНО: используем переданный ctx, НЕ context.Background()
        ...
        return textResult(mcpfmt.Something(out)), nil, nil
    }
}
```

После регистрации добавьте тул в `internal/mcp/server.go::registerTools`
**и** добавьте нужные методы в `LMSClient` интерфейс в том же файле.

### Контракт по описаниям

LLM выбирает тул по `Description`. Чтобы он не промахивался:

1. **Что делает.** Первая фраза — глагол + что возвращает.
2. **Как получить параметры.** Если на вход нужен численный ID, который
   неоткуда взять без другого тула — явно укажите: «Requires numeric
   `theme_id` — discover it via `get_course_structure`».
3. **Тип input в schema.** `course string` (имя или ID) vs `course_id int` —
   тип всегда честный. В описании поля поясните, как формат принимается.

### Контекст

**Всегда используйте переданный `ctx`.** Не пишите `ctx := context.Background()`.
Старая ошибка из ревью: четыре новых тула это делали, и пришлось отдельным
проходом фиксить все остальные.

### Format-слой

Markdown-вывод — в `internal/mcp/format/format.go`. Один файл, функции
называются по сущности (`CourseSummary`, `ThemeSummary`, `Deadlines`). Тулы
не форматируют у себя — только дергают `mcpfmt.X(out)`. Это даёт единый стиль
и упрощает тесты.

## CLI

### Обработка ошибок

В CLI-командах **никогда** не печатаем ошибки в stdout с `return`. Есть хелпер
`internal/cli/command/error.go::exitErrf`:

```go
if err != nil {
    exitErrf("Failed to fetch theme: %v", err)
}
```

Он пишет в stderr и делает `os.Exit(1)`. RootCmd работает с
`SilenceUsage: true, SilenceErrors: true` — usage не выводится при ошибке
runtime, а сами ошибки печатает `main`.

Не используйте `panic` для пользовательских ошибок — была старая бага в
`cookieRequiredError`, чинили отдельно.

### Версия

`cu --version` и `cu version` — оба работают. `Version` поле проставляется в
`main.go` после `version.Set(...)`. Не дублируйте инициализацию в `root.go`.

### Структура команды

```go
var fetchThemeCmd = &cobra.Command{
    Use:   "theme [theme-id]",
    Short: "...",
    Long:  "...",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        ctx := cmd.Context()
        client := mustClient() // авторизация уже проверена внутри
        // ...
    },
}
```

`mustClient()` падает через `cookieRequiredError`, если нет cookie — отдельной
проверки не нужно.

## Тестирование

### Юнит-тесты gateway

Стиль: `testify/suite`, табличные кейсы внутри методов сьюита (см.
`internal/gateway/cu/client_test.go`). Поднимается `httptest.Server`, в
`SetupTest` регистрируются хендлеры на конкретные пути API.

**Правила фикстур:**
- Не подсовывайте `&time.Time{}` (non-nil указатель на zero-time) — тесты
  пройдут, но в форматтерах появится `0001-01-01 00:00:00`. Либо `nil`, либо
  `time.Date(2025, ...)`.
- Один success-тест проверяет и поля данных, и `nil`/`NotNil` на указателях
  (см. `TestGetCourse_Success`).
- Сценарии ошибок — отдельные методы (`*_NoCookie`, `*_InvalidCookie`).

### Ручная проверка нового функционала (обязательный чеклист)

После любых правок в gateway/CLI/MCP пройдите по списку:

1. **Сборка и линт чистые:**
   ```bash
   make build && make lint && go test ./...
   ```
   Линт обязателен — он ловит `dupl`, `goprintffuncname`, форматирование
   `golines`. Не комитьте, пока `0 issues`.

2. **CLI с реальным cookie:**
   ```bash
   ./bin/cu <new-command> <args>
   ```
   Проверяйте golden path и edge-cases (несуществующий ID, пустой ответ).
   Ошибки должны идти в stderr и давать `echo $?` равный 1.

3. **MCP через JSON-RPC stdio** (если добавили/изменили тул):
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./bin/cu mcp
   echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"<tool>","arguments":{...}}}' | ./bin/cu mcp
   ```
   Убедитесь, что:
   - тул есть в `tools/list`, нет дубликатов,
   - все остальные тулы вызываются без регрессий (минимум `list_courses`,
     `list_deadlines`, `search_courses`).

4. **Перезапуск MCP в хосте.** `go install` / `make install` молча НЕ обновит
   `~/go/bin/cu`, пока хост (Claude Code/IDE) держит старый процесс. После
   установки:
   ```bash
   claude mcp restart cu      # или /mcp reconnect в Claude Code
   ```
   `make install` в этом репо делает `cp + codesign` вручную как раз ради
   этого случая — не возвращайтесь к `go install`.

### Интеграционные тесты с реальным API

Их в репо нет. Заменяем их пунктом 2 чеклиста выше (ручная проверка с живым
cookie). Если автоматизируете — выносите за `// +build integration`.

## Workflow коммитов

### Структура веток и сообщений

- Ветки фич: `<owner>/<short-slug>` (например `copilot/integrate-missing-methods`).
- Сообщения — Conventional Commits: `feat:`, `fix:`, `chore:`, `docs:`, `tidy:`.
- В коммите — одно осмысленное изменение. Не складывайте «feat + accidentally
  binary + format fix» в один commit.

### Что нельзя коммитить

- **Бинари** (`cu`, `bin/`, `build/`). `.gitignore` это закрывает; если бинарь
  всё-таки попал — удаляйте его в той же ветке через rebase/squash, **до**
  мержа в `main`. Blob 15 MB остаётся в истории навсегда. Реальный случай:
  ee3530f → пришлось делать `git reset --soft main` + force-push.
- Конфиг-файлы с секретами (`.env`, `.env.local`, cookie-файлы).
- Артефакты сборки (`coverage.html`, `*.out`).

### Squash и force-push

Феча-ветка с историей экспериментов перед мержем в `main` обычно сквошится:

```bash
git add -A
git reset --soft main
git commit -m "feat: ..."
git push --force-with-lease
```

`--force-with-lease` (не `--force`) — чтобы не затереть чужой push.

### Pre-commit обязательное

Перед коммитом всегда: `make lint && go test ./...`. Если правок много —
прогоняйте после каждой логической группы, не в конце.

## Релизы и CHANGELOG

### Поток

1. **`main` → авто-тег** (`.github/workflows/release-tag.yml`). Каждый пуш в
   `main` бампит patch и пушит тег `v0.1.N`. Major/minor — руками через
   `git tag` или GitHub UI.
2. **Тег `v*` → GoReleaser** (`.github/workflows/release.yml`). Собирает
   `linux/{amd64,arm64}`, `darwin/{amd64,arm64}`, `windows/amd64`, кладёт в
   `tar.gz`/`zip` вместе с README/CHANGELOG/LICENSE, считает общий
   `checksums.txt`, публикует Release.
3. **После релиза CHANGELOG.md ротируется автоматически**: `[Unreleased]`
   переименовывается в `[<version>] - <date>`, сверху создаётся новый пустой
   `[Unreleased]`, коммит уходит в `main` с `[skip ci]`.

Тело релиза = содержимое секции `[Unreleased]` на момент тега, извлечённое
скриптом `scripts/extract-changelog.sh`. Если секция пуста — workflow
подставит `Release v<version>` и оставит warning в логах.

### Правила работы с CHANGELOG.md

- **Любая фича/фикс, видимая пользователю, добавляется в `[Unreleased]`
  в том же PR/коммите, что и изменение кода.** Не в отдельном «changelog
  PR».
- Секции — `### Added`, `### Changed`, `### Fixed`, `### Removed`,
  `### Security` (Keep-a-Changelog). Опускайте те, что пустые.
- Записи — для пользователя, не для коммита. «Добавили MCP-тул
  `get_theme_summary`», а не «refactor: extracted handler». Внутренний
  рефакторинг в changelog не пишем.

### Перед релизом — обязательный вопрос инженеру

Прежде чем рекомендовать пуш в `main` (или вручную тегать), **спросите
владельца изменений**:

> «Отражает ли текущий `[Unreleased]` всё, что реально вошло в этот релиз?»

Это страховка от двух типичных ошибок:

1. **Дрейф changelog от реальности.** Кто-то закоммитил фичу без записи в
   CHANGELOG. Авто-релиз уйдёт с пропуском — пользователь не узнает.
2. **Записи без кода.** Кто-то добавил в `[Unreleased]` запись, а фича
   откатилась/не доехала. Релиз обещает то, чего нет.

Не делайте этот checkpoint автоматическим — он именно про человеческую
сверку. В сомнительных случаях сравните `[Unreleased]` с
`git log <prev-tag>..HEAD --oneline` и явно перечислите расхождения.

### Установка из релиза

Конечный пользователь ставит через `install.sh` (или ручной скачкой архива):

```bash
curl -fsSL https://raw.githubusercontent.com/EgorTarasov/cu/main/install.sh | sh
```

Скрипт детектит OS/ARCH, тянет последний release, проверяет sha256,
ставит в `~/.local/bin`, на macOS снимает quarantine. См. README/install.sh
для опций (`CU_VERSION`, `CU_INSTALL_DIR`).

## Эволюция API

`api-spec.yaml` в корне — поверхностное описание схем API LMS. **Должно
совпадать** с DTO в `internal/gateway/cu/dto.go`. При добавлении ручки:

1. Добавьте схему в `api-spec.yaml` (тип, nullable, enum, format).
2. Добавьте/расширьте DTO в `dto.go`.
3. Добавьте константу в `client.go` и метод-обёртку.
4. Спецификация и код едут одним коммитом или в одной PR-цепочке.

Для `number` всегда указывайте `format` (`float`/`double`) — иначе генераторы
выдают неконсистентный тип.

## Чек-лист для нового LMS-эндпоинта

Полный путь «добавить новую ручку API» из реального опыта:

1. **api-spec.yaml** — схема ответа.
2. **gateway/cu/dto.go** — DTO с JSON-тегами.
3. **gateway/cu/client.go** — константа `XxxEndpoint`.
4. **gateway/cu/<area>.go** — метод `func (c *Client) GetXxx(ctx, ...) (*Xxx, error)`
   через `doJSON[Xxx]`.
5. **gateway/cu/client_test.go** — fake-handler + один success-тест с
   реалистичными временами + ассерты nil/NotNil.
6. **CLI:** новая команда в `internal/cli/command/fetch.go` (или своя группа)
   с использованием `exitErrf` для ошибок.
7. **MCP:** папка `internal/mcp/tool/<name>/tool.go` + регистрация в
   `internal/mcp/server.go` + расширение `LMSClient` интерфейса.
8. **Формат:** функция в `internal/mcp/format/format.go`, если данные нужны
   LLM-у в markdown.
9. **README.md** — короткое упоминание новой CLI-команды (если она для
   пользователя).
10. **Локально:** `make lint && go test ./... && make build && ./bin/cu ...`
    + ручной MCP smoke-test.

Если хотя бы один пункт пропущен — поймает либо ревью, либо линтер.
