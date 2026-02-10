package limiter

import "context"

type RedisLimiter struct {
	// Redis implementation placeholder
}

func NewRedisLimiter(addr string) *RedisLimiter {
	return &RedisLimiter{}
}

func (r *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	// TODO: implement token-bucket in Redis (INCR+EXPIRE or Lua script)
	return true, nil
}