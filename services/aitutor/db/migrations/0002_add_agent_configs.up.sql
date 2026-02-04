create table agent_configs (
    agent_id int references agents(agent_id) on delete cascade primary key ,
    system_prompt text
);
