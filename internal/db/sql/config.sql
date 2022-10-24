-- name: GetPollingInterval :one
select time from config limit 1;