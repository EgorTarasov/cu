# CU CLI - Commands Reference

## Authentication

### `cu login`

Opens Chrome browser for Keycloak SSO login. Captures `bff.cookie` automatically.

```bash
cu login
cu login --timeout 10m
```

Cookie is saved to `~/.cu-cli/cookie` and used by all subsequent commands.

---

## Courses

### `cu courses`

List all your published courses with IDs.

```bash
$ cu courses
Your courses (8)

  1. [947] Java Spring (Разработка веб-приложений на Java с использованием Spring)
  2. [824] Soft Lab. Gogol School
  3. [950] SQL и базы данных (для разработчиков)
  4. [901] Алгоритмы и структуры данных. Часть 2
  5. [900] Разработка на языке программирования Go
  ...

Use course ID or name with other commands:
  cu deadlines go
  cu grades алгоритмы
  cu materials 901
```

---

## Deadlines

### `cu deadlines [course]`

Show upcoming deadlines across all courses or for a specific one.

Course can be specified by **ID** or **name** (case-insensitive substring match).

```bash
# All deadlines
$ cu deadlines
All upcoming deadlines

 ! TODO          3h 47m    21 Apr 20:59  ДЗ 5. Неделя 9. «Kubernetes 1»
   Промышленная разработка. Магистратура
 * IN PROGRESS   1d 3h     22 Apr 20:59  ДЗ 5. Неделя 10.
   SQL и базы данных (для разработчиков)
 * TODO          2d 3h     23 Apr 20:59  ДЗ 4. Неделя 10.
   Разработка на языке программирования Go

15 deadline(s) total
  ! = overdue or <24h  * = <3 days
```

```bash
# Filter by course name
$ cu deadlines go
Deadlines: Разработка на языке программирования Go

 * IN PROGRESS   2d 1h     23 Apr 19:00  ДЗ 3. Неделя 8.
   reviewer: Владислав Ханнанов
 * TODO          2d 3h     23 Apr 20:59  ДЗ 4. Неделя 10.

4 deadline(s) total
```

```bash
# Filter by ID
$ cu deadlines 901
```

---

## Grades

### `cu grades [course]`

Show grades. Without arguments shows a progress bar summary. With a course argument shows detailed gradebook.

```bash
# Summary across all courses
$ cu grades
Grades summary

  Java Spring (Разработка веб-приложений на Java с …  [#######-------------] 3.6/10
  SQL и базы данных (для разработчиков)               [####----------------] 2.5/10
  Алгоритмы и структуры данных. Часть 2               [####----------------] 2.0/10
  Разработка на языке программирования Go             [##------------------] 1.1/10

Use: cu grades <course> for detailed view
```

```bash
# Detailed grades for a course
$ cu grades алгоритмы
Grades: Алгоритмы и структуры данных. Часть 2

Activity breakdown:
  Контесты (30%)                       avg=6.7  total=2.0
  Экзамен (50%)                        avg=0.0  total=0.0
  Устная защита                        avg=5.0  total=0.0 [BLOCKER]

  Total score: 2.0

Tasks:
  DONE           10/10  ДЗ 1. Неделя 3. «Контест 1»
  DONE           10/10  ДЗ_1_Неделя 3_Устная защита
  DONE           10/10  ДЗ 2. Неделя 6. «Контест 2»
  TODO            -/10  ДЗ 2_Неделя 6_Устная защита
  TODO            -/10  ДЗ 4. Неделя 12. «Контест 3»

Blockers:
  Устная защита (need avg >= 10)
```

---

## Materials

### `cu materials <course> [flags]`

Download all course PDFs and show external links (git, notion).

| Flag | Description |
|------|-------------|
| `--links` | Only show links, don't download files |
| `--week N` | Download only a specific week |
| `--path DIR` | Output directory (default: `.`) |

```bash
# Download all materials
$ cu materials алгоритмы --path ./downloads
Materials: Алгоритмы и структуры данных. Часть 2

[Неделя 1: Основы теории графов]
  saved: Неделя 1_..._Лекция.pdf
  saved: Неделя 1_..._Семинар.pdf
  saved: Неделя 1_..._Семинар_Указания.pdf
...
Downloaded 32/32 files
```

```bash
# Only show links (no download)
$ cu materials java --links
Materials: Java Spring (...)

[Неделя 1: Экосистема Spring]
  [PDF] Неделя 1_Java Spring_Лекция.pdf (9595.1 KB)
  [link] https://git.culab.ru/acs/java-spring-2026/-/blob/main/week_01/longread_01.md
  [link] https://git.culab.ru/acs/java-spring-2026/-/blob/main/week_01/seminar_01.md
```

```bash
# Download only week 8
$ cu materials go --week 8
```

---

## Task

### `cu task <task-id>`

Show detailed info about a specific task. Task IDs appear in `cu deadlines` and `cu grades` output.

```bash
$ cu task 1536681
Task: ДЗ 3. Неделя 8.
Course: Разработка на языке программирования Go
Theme: Неделя 8: Контексты. Часть 2

State:    IN PROGRESS
Score:    9/10
Activity: Домашние задания (40%)

Deadline: 23 Apr 2026 19:00 (2d 1h left)
Started:  29 Mar 2026 13:14
Submitted: 09 Apr 2026 18:41
Rejected: 20 Apr 2026 06:40

Reviewer: Владислав Ханнанов (v.khannanov@centraluniversity.ru)
Solution: https://git.culab.ru/.../merge_requests/1

Late days balance: 28
```

---

## Legacy Commands

### `cu fetch courses`

Same as `cu courses` but with more verbose output (published date, skill level, progress).

### `cu fetch course <id> [--dump] [--path DIR]`

Fetch course overview. With `--dump` downloads all files (similar to `cu materials`).

---

## Course Resolution

All commands that accept a `[course]` argument support flexible lookup:

| Input | Match |
|-------|-------|
| `900` | Course with ID 900 |
| `go` | Matches "Разработка на языке программирования **Go**" (word boundary) |
| `sql` | Matches "**SQL** и базы данных" |
| `алгоритмы` | Matches "**Алгоритмы** и структуры данных" |
| `spring` | Matches "Java **Spring**" |

If multiple courses match, you'll see all options:

```
Error: multiple courses match "go":
  824  Soft Lab. Gogol School
  900  Разработка на языке программирования Go
specify more precisely or use ID
```

Word-boundary matching is prioritized: `go` resolves to "Go" (not "Gogol") because "Go" is a standalone word.
