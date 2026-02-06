-- name: GetAgent :one
select *
from agents
where agent_id=$1
limit 1;


-- name: GetAgentConfig :one
select *
from agent_configs
where agent_id=$1
limit 1;