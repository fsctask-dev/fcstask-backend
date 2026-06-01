# Worker: описание и архитектура

## Что это такое

**Worker** — это job в GitLab CI, который запускается по push/MR и выполняет проверку. Каждое событие запускает отдельный CI‑пайплайн.

Типовой сценарий:

1) GitLab ловит push/MR
2) GitLab CI запускает job `grade`
3) job клонирует репозиторий, запускает тесты, считает баллы
4) job отправляет результат в backend API

## Как запускается (GitLab CI)

### 1) Webhook не нужен

GitLab сам запускает CI по событию (push/MR). Никакой отдельный webhook в backend не обязателен.

### 2) В `.gitlab-ci.yml` есть job `grade`

- Job запускается на push/MR.
- Внутри job выполняется:
  - checkout нужного commit
  - запуск тестов
  - отправка результата

### 3) Данные приходят из GitLab CI

GitLab CI даёт в job готовые переменные:

- `CI_PROJECT_URL` — URL репозитория
- `CI_COMMIT_SHA` — commit
- `CI_COMMIT_REF_NAME` — ветка
- `CI_MERGE_REQUEST_IID` — MR id (если MR)
- `CI_PROJECT_ID`, `CI_PIPELINE_ID`, `CI_JOB_ID`

## Источники данных (откуда что берётся)

### Репозиторий и commit

- берём из GitLab CI (`CI_PROJECT_URL`, `CI_COMMIT_SHA`)
- git clone/checkout делает сам runner

### Список задач `tasks[]`

Есть два варианта:

**Вариант A (простой):**
- tasks лежат в репозитории (например, `tasks.yml`).
- job читает файл и формирует список.

**Вариант B (централизованный сложный?!!):**
- job запрашивает список задач из backend:
  - `GET /api/courses/:courseId/tasks`

### Идентификация студента

- `student_id` берётся из backend по репозиторию:
  - `POST /api/resolve-student`
  - вход: `repo_url`
  - выход: `student_id`, `course_id`

Это нужно, потому что GitLab CI не знает, кто конкретно студент.

## Куда отправляются результаты

После проверки job отправляет результат в backend:

- `POST /api/grades`

```json
{
  "job_id": "gitlab-job-123",
  "student_id": "student-42",
  "course_id": "course-1",
  "commit_sha": "abc123",
  "total_score": 2.0,
  "tasks": [
    {"task": "task1", "score": 1.0, "passed": true, "log": "...", "duration_ms": 12345},
    {"task": "task2", "score": 1.0, "passed": true, "log": "...", "duration_ms": 9800}
  ]
}
```

## Список задач воркера (вход/выход подробно)

### 1) Получить контекст CI

**Вход:** env переменные GitLab CI
- `CI_PROJECT_URL`
- `CI_COMMIT_SHA`
- `CI_COMMIT_REF_NAME`
- `CI_PIPELINE_ID`
- `CI_JOB_ID`

**Выход:** внутренний `JobContext`

### 2) Определить `student_id` и `course_id`

**Вход:**
- `repo_url` (из `CI_PROJECT_URL`)

**Запрос в backend:**
- `POST /api/resolve-student`

**Выход:**
```json
{ "student_id": "student-42", "course_id": "course-1" }
```

### 3) Получить список задач

**Вариант A:**
- читаем `tasks.yml` из репозитория

**Вариант B:**
- `GET /api/courses/:courseId/tasks`

**Выход:**
```json
["task1", "task2"]
```

### 4) Запустить тесты

**Вход:**
- `repo_dir`
- `tasks[]`

**Выход (сырой результат):**
```json
{ "task": "task1", "exit_code": 0, "stdout": "...", "stderr": "...", "duration_ms": 12345 }
```

### 5) Преобразовать результат в баллы

**Вход:** stdout/stderr/exit_code
**Выход:**
```json
{ "task": "task1", "score": 1.0, "passed": true, "log": "..." }
```

### 6) Собрать итоговый результат

**Вход:** TaskResult[], JobContext
**Выход:** GradeResult (см. выше)

### 7) Отправить результат в backend

**Запрос:** `POST /api/grades`

## Какие контракты обязаны быть в backend

Чтобы CI‑воркер работал, backend должен иметь:

1. `POST /api/resolve-student`
   - вход: `repo_url`
   - выход: `student_id`, `course_id`

2. `GET /api/courses/:courseId/tasks` (если не храним tasks в репозитории)

3. `POST /api/grades`

Если этих трёх пунктов нет — job не сможет корректно отправить оценку.

