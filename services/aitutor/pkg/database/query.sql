-- name: GetAgentFull :one
select agents.agent_id, agents.name, agent_configs.system_prompt
from agents join agent_configs on agents.agent_id = agent_configs.agent_id
where agents.agent_id=$1
limit 1;