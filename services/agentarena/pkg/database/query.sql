-- name: GetAgentConfigFromConversationId :one
select a.config, a.agent_id, a.agent_class_id
from agents as a join conversations as c
    on a.agent_id = c.active_agent_id
where c.conversation_id = $1
limit 1;


-- name: GetAgent :one
select a.agent_id, a.name, a.agent_class_id, a.config
from agents a
where a.agent_id=$1
limit 1;

-- name: createAgentWithManagerAndInteractorUnchecked :one
with new_agent as (
    insert into agents (name, agent_class_id, config)
    values ($1, $2, $3)
    returning agent_id
)
insert into user_agent_permissions (user_id, agent_id, ability)
select
    $4 AS user_id, 
    new_agent.agent_id AS agent_id, 
    'manage'::agent_ability_type AS ability
from new_agent
union all
select $4, new_agent.agent_id, 'interact'
from new_agent
returning agent_id;

-- name: setAgentConfigUnchecked :exec
update agents
set config = $2
where agent_id = $1;


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

-- name: GetConversationMetadata :one
select conversation_id, name, active_agent_id
from conversations c
where c.conversation_id = $1
limit 1;

-- name: CreateConversationWithInitialMessage :one
WITH new_conv AS (
    INSERT INTO conversations (user_id, name, active_agent_id)
    VALUES (@user_id, @conversation_name, @agent_id)
    RETURNING conversation_id, user_id, active_agent_id
)
INSERT INTO messages (
    conversation_id, 
    content, 
    message_type, 
    user_id, 
    agent_id
)
SELECT 
    new_conv.conversation_id, 
    @message_content, 
    'user',
    new_conv.user_id, 
    NULL            -- agent_id is null for user messages
FROM new_conv
RETURNING conversation_id, message_id;

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