-- commit.lua  提交实际用量
-- KEYS[1] = wallet:user:{id}
-- KEYS[2] = reservation:{rid}
-- ARGV[1] = actual_cost
-- 返回 {ok, refund}  ok=1 成功; ok=-1 reservation 不存在; ok=-2 actual > reserved

local reserved = tonumber(redis.call('GET', KEYS[2]) or '0')
if reserved <= 0 then return {-1, 0} end
local actual = tonumber(ARGV[1])
if actual > reserved then return {-2, 0} end
local refund = reserved - actual
redis.call('HINCRBY', KEYS[1], 'reserved', -reserved)
if refund > 0 then redis.call('HINCRBY', KEYS[1], 'balance', refund) end
redis.call('HINCRBY', KEYS[1], 'consumed', actual)
redis.call('DEL', KEYS[2])
return {1, refund}
