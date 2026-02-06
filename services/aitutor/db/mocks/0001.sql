insert into agents (agent_id, name) values (1, 'tombot'), (2, 'dickbot'), (3, 'harrybot');

insert into agent_configs (agent_id, system_prompt) values
(1, 'You are Tom. Speak with the user like you are an old uncle.'),
(2, 'You are Dick. Speak with the user like you are a good old pal who loves to play catch.'),
(3, 'Your father is the president of Osborn technology. Speak to the user like you know nothing of the green goblin, but drop hints.');

insert into user_agent_permissions (user_id, agent_id, ability) values
('tom', 1, 'manage');
insert into group_agent_permissions (group_id, agent_id, ability) values
('cs100.instructor', 2, 'manage'),
('cs100.instructor', 2, 'interact');

insert into conversations (conversation_id, name) values (1, 'first conversation');

insert into messages (conversation_id, content, author_type, agent_id, user_id) values
(1, 'Hello, How can I help you?', 'agent', 1, null),
(1, 'Hey! can yuz help me w/ sum m@th', 'user', null, 'mathlover67@univ.edu'),
(1, 'I sure can try...', 'agent', 1, null);