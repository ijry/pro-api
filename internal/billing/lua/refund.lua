-- refund.lua  全额退还
-- KEYS[1] = wallet:user:{id}
-- KEYS[2] = reservation:{rid}
-- 返回 {ok, refunded}  ok=1 成功; ok=-1 reservation 不存在

local reserved = tonumber(redis.call('GET', KEYS[2]) or '0')
if reserved <= 0 then return {-1, 0} end
redis.call('HINCRBY', KEYS[1], 'reserved', -reserved)
redis.call('HINCRBY', KEYS[1], 'balance', reserved)
redis.call('DEL', KEYS[2])
return {1, reserved}
