package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const accountKeyLeasePrefix = "scheduling:key_leases:"
const accountLastScheduledPrefix = "scheduling:last_selected:"

// Each request has its own renewable member. Removing one request never removes
// another in-flight request belonging to the same key, even on another instance.
const accountKeyActivityLua = `
redis.replicate_commands()
local now = tonumber(redis.call('TIME')[1])
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', now - 90)
local members = redis.call('ZRANGE', KEYS[1], 0, -1)
local seen = {}
local others = 0
local own = 0
for _, member in ipairs(members) do
  local key = string.match(member, '^([^:]+):')
  if key == ARGV[1] then
    own = own + 1
  elseif key and not seen[key] then
    seen[key] = true
    others = others + 1
  end
end
`

var readAccountKeyActivityScript = redis.NewScript(accountKeyActivityLua + `
return {others, own, tonumber(redis.call('GET', KEYS[2]) or '0')}
`)

var acquireAccountKeyLeaseScript = redis.NewScript(accountKeyActivityLua + `
local expected = tonumber(ARGV[3])
if #KEYS == 2 and expected >= 0 and others > expected then return 0 end
for i = 3, #KEYS, 3 do
  local limit = tonumber(ARGV[4 + (i - 3) / 3])
  local concurrency = redis.call('ZCARD', KEYS[i+1]) + redis.call('ZCARD', KEYS[i+2])
  if limit <= 0 or concurrency < limit then
    redis.call('ZREMRANGEBYSCORE', KEYS[i], '-inf', now - 90)
    local peerSeen = {}
    local peerOthers = 0
    local peerOwn = 0
    for _, member in ipairs(redis.call('ZRANGE', KEYS[i], 0, -1)) do
      local key = string.match(member, '^([^:]+):')
      if key == ARGV[1] then
        peerOwn = peerOwn + 1
      elseif key and not peerSeen[key] then
        peerSeen[key] = true
        peerOthers = peerOthers + 1
      end
    end
    if peerOthers < others or (peerOthers == others and peerOwn > 0 and own == 0) then return 0 end
  end
end
redis.call('ZADD', KEYS[1], now, ARGV[1] .. ':' .. ARGV[2])
redis.call('EXPIRE', KEYS[1], 90)
redis.call('SET', KEYS[2], now, 'EX', 5184000)
return 1
`)

var refreshAccountKeyLeaseScript = redis.NewScript(`
redis.replicate_commands()
local member = ARGV[1] .. ':' .. ARGV[2]
if redis.call('ZSCORE', KEYS[1], member) then
  redis.call('ZADD', KEYS[1], tonumber(redis.call('TIME')[1]), member)
  redis.call('EXPIRE', KEYS[1], 90)
end
return 1
`)

func (c *gatewayCache) GetAccountKeyActivity(ctx context.Context, ids []string, keyID string) (map[string]service.AccountKeyActivity, error) {
	pipe := c.rdb.Pipeline()
	cmds := make([]*redis.Cmd, len(ids))
	for i, id := range ids {
		cmds[i] = readAccountKeyActivityScript.Eval(ctx, pipe, []string{accountKeyLeasePrefix + id, accountLastScheduledPrefix + id}, keyID)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	result := make(map[string]service.AccountKeyActivity, len(ids))
	for i, cmd := range cmds {
		values, err := cmd.Slice()
		if err != nil {
			return nil, err
		}
		if len(values) != 3 {
			return nil, fmt.Errorf("invalid account key activity response")
		}
		numbers := [3]int64{}
		for j, value := range values {
			numbers[j], err = strconv.ParseInt(fmt.Sprint(value), 10, 64)
			if err != nil {
				return nil, err
			}
		}
		activity := service.AccountKeyActivity{OtherKeys: int(numbers[0]), OwnRequests: int(numbers[1])}
		if numbers[2] > 0 {
			activity.LastScheduledAt = time.Unix(numbers[2], 0)
		}
		result[ids[i]] = activity
	}
	return result, nil
}

func (c *concurrencyCache) AcquireAccountKeyLease(ctx context.Context, accountID, keyID, requestID string, expected int, peers []service.AccountKeyPeer) (bool, error) {
	keys := []string{accountKeyLeasePrefix + accountID, accountLastScheduledPrefix + accountID}
	args := []any{keyID, requestID, expected}
	for _, peer := range peers {
		keys = append(keys, accountKeyLeasePrefix+peer.ID, accountSlotKeyPrefix+peer.ID, liveAccountSlotKeyPrefix+peer.ID)
		args = append(args, peer.MaxConcurrency)
	}
	value, err := acquireAccountKeyLeaseScript.Run(ctx, c.rdb, keys, args...).Int()
	return value == 1, err
}

func (c *concurrencyCache) RefreshAccountKeyLease(ctx context.Context, accountID, keyID, requestID string) error {
	return refreshAccountKeyLeaseScript.Run(ctx, c.rdb, []string{accountKeyLeasePrefix + accountID}, keyID, requestID).Err()
}

func (c *concurrencyCache) ReleaseAccountKeyLease(ctx context.Context, accountID, keyID, requestID string) error {
	return c.rdb.ZRem(ctx, accountKeyLeasePrefix+accountID, keyID+":"+requestID).Err()
}
