-- name: GetAgentFull :one
select agents.agent_id, agents.name, agent_configs.system_prompt
from agents join agent_configs on agents.agent_id = agent_configs.agent_id
where agents.agent_id=$1
limit 1;

-- name: HasAgentPermission :one
SELECT EXISTS (
    SELECT 1 FROM user_agent_permissions uap
    WHERE uap.user_id = @user_id 
      AND uap.agent_id = @agent_id 
      AND uap.ability = @ability
    
    UNION ALL
    
    SELECT 1 FROM group_agent_permissions
    WHERE group_id = ANY(@group_ids::text[])
      AND agent_id = @agent_id 
      AND ability = @ability
);

-- name: GetConversations :many
select conversation_id, name
from conversations
where user_id = $1;