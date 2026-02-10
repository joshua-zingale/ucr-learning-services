-- name: GetAgentClassFromSlug :one
select agent_class_id, name, config_schema
from agent_classes
where slug = $1
limit 1;

-- name: GetAgentClassIdFromSlug :one
select agent_class_id
from agent_classes
where slug = $1
limit 1;

-- name: GetAgentFull :one
select agents.agent_id, agents.name, agents.agent_class_id, agents.config
from agents
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


-- name: GetConversationMessages :many
select m.message_id, m.content, m.sent_at, m.author_type, m.agent_id, m.user_id
from messages as m
where m.conversation_id = $1
order by m.sent_at desc, m.message_id asc;


-- name: StartedConversation :one
select exists (
    select 1
    from conversations
    where conversation_id = $1 AND user_id = $2
);