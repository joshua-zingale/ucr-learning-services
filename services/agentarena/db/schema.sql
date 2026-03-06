CREATE TYPE agent_ability_type AS ENUM ('manage', 'interact');
CREATE TYPE message_type AS ENUM ('user', 'agent');


create table agent_classes(
    agent_class_id varchar(32) primary key,
    name varchar(32) not null
);

create table agents (
    agent_id serial primary key,
    name varchar(32) not null,

    agent_class_id varchar(32) references agent_classes(agent_class_id) not null,
    config jsonb not null,

    created_at timestamptz default now()
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
    name varchar(32) not null,
    active_agent_id int references agents(agent_id)
);

create table messages (
    message_id serial primary key,
    conversation_id int not null references conversations(conversation_id) on delete cascade,
    content text not null,
    sent_at timestamptz default NOW(),

    message_type message_type not null,
    agent_id int references agents(agent_id),
    user_id varchar(32),
    check (
        (message_type = 'agent' AND agent_id is not null AND user_id is null)
        OR
        (message_type = 'user' AND user_id is not null AND agent_id is null))
);