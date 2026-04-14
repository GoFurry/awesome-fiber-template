package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestRedisServiceLifecycleAndHelpers(t *testing.T) {
	server := miniredis.RunT(t)

	service, err := New(context.Background(), Config{
		Addr: server.Addr(),
		DB:   1,
	})
	if err != nil {
		t.Fatalf("new redis service failed: %v", err)
	}
	t.Cleanup(func() {
		if err := service.Close(); err != nil {
			t.Fatalf("close redis service failed: %v", err)
		}
	})

	if service.Raw() == nil {
		t.Fatalf("expected raw redis client")
	}

	if err := service.Ping(context.Background()); err != nil {
		t.Fatalf("ping redis failed: %v", err)
	}

	if ok, err := service.SetNX(context.Background(), "lock:user:1", "1", 5*time.Minute); err != nil || !ok {
		t.Fatalf("setnx failed: ok=%v err=%v", ok, err)
	}

	if err := service.Set(context.Background(), "plain:key", " value "); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	value, err := service.GetString(context.Background(), "plain:key")
	if err != nil {
		t.Fatalf("get string failed: %v", err)
	}
	if value != "value" {
		t.Fatalf("unexpected value: %q", value)
	}

	if err := service.HSetMap(context.Background(), "user:1", map[string]string{"name": "Alice", "status": "active"}); err != nil {
		t.Fatalf("hset map failed: %v", err)
	}
	if err := service.HSet(context.Background(), "user:1", "role", "admin"); err != nil {
		t.Fatalf("hset failed: %v", err)
	}

	role, err := service.HGet(context.Background(), "user:1", "role")
	if err != nil {
		t.Fatalf("hget failed: %v", err)
	}
	if role != "admin" {
		t.Fatalf("unexpected role: %q", role)
	}

	values, err := service.HMGet(context.Background(), "user:1", "name", "status")
	if err != nil {
		t.Fatalf("hmget failed: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("unexpected hmget size: %d", len(values))
	}

	hash, err := service.HGetAll(context.Background(), "user:1")
	if err != nil {
		t.Fatalf("hgetall failed: %v", err)
	}
	if hash["name"] != "Alice" {
		t.Fatalf("unexpected hash payload: %#v", hash)
	}

	deletedFields, err := service.HDel(context.Background(), "user:1", "role")
	if err != nil {
		t.Fatalf("hdel failed: %v", err)
	}
	if deletedFields != 1 {
		t.Fatalf("unexpected deleted field count: %d", deletedFields)
	}

	total, err := service.Incr(context.Background(), "counters:requests")
	if err != nil {
		t.Fatalf("incr failed: %v", err)
	}
	if total != 1 {
		t.Fatalf("unexpected incr result: %d", total)
	}

	if err := service.Set(context.Background(), "session:1", "active"); err != nil {
		t.Fatalf("set session failed: %v", err)
	}
	if err := service.Set(context.Background(), "session:2", "active"); err != nil {
		t.Fatalf("set session failed: %v", err)
	}

	count, err := service.CountByPrefix(context.Background(), "session:")
	if err != nil {
		t.Fatalf("count by prefix failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("unexpected count by prefix: %d", count)
	}

	keys, err := service.FindByPrefix(context.Background(), "session:")
	if err != nil {
		t.Fatalf("find by prefix failed: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("unexpected key count: %d", len(keys))
	}

	if err := service.DelByPrefix(context.Background(), "session:"); err != nil {
		t.Fatalf("delete by prefix failed: %v", err)
	}

	count, err = service.CountByPrefix(context.Background(), "session:")
	if err != nil {
		t.Fatalf("count after delete failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected session keys to be deleted, got %d", count)
	}

	if err := service.PipelineExec(context.Background(), func(pipe Pipeliner) {
		pipe.Set(context.Background(), "pipe:1", "ok", 0)
		pipe.Set(context.Background(), "pipe:2", "ok", 0)
	}); err != nil {
		t.Fatalf("pipeline exec failed: %v", err)
	}

	if err := service.Del(context.Background(), "pipe:1", "pipe:2"); err != nil {
		t.Fatalf("del failed: %v", err)
	}
}

func TestRedisNewValidation(t *testing.T) {
	if _, err := New(context.Background(), Config{}); err == nil {
		t.Fatalf("expected missing addr error")
	}
}
