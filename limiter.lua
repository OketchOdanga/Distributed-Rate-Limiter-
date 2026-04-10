-- Add debug logging to Lua (optional, view via Redis MONITOR)
-- KEYS[1]: rate_limit:{user_id}
-- ARGV[1]: limit, ARGV[2]: window_size, ARGV[3]: now

local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local current_window = math.floor(now / window)
local prev_window = current_window - 1
local current_key = key .. ":" .. current_window
local prev_key = key .. ":" .. prev_window

local current_count = tonumber(redis.call('GET', current_key) or "0")
local prev_count = tonumber(redis.call('GET', prev_key) or "0")

local window_elapsed = now % window
local weight = (window - window_elapsed) / window
local weighted_prev = prev_count * weight
local total_count = current_count + weighted_prev

-- Debug: redis.log(redis.LOG_WARNING, string.format("DEBUG: curr=%d prev=%d weight=%.2f total=%.2f limit=%d", current_count, prev_count, weight, total_count, limit))

if total_count >= limit then
    return 0
end

redis.call('INCR', current_key)
redis.call('EXPIRE', current_key, window * 2)
return 1