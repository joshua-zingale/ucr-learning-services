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

insert into conversations (conversation_id, user_id, name) values
(1, 'tom', 'conversation 1'),
(2, 'tom', 'conversation 2'),
(3, 'tom', 'conversation 3'),
(4, 'tom', 'conversation 4'),
(5, 'tom', 'conversation 5'),
(6, 'tom', 'conversation 6'),
(7, 'tom', 'conversation 7'),
(8, 'tom', 'conversation 8'),
(9, 'tom', 'conversation 9'),
(10, 'tom', 'conversation 10'),
(11, 'tom', 'conversation 11'),
(12, 'tom', 'conversation 12'),
(13, 'tom', 'conversation 13'),
(14, 'tom', 'conversation 14'),
(15, 'tom', 'conversation 15'),
(16, 'tom', 'conversation 16'),
(17, 'tom', 'conversation 17'),
(18, 'tom', 'conversation 18'),
(19, 'tom', 'conversation 19'),
(20, 'tom', 'conversation 20'),
(21, 'tom', 'conversation 21'),
(22, 'tom', 'conversation 22'),
(23, 'tom', 'conversation 23'),
(24, 'tom', 'conversation 24'),
(25, 'tom', 'conversation 25'),
(26, 'tom', 'conversation 26'),
(27, 'tom', 'conversation 27'),
(28, 'tom', 'conversation 28'),
(29, 'tom', 'conversation 29'),
(30, 'tom', 'conversation 30');

insert into messages (conversation_id, content, author_type, agent_id, user_id) values
(1, 'Hello, How can I help you?', 'agent', 1, null),
(1, 'Hey! can yuz help me w/ sum m@th', 'user', null, 'tom'),
(1, 'I sure can try...', 'agent', 1, null);