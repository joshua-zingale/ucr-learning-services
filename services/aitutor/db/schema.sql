CREATE TYPE ability_type AS ENUM ('read', 'write');


create table agents (
    agent_id serial primary key,
    name varchar(32) not null
);

create table agent_configs (
    agent_id int references agents(agent_id) on delete cascade primary key ,
    system_prompt text not null
);

create table user_agent_permissions (
    user_id varchar(32),
    agent_id int references agents(agent_id) on delete cascade,
    ability ability_type,
    primary key (user_id, agent_id, ability)
);

create table group_agent_permissions (
    group_id varchar(32),
    agent_id int references agents(agent_id) on delete cascade,
    ability ability_type,
    primary key (group_id, agent_id, ability)
);

create table conversations (
    conversation_id serial primary key,
    name varchar(32)
);

create table messages (
    message_id serial primary key,
    conversation_id int not null references conversations(conversation_id) on delete cascade,
    content text,
    sent_at timestamp default NOW(),

    author_type varchar(6) check (author_type in ('agent', 'user')),
    agent_id int references agents(agent_id),
    user_id varchar(32),
    check (
        (author_type = 'agent' AND agent_id is not null AND user_id is null)
        OR
        (author_type = 'user' AND user_id is not null AND agent_id is null))
);