// pkg/lock/distributed_lock.go
package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type DistributedLock struct {
	redis *redis.Client
}

// NewDistributedLock cria uma nova instância do lock distribuído
func NewDistributedLock(redis *redis.Client) *DistributedLock {
	return &DistributedLock{redis: redis}
}

// AcquireLock tenta adquirir um lock com TTL.
// Retorna o valor único do lock (para liberação segura) ou erro.
func (l *DistributedLock) AcquireLock(ctx context.Context, resource string, ttl time.Duration) (string, error) {
	lockKey := fmt.Sprintf("lock:%s", resource)
	lockValue := uuid.New().String()

	// SET key value NX PX milliseconds
	ok, err := l.redis.SetNX(ctx, lockKey, lockValue, ttl).Result()
	if err != nil {
		return "", fmt.Errorf("erro ao tentar adquirir lock: %w", err)
	}
	if !ok {
		return "", fmt.Errorf("lock já adquirido por outro processo: %s", resource)
	}

	return lockValue, nil
}

// ReleaseLock libera o lock de forma atômica usando Lua Script
// Só libera se o valor for exatamente o mesmo que foi adquirido
func (l *DistributedLock) ReleaseLock(ctx context.Context, resource, lockValue string) error {
	lockKey := fmt.Sprintf("lock:%s", resource)

	// Script Lua: só deleta se o valor bater exatamente
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.redis.Eval(ctx, script, []string{lockKey}, lockValue).Result()
	if err != nil {
		return fmt.Errorf("erro ao executar script de liberação: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("lock não pertence a este cliente ou já expirou")
	}

	return nil
}

// ExtendLock estende a expiração do lock (útil em operações longas)
func (l *DistributedLock) ExtendLock(ctx context.Context, resource, lockValue string, ttl time.Duration) error {
	lockKey := fmt.Sprintf("lock:%s", resource)

	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("EXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := l.redis.Eval(ctx, script, []string{lockKey}, lockValue, ttl.Seconds()).Result()
	if err != nil {
		return fmt.Errorf("erro ao estender lock: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("não foi possível estender: lock não pertence a este cliente")
	}

	return nil
}