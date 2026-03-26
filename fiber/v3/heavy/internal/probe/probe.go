package probe

import (
	env "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/config"
	cache "github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/infra/cache"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/infra/db"
	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/internal/runtimestate"
)

func Live() bool {
	return true
}

func Started() bool {
	return runtimestate.Started()
}

func Ready() bool {
	if !runtimestate.Started() {
		return false
	}

	cfg := env.GetServerConfig()
	if cfg.DataBase.Enabled && !db.Orm.Ready() {
		return false
	}
	if cfg.Redis.Enabled && !cache.RedisReady() {
		return false
	}

	return true
}
