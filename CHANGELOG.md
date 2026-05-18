# Changelog

Все изменения проекта документируются в этом файле.

Формат — [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/), версионирование —
[SemVer](https://semver.org/lang/ru/).

Раздел `[Unreleased]` накапливает изменения, которые попадут в следующий релиз.
При публикации тега workflow `release.yml` берёт его содержимое как тело релиза,
после чего автоматически переименовывает раздел в `[<version>] - <date>` и
создаёт новый пустой `[Unreleased]` сверху.

## [Unreleased]

> **Примечание для пользователей с v0.1.2 и старше.** Релиз `v0.1.3` был
> опубликован на GitHub без бинарных артефактов (GoReleaser не отработал),
> поэтому все его изменения по факту впервые доезжают до пользователей
> именно в этом релизе. Соответствующая секция `[0.1.3]` ниже остаётся как
> исторический референс — функциональность из неё уже включена.

### Added
- **Архивные курсы.** `cu courses`, `cu fetch courses`, `cu show` и MCP-тулы
  `list_courses` / `search_courses` теперь возвращают активные и архивные
  курсы отдельными секциями с пометкой статуса. Архивные тянутся через
  `/api/micro-lms/courses/student?state=archived`. `cu show` дополнительно
  принимает numeric `course-id` (раньше работал только fuzzy-поиск по имени).
- **Авто-проверка обновлений.** После каждой команды `cu` раз в 7 дней
  запрашивает последний релиз с GitHub и печатает в stderr подсказку, если
  установленная версия отстаёт. Состояние кэшируется в
  `~/.cu-cli/update-check`; сборки с `version == "dev"` пропускают проверку.
  Сетевой запрос с 3-секундным таймаутом, любые ошибки — silent (не мешают
  основной команде).

### Changed
- `ResolveCourse` (по имени/ID) теперь ищет как среди активных, так и среди
  архивных курсов — раньше архивные были невидимы и обращаться к ним
  приходилось только по numeric ID через `get_course_summary` /
  `get_course_structure`.
- `api-spec.yaml`: у `GET /micro-lms/courses/student` явно описан
  `state=archived` (плюс `offset`); у `/courses/{id}/overview` отмечено, что
  путь общий для активных и архивных курсов.

### Fixed
- Подавление `gosec G101` для трёх env-констант (`CU_MM_TOKEN`,
  `CU_KTALK_NG_TOKEN`, `CU_KTALK_SESSION_TOKEN`) переведено на формат
  `// #nosec G101 ...`, чтобы standalone-gosec в CI тоже признавал
  директиву (golangci-lint видел её и раньше).

## [0.1.3] - 2026-05-17

### Added
- Установочный скрипт `install.sh` для macOS и Linux (Ubuntu/Fedora, amd64/arm64).
- Автоматическая публикация релизов через GoReleaser в `.github/workflows/release.yml`.
- Пакет `internal/render` — общие writer-хелперы для plain/markdown вывода
  (первый шаг к Renderer/Presenter паттерну).
- Авторизация в `time.cu.ru` через браузер: `cu login --time`
  ловит cookie `MMAUTHTOKEN` и сохраняет в `~/.cu-cli/mm-cookie`.
- Команды `cu time sync` и `cu time posts` — инкрементальный pull сообщений от
  `@cu_notification_bot` в локальный JSONL-сторадж
  (`~/.cu-cli/mm/<channel-id>/posts.jsonl`). Настройки через env
  `CU_MM_BASE_URL`, `CU_MM_TEAM`, `CU_MM_BOT_USERNAME`, `CU_MM_TOKEN`.
  Алиас `cu mm ...` сохранён.
- Команда `cu time recordings [query]` — поиск записей занятий по предмету в
  локальном сторадже с интерактивным выбором, флаг `--all` для печати всех
  совпадений. Парсер живёт в `internal/usecase/recordings` и доступен другим
  фичам (например, для обогащения сводок курсов) через `UseCase.ForCourse`.
- Авторизация в Ktalk (`centraluniversity.ktalk.ru`) через браузер:
  `cu login --ktalk` ловит cookie `ngtoken`/`kontur_ngtoken` плюс
  localStorage/sessionStorage и сохраняет в `~/.cu-cli/ktalk-tokens.json`.
- Команда `cu time notifications` — извлекает из бот-сообщений ссылки на новые
  задачи и оценки, индексирует по `course_id`/`longread_id` и печатает
  пары `LMS-ссылка` + `time-permalink`. Фильтры `--course`, `--longread`.
- `cu time recordings` теперь печатает permalink на исходное time-сообщение
  рядом с ссылками на запись (поле `PostURL` у `Recording`).
- Команда `cu show [query]` — интерактивный поиск курса (fuzzy по словам)
  и печать полного дерева тем/лонгридов. Если настроен time.cu.ru, к курсу
  подмешиваются записи занятий, а к каждому лонгриду — permalink-и на
  сообщения бота в чате.

### Changed
- Сборка теперь упакована в `tar.gz`/`zip` со списком файлов (бинарь, README,
  CHANGELOG, LICENSE) и единым `checksums.txt`.
- Внутренний пакет `internal/gateway/mm` переименован в
  `internal/gateway/timeclient`. Внешние пути файлов (`~/.cu-cli/mm-cookie`,
  `~/.cu-cli/mm/<channel>/`) и env-переменные `CU_MM_*` оставлены без изменений,
  поэтому уже залогиненная сессия и synced storage продолжают работать.

## [0.1.1]

Базовый набор CLI и MCP-сервера. История до этого момента восстановлена в
дальнейших релизах по необходимости — см. git log.
