# Changelog

Все изменения проекта документируются в этом файле.

Формат — [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/), версионирование —
[SemVer](https://semver.org/lang/ru/).

Раздел `[Unreleased]` накапливает изменения, которые попадут в следующий релиз.
При публикации тега workflow `release.yml` берёт его содержимое как тело релиза,
после чего автоматически переименовывает раздел в `[<version>] - <date>` и
создаёт новый пустой `[Unreleased]` сверху.

## [Unreleased]

### Added
- Установочный скрипт `install.sh` для macOS и Linux (Ubuntu/Fedora, amd64/arm64).
- Автоматическая публикация релизов через GoReleaser в `.github/workflows/release.yml`.
- Пакет `internal/render` — общие writer-хелперы для plain/markdown вывода
  (первый шаг к Renderer/Presenter паттерну).

### Changed
- Сборка теперь упакована в `tar.gz`/`zip` со списком файлов (бинарь, README,
  CHANGELOG, LICENSE) и единым `checksums.txt`.

## [0.1.1]

Базовый набор CLI и MCP-сервера. История до этого момента восстановлена в
дальнейших релизах по необходимости — см. git log.
