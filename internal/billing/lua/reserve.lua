-- reserve.lua  预扣 quota
-- KEYS[1] = wallet:user:{id}     hash {balance, reserved, consumed}
-- KEYS[2] = reservation:{rid}    string value=reserved_quota
-- ARGV[1] = est_cost
-- ARGV[2] = ttl_seconds
-- 返回 {ok, balance_after}  ok=1 成功; ok=0 余额不足

local balance = tonumber(redis.call('HGET', KEYS[1], 'balance') or '0')
local est = tonumber(ARGV[1])
if balance < est then
    return {0, balance}
end
redis.call('HINCRBY', KEYS[1], 'balance', -est)
redis.call('HINCRBY', KEYS[1], 'reserved', est)
redis.call('SET', KEYS[2], est, 'EX', tonumber(ARGV[2]))
return {1, balance - est}
