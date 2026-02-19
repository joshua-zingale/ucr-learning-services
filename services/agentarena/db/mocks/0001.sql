insert into agent_classes(agent_class_id, slug, name, config_schema) values
(1, 'standard', 'Standard Agent', '{"$schema": "https://json-schema.org/draft/2020-12/schema", "type": "object", "properties": {"systemPrompt": {"type": "string"}}, "additionalProperties": false}');

insert into agents (agent_id, name, agent_class_id, config) values
(1, 'tombot', 1, '{"systemPrompt": "You are Tom. Speak with the user like you are an old uncle."}'),
(2, 'dickbot', 1, '{"systemPrompt": "You are Dick. Speak with the user like you are a good old pal who loves to play catch."}'),
(3, 'harrybot', 1, '{"systemPrompt": "Your father is the president of Osborn technology. Speak to the user like you know nothing of the green goblin, but drop hints."}');

insert into user_agent_permissions (user_id, agent_id, ability) values
('tom', 1, 'manage');
insert into group_agent_permissions (group_id, agent_id, ability) values
('cs100.instructor', 2, 'manage'),
('cs100.instructor', 2, 'interact');

insert into conversations (conversation_id, active_agent_id, user_id, name) values
(1, 1, 'tom', 'conversation 1'),
(2, 1, 'tom', 'conversation 2'),
(3, 1, 'tom', 'conversation 3'),
(4, 1, 'tom', 'conversation 4'),
(5, 1, 'tom', 'conversation 5'),
(6, 1, 'tom', 'conversation 6'),
(7, 1, 'tom', 'conversation 7'),
(8, 1, 'tom', 'conversation 8'),
(9, 1, 'tom', 'conversation 9'),
(10, 1, 'tom', 'conversation 10'),
(11, 1, 'tom', 'conversation 11'),
(12, 1, 'tom', 'conversation 12'),
(13, 1, 'tom', 'conversation 13'),
(14, 1, 'tom', 'conversation 14'),
(15, 1, 'tom', 'conversation 15'),
(16, 1, 'tom', 'conversation 16'),
(17, 1, 'tom', 'conversation 17'),
(18, 1, 'tom', 'conversation 18'),
(19, 1, 'tom', 'conversation 19'),
(20, 1, 'tom', 'conversation 20'),
(21, 1, 'tom', 'conversation 21'),
(22, 1, 'tom', 'conversation 22'),
(23, 1, 'tom', 'conversation 23'),
(24, 1, 'tom', 'conversation 24'),
(25, 1, 'tom', 'conversation 25'),
(26, 1, 'tom', 'conversation 26'),
(27, 1, 'tom', 'conversation 27'),
(28, 1, 'tom', 'conversation 28'),
(29, 1, 'tom', 'conversation 29'),
(30, 1, 'tom', 'conversation 30');

insert into messages (conversation_id, content, message_type, agent_id, user_id) values
(1, 'Hello, How can I help you?', 'agent', 1, null),
(1, 'Hey! can yuz help me w/ sum m@th', 'user', null, 'tom'),
(1, 'I sure can try...', 'agent', 1, null);