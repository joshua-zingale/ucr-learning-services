CREATE TYPE agent_ability_type AS ENUM ('manage', 'interact');
CREATE TYPE message_type AS ENUM ('user', 'agent');

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
    ability agent_ability_type,
    primary key (user_id, agent_id, ability)
);

create table group_agent_permissions (
    group_id varchar(32),
    agent_id int references agents(agent_id) on delete cascade,
    ability agent_ability_type,
    primary key (group_id, agent_id, ability)
);

create table conversations (
    conversation_id serial primary key,
    user_id varchar(32) not null,
    name varchar(32) not null
);

create table messages (
    message_id serial primary key,
    conversation_id int not null references conversations(conversation_id) on delete cascade,
    content text not null,
    sent_at timestamp default NOW(),

    author_type message_type,
    agent_id int references agents(agent_id),
    user_id varchar(32),
    check (
        (author_type = 'agent' AND agent_id is not null AND user_id is null)
        OR
        (author_type = 'user' AND user_id is not null AND agent_id is null))
);