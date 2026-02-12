-- name: GetAgentClassFromSlug :one
select agent_class_id, name, config_schema
from agent_classes
where slug = $1
limit 1;

-- name: GetAgentConfigFromConversationId :one
select a.config, a.agent_id, a.agent_class_id
from agents as a join conversations as c
    on a.agent_id = c.active_agent_id
where c.conversation_id = $1
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

-- name: GetUserConversations :many
select c.conversation_id, c.name
from conversations c
left join messages m on c.conversation_id = m.conversation_id
where c.user_id = $1
group by c.conversation_id, c.name
order by MAX(m.sent_at) desc nulls last;


-- name: GetConversationMessages :many
select m.message_id, m.content, m.sent_at, m.message_type, m.agent_id, m.user_id
from messages as m
where m.conversation_id = $1
order by m.sent_at asc, m.message_id asc;


-- name: PostMessageToConversation :one
insert into messages (conversation_id, content, message_type, agent_id, user_id) values
(@conversation_id, @content, @message_type, @agent_id, @user_id)
returning message_id as message_id, content as content;


-- name: StartedConversation :one
select exists (
    select 1
    from conversations
    where conversation_id = $1 AND user_id = $2
);