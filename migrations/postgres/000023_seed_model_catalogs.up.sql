INSERT INTO model_catalogs (id, name, family, capabilities, default_input_ratio, default_output_ratio,
    default_cached_ratio, default_reasoning_ratio, max_input_tokens, status, owned_by, created_at, updated_at) VALUES
-- OpenAI
(1,  'gpt-4o',                     'chat',  '["chat","stream","vision","tool_use"]'::jsonb,   2.5,  10.0, 1.25, NULL,  128000, 0, 'openai',    NOW(), NOW()),
(2,  'gpt-4o-mini',                'chat',  '["chat","stream","vision","tool_use"]'::jsonb,   0.15,  0.6, 0.075,NULL,  128000, 0, 'openai',    NOW(), NOW()),
(3,  'gpt-4-turbo',                'chat',  '["chat","stream","vision","tool_use"]'::jsonb,  10.0,  30.0, NULL, NULL,  128000, 0, 'openai',    NOW(), NOW()),
(4,  'gpt-4',                      'chat',  '["chat","stream","tool_use"]'::jsonb,            30.0,  60.0, NULL, NULL,    8192, 0, 'openai',    NOW(), NOW()),
(5,  'gpt-3.5-turbo',              'chat',  '["chat","stream","tool_use"]'::jsonb,             0.5,   1.5, NULL, NULL,   16384, 0, 'openai',    NOW(), NOW()),
(6,  'o1',                         'chat',  '["chat","reasoning"]'::jsonb,                   15.0,  60.0,  7.5, 60.0,  200000, 0, 'openai',    NOW(), NOW()),
(7,  'o1-mini',                    'chat',  '["chat","reasoning"]'::jsonb,                    3.0,  12.0,  1.5, 12.0,  128000, 0, 'openai',    NOW(), NOW()),
(8,  'text-embedding-3-large',     'embed', '["embed"]'::jsonb,                               0.13,   0.0, NULL, NULL,    8191, 0, 'openai',    NOW(), NOW()),
(9,  'text-embedding-3-small',     'embed', '["embed"]'::jsonb,                               0.02,   0.0, NULL, NULL,    8191, 0, 'openai',    NOW(), NOW()),
-- Anthropic
(10, 'claude-3-5-sonnet-20241022', 'chat',  '["chat","stream","vision","tool_use"]'::jsonb,   3.0,  15.0,  0.3, NULL,  200000, 0, 'anthropic', NOW(), NOW()),
(11, 'claude-3-5-sonnet-latest',   'chat',  '["chat","stream","vision","tool_use"]'::jsonb,   3.0,  15.0,  0.3, NULL,  200000, 0, 'anthropic', NOW(), NOW()),
(12, 'claude-3-5-haiku-20241022',  'chat',  '["chat","stream","tool_use"]'::jsonb,             0.8,   4.0, 0.08, NULL,  200000, 0, 'anthropic', NOW(), NOW()),
(13, 'claude-3-opus-20240229',     'chat',  '["chat","stream","vision","tool_use"]'::jsonb,  15.0,  75.0,  1.5, NULL,  200000, 0, 'anthropic', NOW(), NOW()),
-- Gemini
(14, 'gemini-2.0-flash-exp',       'chat',  '["chat","stream","vision"]'::jsonb,              0.0,   0.0, NULL, NULL, 1048576, 0, 'google',    NOW(), NOW()),
(15, 'gemini-1.5-pro',             'chat',  '["chat","stream","vision","tool_use"]'::jsonb,   1.25,  5.0, NULL, NULL, 2097152, 0, 'google',    NOW(), NOW()),
(16, 'gemini-1.5-flash',           'chat',  '["chat","stream","vision","tool_use"]'::jsonb,   0.075, 0.3, NULL, NULL, 1048576, 0, 'google',    NOW(), NOW()),
(17, 'text-embedding-004',         'embed', '["embed"]'::jsonb,                               0.0,   0.0, NULL, NULL,    2048, 0, 'google',    NOW(), NOW()),
-- DeepSeek
(18, 'deepseek-chat',              'chat',  '["chat","stream","tool_use"]'::jsonb,             0.14,  0.28,0.014, NULL,  65536, 0, 'deepseek',  NOW(), NOW()),
(19, 'deepseek-reasoner',          'chat',  '["chat","stream","reasoning"]'::jsonb,            0.55,  2.19, 0.14, 2.19,  65536, 0, 'deepseek',  NOW(), NOW()),
-- Moonshot
(20, 'moonshot-v1-8k',             'chat',  '["chat","stream","tool_use"]'::jsonb,             1.68,  1.68, NULL, NULL,    8192, 0, 'moonshot',  NOW(), NOW()),
(21, 'moonshot-v1-32k',            'chat',  '["chat","stream","tool_use"]'::jsonb,             3.36,  3.36, NULL, NULL,   32768, 0, 'moonshot',  NOW(), NOW()),
(22, 'moonshot-v1-128k',           'chat',  '["chat","stream","tool_use"]'::jsonb,             8.4,   8.4,  NULL, NULL,  131072, 0, 'moonshot',  NOW(), NOW()),
-- Zhipu
(23, 'glm-4-plus',                 'chat',  '["chat","stream","vision","tool_use"]'::jsonb,   7.0,   7.0,  NULL, NULL,  128000, 0, 'zhipu',     NOW(), NOW()),
(24, 'glm-4-air',                  'chat',  '["chat","stream","tool_use"]'::jsonb,             0.14,  0.14, NULL, NULL,  128000, 0, 'zhipu',     NOW(), NOW()),
(25, 'glm-4-flash',                'chat',  '["chat","stream","tool_use"]'::jsonb,             0.0,   0.0,  NULL, NULL,  128000, 0, 'zhipu',     NOW(), NOW()),
(26, 'embedding-3',                'embed', '["embed"]'::jsonb,                               0.07,  0.0,  NULL, NULL,    8192, 0, 'zhipu',     NOW(), NOW()),
-- Qwen
(27, 'qwen-max',                   'chat',  '["chat","stream","tool_use"]'::jsonb,             2.8,   8.4,  NULL, NULL,   32768, 0, 'qwen',      NOW(), NOW()),
(28, 'qwen-plus',                  'chat',  '["chat","stream","tool_use"]'::jsonb,             0.56,  1.68, NULL, NULL,  131072, 0, 'qwen',      NOW(), NOW()),
(29, 'qwen-turbo',                 'chat',  '["chat","stream","tool_use"]'::jsonb,             0.14,  0.28, NULL, NULL, 1008000, 0, 'qwen',      NOW(), NOW()),
(30, 'text-embedding-v3',          'embed', '["embed"]'::jsonb,                               0.49,  0.0,  NULL, NULL,    8192, 0, 'qwen',      NOW(), NOW()),
-- Doubao
(31, 'doubao-pro-32k',             'chat',  '["chat","stream","tool_use"]'::jsonb,             0.8,   2.0,  NULL, NULL,   32768, 0, 'doubao',    NOW(), NOW()),
(32, 'doubao-pro-128k',            'chat',  '["chat","stream","tool_use"]'::jsonb,             5.0,  10.0,  NULL, NULL,  131072, 0, 'doubao',    NOW(), NOW()),
-- Grok Build
(33, 'grok-4',                     'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-build', NOW(), NOW()),
(34, 'grok-3',                     'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-build', NOW(), NOW()),
(35, 'grok-3-mini',                'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-build', NOW(), NOW()),
(36, 'grok-3-mini-fast',           'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-build', NOW(), NOW()),
-- Grok Web
(37, 'grok-web/grok-3',            'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(38, 'grok-web/grok-3-mini',       'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(39, 'grok-web/grok-3-thinking',   'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(40, 'grok-web/grok-4',            'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(41, 'grok-web/grok-4-mini',       'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(42, 'grok-web/grok-4-thinking',   'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(43, 'grok-web/grok-4-heavy',      'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(44, 'grok-web/grok-4.1-mini',     'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(45, 'grok-web/grok-4.1-fast',     'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(46, 'grok-web/grok-4.1-expert',   'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW()),
(47, 'grok-web/grok-4.1-thinking', 'chat',  '["chat","stream"]'::jsonb,                         0.0,   0.0,  NULL, NULL,  131072, 0, 'grok-web',   NOW(), NOW())
ON CONFLICT (name) DO UPDATE SET family = EXCLUDED.family;
