-- name: GetTask :one
SELECT * FROM tasks
WHERE id = $1 LIMIT 1;

-- name: GetTasks :many
SELECT * FROM tasks
ORDER BY id;

-- name: CreateTask :exec
INSERT INTO tasks (
    id, command, status
) VALUES (
             $1, $2, 'queued'
         );

-- name: GetNextTask :one
UPDATE tasks
SET status = 'started'
WHERE  id = (
    SELECT id
    FROM   tasks
    WHERE started_at is null
    LIMIT  1
)
RETURNING id, command;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;

-- name: FinishTask :exec
UPDATE tasks SET  started_at = $2, finished_at = $3, status = $4, stdout = $5, stderr = $6, exit_code = $7 WHERE id = $1;

