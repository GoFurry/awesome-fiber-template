package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Config struct {
	Addr     string
	Username string
	Password string
	DB       int
	PoolSize int
}

type Service struct {
	cfg Config
	raw *goredis.Client
}

type Pipeliner = goredis.Pipeliner

func New(ctx context.Context, cfg Config) (*Service, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:      normalized.Addr,
		Username:  normalized.Username,
		Password:  normalized.Password,
		DB:        normalized.DB,
		PoolSize:  normalized.PoolSize,
		OnConnect: func(ctx context.Context, cn *goredis.Conn) error { return nil },
	})

	service := &Service{
		cfg: normalized,
		raw: client,
	}

	if err := service.Ping(ctx); err != nil {
		_ = service.Close()
		return nil, fmt.Errorf("ping redis failed: %w", err)
	}

	return service, nil
}

func (s *Service) Raw() *goredis.Client {
	if s == nil {
		return nil
	}
	return s.raw
}

func (s *Service) Ping(ctx context.Context) error {
	if s == nil || s.raw == nil {
		return errors.New("redis service is not initialized")
	}
	return s.raw.Ping(ctxOrBackground(ctx)).Err()
}

func (s *Service) Close() error {
	if s == nil || s.raw == nil {
		return nil
	}
	err := s.raw.Close()
	s.raw = nil
	return err
}

func (s *Service) Del(ctx context.Context, keys ...string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return s.raw.Del(ctxOrBackground(ctx), keys...).Err()
}

func (s *Service) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	if err := s.ensureReady(); err != nil {
		return false, err
	}
	return s.raw.SetNX(ctxOrBackground(ctx), key, value, expiration).Result()
}

func (s *Service) Set(ctx context.Context, key string, value any) error {
	return s.SetExpire(ctx, key, value, 0)
}

func (s *Service) SetExpire(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	return s.raw.Set(ctxOrBackground(ctx), key, value, expiration).Err()
}

func (s *Service) GetString(ctx context.Context, key string) (string, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}

	value, err := s.raw.Get(ctxOrBackground(ctx), key).Result()
	switch {
	case errors.Is(err, goredis.Nil):
		return "", nil
	case err != nil:
		return "", err
	default:
		return strings.TrimSpace(value), nil
	}
}

func (s *Service) HSetMap(ctx context.Context, key string, values map[string]string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	return s.raw.HSet(ctxOrBackground(ctx), key, values).Err()
}

func (s *Service) HSet(ctx context.Context, key string, fieldName string, fieldValue string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	return s.raw.HSet(ctxOrBackground(ctx), key, fieldName, fieldValue).Err()
}

func (s *Service) HGet(ctx context.Context, key string, fieldName string) (string, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}

	value, err := s.raw.HGet(ctxOrBackground(ctx), key, fieldName).Result()
	switch {
	case errors.Is(err, goredis.Nil):
		return "", fmt.Errorf("redis hash field not found: %s.%s", key, fieldName)
	case err != nil:
		return "", err
	default:
		return value, nil
	}
}

func (s *Service) HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	value, err := s.raw.HMGet(ctxOrBackground(ctx), key, fields...).Result()
	switch {
	case errors.Is(err, goredis.Nil):
		return nil, errors.New("redis hash not found")
	case err != nil:
		return nil, err
	default:
		return value, nil
	}
}

func (s *Service) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	value, err := s.raw.HGetAll(ctxOrBackground(ctx), key).Result()
	switch {
	case errors.Is(err, goredis.Nil):
		return nil, errors.New("redis hash not found")
	case err != nil:
		return nil, err
	default:
		return value, nil
	}
}

func (s *Service) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	if err := s.ensureReady(); err != nil {
		return 0, err
	}

	value, err := s.raw.HDel(ctxOrBackground(ctx), key, fields...).Result()
	switch {
	case errors.Is(err, goredis.Nil):
		return 0, errors.New("redis hash not found")
	case err != nil:
		return 0, err
	default:
		return value, nil
	}
}

func (s *Service) Incr(ctx context.Context, key string) (int64, error) {
	if err := s.ensureReady(); err != nil {
		return 0, err
	}
	return s.raw.Incr(ctxOrBackground(ctx), key).Result()
}

func (s *Service) CountByPrefix(ctx context.Context, prefix string) (int64, error) {
	if err := s.ensureReady(); err != nil {
		return 0, err
	}

	var (
		cursor uint64
		count  int
	)
	pattern := prefix + "*"

	for {
		keys, nextCursor, err := s.raw.Scan(ctxOrBackground(ctx), cursor, pattern, 100).Result()
		if err != nil {
			return 0, err
		}

		count += len(keys)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return int64(count), nil
}

func (s *Service) FindByPrefix(ctx context.Context, prefix string) ([]string, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}

	var (
		cursor  uint64
		results []string
	)
	pattern := prefix + "*"

	for {
		keys, nextCursor, err := s.raw.Scan(ctxOrBackground(ctx), cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		results = append(results, keys...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return results, nil
}

func (s *Service) DelByPrefix(ctx context.Context, prefix string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}

	keys, err := s.FindByPrefix(ctx, prefix)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return s.Del(ctx, keys...)
}

func (s *Service) PipelineExec(ctx context.Context, fn func(pipe goredis.Pipeliner)) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if fn == nil {
		return errors.New("redis pipeline function is required")
	}

	pipe := s.raw.Pipeline()
	fn(pipe)
	_, err := pipe.Exec(ctxOrBackground(ctx))
	return err
}

func (s *Service) ensureReady() error {
	if s == nil || s.raw == nil {
		return errors.New("redis service is not initialized")
	}
	return nil
}

func normalizeConfig(cfg Config) (Config, error) {
	normalized := cfg
	normalized.Addr = strings.TrimSpace(normalized.Addr)
	normalized.Username = strings.TrimSpace(normalized.Username)
	normalized.Password = strings.TrimSpace(normalized.Password)

	if normalized.Addr == "" {
		return Config{}, errors.New("redis addr is required")
	}
	if normalized.PoolSize < 0 {
		return Config{}, errors.New("redis pool size cannot be negative")
	}
	return normalized, nil
}

func ctxOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
