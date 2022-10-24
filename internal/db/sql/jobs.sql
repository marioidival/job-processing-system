-- name: GetAllJobs :many
select * from jobs order by created_at;


-- name: GetJobsByStatus :many
select * from jobs where status = sqlc.arg(status) order by created_at;


-- name: CreateJob :exec
insert into jobs (action, data)
values (sqlc.arg(action), sqlc.arg(data))
returning *;


-- name: GetPendingJobs :many
select id, data, "action" from jobs
where status = 'PENDING'
order by created_at
for update skip locked;


-- name: UpdateJob :exec
update jobs
set result = sqlc.arg(result), status = sqlc.arg(status), updated_at = now()
where id = sqlc.arg(job);