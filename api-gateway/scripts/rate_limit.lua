--
-- KEYS[1]  = redis key (e.g. "rl:ip:1.2.3.4")
-- ARGV[1]  = capacity   (số token tối đa = limit)
-- ARGV[2]  = window_sec (thời gian cửa sổ, giây)
-- ARGV[3]  = now_ms     (timestamp hiện tại, milliseconds)

local key         = KEYS[1]
local capacity    = tonumber(ARGV[1])
local window_sec  = tonumber(ARGV[2])
local now_ms      = tonumber(ARGV[3])

local refill_rate_ms = (window_sec * 1000) / capacity

local data = redis.call('HMGET', key, 'tokens', 'last_ms')
local tokens   = tonumber(data[1])
local last_ms  = tonumber(data[2])

if tokens == nil then
    tokens  = capacity - 1
    last_ms = now_ms

    redis.call('HSET', key, 'tokens', tokens, 'last_ms', last_ms)
    redis.call('EXPIRE', key, window_sec * 2)

    return {1, tokens, 0}
end

local elapsed_ms    = math.max(0, now_ms - last_ms)
local refill_tokens = math.floor(elapsed_ms / refill_rate_ms)

if refill_tokens > 0 then
    tokens  = math.min(capacity, tokens + refill_tokens)
    if tokens == capacity then
        last_ms = now_ms
    else
        last_ms = last_ms + (refill_tokens * refill_rate_ms)
    end
end

if tokens <= 0 then
    local wait_ms = math.ceil(refill_rate_ms - (elapsed_ms % refill_rate_ms))
    return {0, 0, wait_ms}
end

tokens = tokens - 1

redis.call('HSET', key, 'tokens', tokens, 'last_ms', last_ms)
redis.call('EXPIRE', key, window_sec * 2)

return {1, tokens, 0}