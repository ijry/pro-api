INSERT INTO model_catalogs (id, name, family, capabilities, default_input_ratio, default_output_ratio,
    default_cached_ratio, default_reasoning_ratio, max_input_tokens, status, owned_by, created_at, updated_at) VALUES
-- OpenAI
(1,  'gpt-4o',                     'chat',  JSON_ARRAY('chat','stream','vision','tool_use'),  2.5,  10.0, 1.25, NULL,  128000, 0, 'openai',    NOW(3), NOW(3)),
(2,  'gpt-4o-mini',                'chat',  JSON_ARRAY('chat','stream','vision','tool_use'),  0.15,  0.6, 0.075,NULL,  128000, 0, 'openai',    NOW(3), NOW(3)),
(3,  'gpt-4-turbo',                'chat',  JSON_ARRAY('chat','stream','vision','tool_use'), 10.0,  30.0, NULL, NULL,  128000, 0, 'openai',    NOW(3), NOW(3)),
(4,  'gpt-4',                      'chat',  JSON_ARRAY('chat','stream','tool_use'),          30.0,  60.0, NULL, NULL,    8192, 0, 'openai',    NOW(3), NOW(3)),
(5,  'gpt-3.5-turbo',              'chat',  JSON_ARRAY('chat','stream','tool_use'),           0.5,   1.5, NULL, NULL,   16384, 0, 'openai',    NOW(3), NOW(3)),
(6,  'o1',                         'chat',  JSON_ARRAY('chat','reasoning'),                 15.0,  60.0,  7.5, 60.0,  200000, 0, 'openai',    NOW(3), NOW(3)),
(7,  'o1-mini',                    'chat',  JSON_ARRAY('chat','reasoning'),                  3.0,  12.0,  1.5, 12.0,  128000, 0, 'openai',    NOW(3), NOW(3)),
(8,  'text-embedding-3-large',     'embed', JSON_ARRAY('embed'),                             0.13,   0.0, NULL, NULL,    8191, 0, 'openai',    NOW(3), NOW(3)),
(9,  'text-embedding-3-small',     'embed', JSON_ARRAY('embed'),                             0.02,   0.0, NULL, NULL,    8191, 0, 'openai',    NOW(3), NOW(3)),
-- Anthropic
(10, 'claude-3-5-sonnet-20241022', 'chat',  JSON_ARRAY('chat','stream','vision','tool_use'),  3.0,  15.0,  0.3, NULL,  200000, 0, 'anthropic', NOW(3), NOW(3)),
(11, 'claude-3-5-sonnet-latest',   'chat',  JSON_ARRAY('chat','stream','vision','tool_use'),  3.0,  15.0,  0.3, NULL,  200000, 0, 'anthropic', NOW(3), NOW(3)),
(12, 'claude-3-5-haiku-20241022',  'chat',  JSON_ARRAY('chat','stream','tool_use'),           0.8,   4.0, 0.08, NULL,  200000, 0, 'anthropic', NOW(3), NOW(3)),
(13, 'claude-3-opus-20240229',     'chat',  JSON_ARRAY('chat','stream','vision','tool_use'), 15.0,  75.0,  1.5, NULL,  200000, 0, 'anthropic', NOW(3), NOW(3)),
-- Gemini
(14, 'gemini-2.0-flash-exp',       'chat',  JSON_ARRAY('chat','stream','vision'),             0.0,   0.0, NULL, NULL, 1048576, 0, 'google',    NOW(3), NOW(3)),
(15, 'gemini-1.5-pro',             'chat',  JSON_ARRAY('chat','stream','vision','tool_use'),  1.25,  5.0, NULL, NULL, 2097152, 0, 'google',    NOW(3), NOW(3)),
(16, 'gemini-1.5-flash',           'chat',  JSON_ARRAY('chat','stream','vision','tool_use'),  0.075, 0.3, NULL, NULL, 1048576, 0, 'google',    NOW(3), NOW(3)),
(17, 'text-embedding-004',         'embed', JSON_ARRAY('embed'),                              0.0,   0.0, NULL, NULL,    2048, 0, 'google',    NOW(3), NOW(3)),
-- DeepSeek
(18, 'deepseek-chat',              'chat',  JSON_ARRAY('chat','stream','tool_use'),           0.14,  0.28,0.014, NULL,  65536, 0, 'deepseek',  NOW(3), NOW(3)),
(19, 'deepseek-reasoner',          'chat',  JSON_ARRAY('chat','stream','reasoning'),          0.55,  2.19, 0.14, 2.19,  65536, 0, 'deepseek',  NOW(3), NOW(3)),
-- Moonshot
(20, 'moonshot-v1-8k',             'chat',  JSON_ARRAY('chat','stream','tool_use'),           1.68,  1.68, NULL, NULL,    8192, 0, 'moonshot',  NOW(3), NOW(3)),
(21, 'moonshot-v1-32k',            'chat',  JSON_ARRAY('chat','stream','tool_use'),           3.36,  3.36, NULL, NULL,   32768, 0, 'moonshot',  NOW(3), NOW(3)),
(22, 'moonshot-v1-128k',           'chat',  JSON_ARRAY('chat','stream','tool_use'),           8.4,   8.4,  NULL, NULL,  131072, 0, 'moonshot',  NOW(3), NOW(3)),
-- Zhipu
(23, 'glm-4-plus',                 'chat',  JSON_ARRAY('chat','stream','vision','tool_use'),  7.0,   7.0,  NULL, NULL,  128000, 0, 'zhipu',     NOW(3), NOW(3)),
(24, 'glm-4-air',                  'chat',  JSON_ARRAY('chat','stream','tool_use'),           0.14,  0.14, NULL, NULL,  128000, 0, 'zhipu',     NOW(3), NOW(3)),
(25, 'glm-4-flash',                'chat',  JSON_ARRAY('chat','stream','tool_use'),           0.0,   0.0,  NULL, NULL,  128000, 0, 'zhipu',     NOW(3), NOW(3)),
(26, 'embedding-3',                'embed', JSON_ARRAY('embed'),                              0.07,  0.0,  NULL, NULL,    8192, 0, 'zhipu',     NOW(3), NOW(3)),
-- Qwen
(27, 'qwen-max',                   'chat',  JSON_ARRAY('chat','stream','tool_use'),           2.8,   8.4,  NULL, NULL,   32768, 0, 'qwen',      NOW(3), NOW(3)),
(28, 'qwen-plus',                  'chat',  JSON_ARRAY('chat','stream','tool_use'),           0.56,  1.68, NULL, NULL,  131072, 0, 'qwen',      NOW(3), NOW(3)),
(29, 'qwen-turbo',                 'chat',  JSON_ARRAY('chat','stream','tool_use'),           0.14,  0.28, NULL, NULL, 1008000, 0, 'qwen',      NOW(3), NOW(3)),
(30, 'text-embedding-v3',          'embed', JSON_ARRAY('embed'),                              0.49,  0.0,  NULL, NULL,    8192, 0, 'qwen',      NOW(3), NOW(3)),
-- Doubao
(31, 'doubao-pro-32k',             'chat',  JSON_ARRAY('chat','stream','tool_use'),           0.8,   2.0,  NULL, NULL,   32768, 0, 'doubao',    NOW(3), NOW(3)),
(32, 'doubao-pro-128k',            'chat',  JSON_ARRAY('chat','stream','tool_use'),           5.0,  10.0,  NULL, NULL,  131072, 0, 'doubao',    NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE family = VALUES(family);
