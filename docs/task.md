# cu task

Показывает детальную информацию о конкретном задании: статус, оценку, дедлайн, ревьюера, ссылку на решение.

## Использование

```bash
cu task <task-id>
```

ID задачи можно найти в выводе `cu deadlines` или `cu grades`.

## Пример

```
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

## Что показывается

- **State** — текущий статус (TODO, IN PROGRESS, SUBMITTED, DONE, FAILED)
- **Score** — текущая оценка / максимум
- **Activity** — тип активности и вес в итоговой оценке
- **Deadline** — дедлайн и сколько времени осталось
- **Timeline** — когда начато, отправлено, отклонено, оценено
- **Reviewer** — кто проверяет, с email
- **Solution** — ссылка на ваш MR / решение
- **Late days** — сколько late days осталось на балансе
