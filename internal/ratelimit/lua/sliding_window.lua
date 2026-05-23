-- 滑动窗口限流(投产版)
--
-- KEYS[1] = ratelimit key,sorted set
--
-- ARGV[1] = window_ms   窗口长度(毫秒)
-- ARGV[2] = limit       阈值(0 = 不限,直接通过且不写入)
-- ARGV[3] = now_ms      调用方时钟(避免 Redis / app 时钟漂移)
-- ARGV[4] = cost        本次写入数量(>=0)
-- ARGV[5] = enforce     1 = gating(超额不写,返回 ok=0);0 = counting(超额也写,返回 ok=0)
-- ARGV[6] = nonce_base  调用方给的随机字符串,与循环 index 拼接保证成员唯一
--
-- 返回 {ok, count_after, limit, reset_ms, wrote}
--   ok = 1            通过
--   ok = 0            被拒 / 越界(enforce=1 时未写;enforce=0 时已写)
--   count_after       窗口内现有计数(写入后或被拒前)
--   limit             入参回显(0 = 不限)
--   reset_ms          最早成员的过期时刻(无成员 → now + window)
--   wrote             本次实际写入数量(被 enforce 拒时 = 0)

local key       = KEYS[1]
local window    = tonumber(ARGV[1])
local limit     = tonumber(ARGV[2])
local now       = tonumber(ARGV[3])
local cost      = tonumber(ARGV[4])
local enforce   = tonumber(ARGV[5])
local nonce     = ARGV[6]

-- limit = 0 表示该维度不限,直接通过 + 不写入
if limit <= 0 then
    return {1, 0, 0, now + window, 0}
end

-- cost < 0 防御
if cost < 0 then cost = 0 end

-- 清理过期成员(score 是 now_ms 写入时刻)
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

local count = redis.call('ZCARD', key)
local reset = now + window
if count > 0 then
    local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
    if oldest[2] then
        reset = tonumber(oldest[2]) + window
    end
end

-- gating 模式:超额不写,直接返回 0
if enforce == 1 and (count + cost) > limit then
    return {0, count, limit, reset, 0}
end

-- 写入 cost 个成员;成员名 = "now:nonce:i",保证唯一(同毫秒同 nonce 也不会撞)
local wrote = 0
if cost > 0 then
    local args = {}
    for i = 1, cost do
        table.insert(args, now)
        table.insert(args, tostring(now) .. ':' .. nonce .. ':' .. tostring(i))
    end
    redis.call('ZADD', key, unpack(args))
    wrote = cost
    -- TTL 续命:window + 1s,防 key 永驻
    redis.call('PEXPIRE', key, window + 1000)
end

local newCount = count + wrote
local ok = 1
if newCount > limit then
    ok = 0  -- counting 模式下超额标记,但已写入
end

return {ok, newCount, limit, reset, wrote}
