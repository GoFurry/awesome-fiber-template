package env

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GoFurry/awesome-go-template/fiber/v3/heavy/pkg/common"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var (
	configuration *serverConfig
	configErr     error
	configOnce    sync.Once
	configOptions = configLoaderOptions{
		projectName: common.COMMON_PROJECT_NAME,
		fileName:    "server.yaml",
	}
	configOptionsMu sync.Mutex
)

type configLoaderOptions struct {
	projectName string
	fileName    string
	configFile  string
}

type serverConfig struct {
	ClusterId  int              `yaml:"cluster_id"`
	Server     ServerConfig     `yaml:"server"`
	Key        KeyConfig        `yaml:"key"`
	DataBase   DataBaseConfig   `yaml:"database"`
	Log        LogConfig        `yaml:"log"`
	Redis      RedisConfig      `yaml:"redis"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	Waf        WafConfig        `yaml:"waf"`
	Proxy      ProxyConfig      `yaml:"proxy"`
	Auth       AuthConfig       `yaml:"auth"`
	Prometheus PrometheusConfig `yaml:"prometheus"`
	Schedule   ScheduleConfig   `yaml:"schedule"`
}

type ServerConfigHolder = serverConfig

type PrometheusConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Url            string   `yaml:"url"`
	AuthName       string   `yaml:"auth_name"`
	AuthPass       string   `yaml:"auth_pass"`
	Namespace      string   `yaml:"namespace"`
	Path           string   `yaml:"path"`
	ServiceMetrics []string `yaml:"service_metrics"`
}

type ScheduleConfig struct {
	Enabled bool `yaml:"enabled"`
}

type AuthConfig struct {
	AuthSalt  string `yaml:"auth_salt"`
	JwtSecret string `yaml:"jwt_secret"`
}

type ProxyConfig struct {
	Url  string `yaml:"url"`
	Name string `yaml:"name"`
	Pass string `yaml:"pass"`
}

type WafConfig struct {
	Enabled  bool     `yaml:"enabled"`
	ConfPath []string `yaml:"conf_path"`
}

type MiddlewareConfig struct {
	Swagger SwaggerConfig `yaml:"swagger"`
	Cors    CorsConfig    `yaml:"cors"`
	Limiter LimiterConfig `yaml:"limiter"`
}

type LimiterConfig struct {
	Enabled     bool          `yaml:"enabled"`
	MaxRequests int           `yaml:"max_requests"`
	Expiration  time.Duration `yaml:"expiration"`
}

type CorsConfig struct {
	AllowOrigins []string `yaml:"allow_origins"`
}

type SwaggerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	FilePath string `yaml:"file_path"`
	BasePath string `yaml:"base_path"`
	Path     string `yaml:"path"`
	Title    string `yaml:"title"`
}

type RedisConfig struct {
	Enabled       bool   `yaml:"enabled"`
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
}

type LogConfig struct {
	LogLevel      string `yaml:"log_level"`
	LogMode       string `yaml:"log_mode"`
	LogPath       string `yaml:"log_path"`
	LogMaxSize    int    `yaml:"log_max_size"`
	LogMaxBackups int    `yaml:"log_max_backups"`
	LogMaxAge     int    `yaml:"log_max_age"`
}

type DataBaseConfig struct {
	Enabled     bool                 `yaml:"enabled"`
	AutoMigrate bool                 `yaml:"auto_migrate"`
	DBType      string               `yaml:"db_type"`
	SQLite      SQLiteDataBaseConfig `yaml:"sqlite"`
	Postgres    SQLDataBaseConfig    `yaml:"postgres"`
	MySQL       SQLDataBaseConfig    `yaml:"mysql"`
	DSN         string               `yaml:"dsn"`
	DBName      string               `yaml:"db_name"`
	DBHost      string               `yaml:"db_host"`
	DBPort      string               `yaml:"db_port"`
	DBUser      string               `yaml:"db_username"`
	DBPass      string               `yaml:"db_password"`
	SQLPath     string               `yaml:"sqlite_path"`
}

type SQLDataBaseConfig struct {
	DSN    string `yaml:"dsn"`
	DBName string `yaml:"db_name"`
	DBHost string `yaml:"db_host"`
	DBPort string `yaml:"db_port"`
	DBUser string `yaml:"db_username"`
	DBPass string `yaml:"db_password"`
}

type SQLiteDataBaseConfig struct {
	DSN  string `yaml:"dsn"`
	Path string `yaml:"path"`
}

type ServerConfig struct {
	AppID         string `yaml:"app_id"`
	AppName       string `yaml:"app_name"`
	AppVersion    string `yaml:"app_version"`
	Mode          string `yaml:"mode"`
	IPAddress     string `yaml:"ip_address"`
	Port          string `yaml:"port"`
	MemoryLimit   int    `yaml:"memory_limit"`
	GCPercent     int    `yaml:"gc_percent"`
	Network       string `yaml:"network"`
	EnablePrefork bool   `yaml:"enable_prefork"`
	IsFullStack   bool   `yaml:"is_full_stack"`
}

type KeyConfig struct {
	TLSKey       string `yaml:"tls_key"`
	TLSPem       string `yaml:"tls_pem"`
	LoginPrivate string `yaml:"login_private"`
	LoginPublic  string `yaml:"login_public"`
}

type configKey struct {
	name string
	kind string
}

var knownConfigKeys = []configKey{
	{name: "cluster_id", kind: "int"},
	{name: "server.app_id", kind: "string"},
	{name: "server.app_name", kind: "string"},
	{name: "server.app_version", kind: "string"},
	{name: "server.mode", kind: "string"},
	{name: "server.ip_address", kind: "string"},
	{name: "server.port", kind: "string"},
	{name: "server.memory_limit", kind: "int"},
	{name: "server.gc_percent", kind: "int"},
	{name: "server.network", kind: "string"},
	{name: "server.enable_prefork", kind: "bool"},
	{name: "server.is_full_stack", kind: "bool"},
	{name: "key.tls_key", kind: "string"},
	{name: "key.tls_pem", kind: "string"},
	{name: "key.login_private", kind: "string"},
	{name: "key.login_public", kind: "string"},
	{name: "database.enabled", kind: "bool"},
	{name: "database.auto_migrate", kind: "bool"},
	{name: "database.db_type", kind: "string"},
	{name: "database.sqlite.dsn", kind: "string"},
	{name: "database.sqlite.path", kind: "string"},
	{name: "database.postgres.dsn", kind: "string"},
	{name: "database.postgres.db_name", kind: "string"},
	{name: "database.postgres.db_host", kind: "string"},
	{name: "database.postgres.db_port", kind: "string"},
	{name: "database.postgres.db_username", kind: "string"},
	{name: "database.postgres.db_password", kind: "string"},
	{name: "database.mysql.dsn", kind: "string"},
	{name: "database.mysql.db_name", kind: "string"},
	{name: "database.mysql.db_host", kind: "string"},
	{name: "database.mysql.db_port", kind: "string"},
	{name: "database.mysql.db_username", kind: "string"},
	{name: "database.mysql.db_password", kind: "string"},
	{name: "database.dsn", kind: "string"},
	{name: "database.db_name", kind: "string"},
	{name: "database.db_host", kind: "string"},
	{name: "database.db_port", kind: "string"},
	{name: "database.db_username", kind: "string"},
	{name: "database.db_password", kind: "string"},
	{name: "database.sqlite_path", kind: "string"},
	{name: "log.log_level", kind: "string"},
	{name: "log.log_mode", kind: "string"},
	{name: "log.log_path", kind: "string"},
	{name: "log.log_max_size", kind: "int"},
	{name: "log.log_max_backups", kind: "int"},
	{name: "log.log_max_age", kind: "int"},
	{name: "redis.enabled", kind: "bool"},
	{name: "redis.redis_addr", kind: "string"},
	{name: "redis.redis_password", kind: "string"},
	{name: "middleware.swagger.enabled", kind: "bool"},
	{name: "middleware.swagger.file_path", kind: "string"},
	{name: "middleware.swagger.base_path", kind: "string"},
	{name: "middleware.swagger.path", kind: "string"},
	{name: "middleware.swagger.title", kind: "string"},
	{name: "middleware.cors.allow_origins", kind: "string_slice"},
	{name: "middleware.limiter.enabled", kind: "bool"},
	{name: "middleware.limiter.max_requests", kind: "int"},
	{name: "middleware.limiter.expiration", kind: "int"},
	{name: "waf.enabled", kind: "bool"},
	{name: "waf.conf_path", kind: "string_slice"},
	{name: "proxy.url", kind: "string"},
	{name: "proxy.name", kind: "string"},
	{name: "proxy.pass", kind: "string"},
	{name: "auth.auth_salt", kind: "string"},
	{name: "auth.jwt_secret", kind: "string"},
	{name: "prometheus.enabled", kind: "bool"},
	{name: "prometheus.url", kind: "string"},
	{name: "prometheus.auth_name", kind: "string"},
	{name: "prometheus.auth_pass", kind: "string"},
	{name: "prometheus.namespace", kind: "string"},
	{name: "prometheus.path", kind: "string"},
	{name: "prometheus.service_metrics", kind: "string_slice"},
	{name: "schedule.enabled", kind: "bool"},
}

func ConfigureServerConfig(projectName, fileName, configFile string) {
	configOptionsMu.Lock()
	defer configOptionsMu.Unlock()

	if configuration != nil {
		return
	}

	if projectName = strings.TrimSpace(projectName); projectName != "" {
		configOptions.projectName = projectName
	}
	if fileName = strings.TrimSpace(fileName); fileName != "" {
		configOptions.fileName = fileName
	}
	configOptions.configFile = strings.TrimSpace(configFile)
}

func InitServerConfig(projectName string) error {
	ConfigureServerConfig(projectName, "", "")
	ensureServerConfig()
	return configErr
}

func MustInitServerConfig(projectName, configFile string) {
	ConfigureServerConfig(projectName, "server.yaml", configFile)
	ensureServerConfig()
	if configErr != nil {
		panic(configErr)
	}
}

func (cfg *serverConfig) normalize() {
	if cfg.ClusterId == 0 {
		cfg.ClusterId = 1
	}

	if cfg.Server.AppID == "" {
		cfg.Server.AppID = common.COMMON_PROJECT_NAME
	}
	if cfg.Server.AppName == "" {
		cfg.Server.AppName = cfg.Server.AppID
	}
	if cfg.Server.AppVersion == "" {
		cfg.Server.AppVersion = "v1.0.0"
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Server.IPAddress == "" {
		cfg.Server.IPAddress = "127.0.0.1"
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "9999"
	}
	if cfg.Server.Network == "" {
		cfg.Server.Network = "tcp"
	}

	cfg.DataBase.normalize()

	if cfg.Middleware.Swagger.Title == "" {
		cfg.Middleware.Swagger.Title = cfg.Server.AppName
	}
	if cfg.Prometheus.Namespace == "" {
		cfg.Prometheus.Namespace = normalizeMetricNamespace(cfg.Server.AppID)
	}
	if cfg.Prometheus.Path == "" {
		cfg.Prometheus.Path = "/metrics"
	}
	cfg.Prometheus.ServiceMetrics = normalizeMetricPrefixes(cfg.Prometheus.ServiceMetrics)
}

func (cfg *serverConfig) validate() error {
	var errs []error

	switch cfg.Server.Mode {
	case "debug", "release", "prod":
	default:
		errs = append(errs, fmt.Errorf("server.mode must be one of debug, release, prod"))
	}

	if port, err := strconv.Atoi(cfg.Server.Port); err != nil || port <= 0 || port > 65535 {
		errs = append(errs, fmt.Errorf("server.port must be a valid port"))
	}
	if cfg.Server.MemoryLimit < 0 {
		errs = append(errs, fmt.Errorf("server.memory_limit must be >= 0"))
	}
	if cfg.Server.GCPercent < 0 {
		errs = append(errs, fmt.Errorf("server.gc_percent must be >= 0"))
	}

	if cfg.Redis.Enabled && strings.TrimSpace(cfg.Redis.RedisAddr) == "" {
		errs = append(errs, fmt.Errorf("redis.redis_addr is required when redis.enabled is true"))
	}

	if cfg.Middleware.Limiter.Enabled {
		if cfg.Middleware.Limiter.MaxRequests <= 0 {
			errs = append(errs, fmt.Errorf("middleware.limiter.max_requests must be > 0 when limiter is enabled"))
		}
		if cfg.Middleware.Limiter.Expiration <= 0 {
			errs = append(errs, fmt.Errorf("middleware.limiter.expiration must be > 0 when limiter is enabled"))
		}
	}

	if cfg.Prometheus.Enabled {
		if strings.TrimSpace(cfg.Prometheus.Path) == "" {
			errs = append(errs, fmt.Errorf("prometheus.path is required when prometheus.enabled is true"))
		}
	}

	if cfg.Waf.Enabled && len(cfg.Waf.ConfPath) == 0 {
		errs = append(errs, fmt.Errorf("waf.conf_path is required when waf.enabled is true"))
	}

	switch cfg.DataBase.DBType {
	case "postgres", "postgresql", "mysql", "sqlite":
	case "":
		errs = append(errs, fmt.Errorf("database.db_type is required when database.enabled is true"))
	default:
		errs = append(errs, fmt.Errorf("database.db_type %q is not supported", cfg.DataBase.DBType))
	}

	if cfg.DataBase.Enabled && cfg.DataBase.DBType == "sqlite" {
		if strings.TrimSpace(cfg.DataBase.SQLite.DSN) == "" && strings.TrimSpace(cfg.DataBase.SQLite.Path) == "" {
			errs = append(errs, fmt.Errorf("database.sqlite.path or database.sqlite.dsn is required when sqlite is enabled"))
		}
	}

	return errors.Join(errs...)
}

func (cfg *DataBaseConfig) normalize() {
	cfg.DBType = strings.ToLower(strings.TrimSpace(cfg.DBType))
	if cfg.DBType == "" {
		cfg.DBType = "postgres"
	}

	cfg.applyLegacyConfig()

	if cfg.SQLite.Path == "" {
		cfg.SQLite.Path = "./data/app.db"
	}

	normalizeSQLDefaults(&cfg.Postgres, SQLDataBaseConfig{
		DBHost: "127.0.0.1",
		DBPort: "5432",
		DBName: "gf",
		DBUser: "postgres",
		DBPass: "123456",
	})

	normalizeSQLDefaults(&cfg.MySQL, SQLDataBaseConfig{
		DBHost: "127.0.0.1",
		DBPort: "3306",
		DBName: "gf",
		DBUser: "root",
		DBPass: "123456",
	})
}

func (cfg *DataBaseConfig) applyLegacyConfig() {
	switch cfg.DBType {
	case "sqlite":
		if cfg.SQLite.DSN == "" {
			cfg.SQLite.DSN = strings.TrimSpace(cfg.DSN)
		}
		if cfg.SQLite.Path == "" {
			cfg.SQLite.Path = strings.TrimSpace(cfg.SQLPath)
		}
		if cfg.SQLite.Path == "" {
			cfg.SQLite.Path = strings.TrimSpace(cfg.DBName)
		}
	case "mysql":
		applyLegacySQLConfig(&cfg.MySQL, cfg)
	default:
		applyLegacySQLConfig(&cfg.Postgres, cfg)
	}
}

func applyLegacySQLConfig(target *SQLDataBaseConfig, legacy *DataBaseConfig) {
	if target.DSN == "" {
		target.DSN = strings.TrimSpace(legacy.DSN)
	}
	if target.DBName == "" {
		target.DBName = strings.TrimSpace(legacy.DBName)
	}
	if target.DBHost == "" {
		target.DBHost = strings.TrimSpace(legacy.DBHost)
	}
	if target.DBPort == "" {
		target.DBPort = strings.TrimSpace(legacy.DBPort)
	}
	if target.DBUser == "" {
		target.DBUser = strings.TrimSpace(legacy.DBUser)
	}
	if target.DBPass == "" {
		target.DBPass = strings.TrimSpace(legacy.DBPass)
	}
}

func normalizeSQLDefaults(target *SQLDataBaseConfig, defaults SQLDataBaseConfig) {
	if target.DBHost == "" {
		target.DBHost = defaults.DBHost
	}
	if target.DBPort == "" {
		target.DBPort = defaults.DBPort
	}
	if target.DBName == "" {
		target.DBName = defaults.DBName
	}
	if target.DBUser == "" {
		target.DBUser = defaults.DBUser
	}
	if target.DBPass == "" {
		target.DBPass = defaults.DBPass
	}
}

func normalizeMetricPrefixes(prefixes []string) []string {
	if len(prefixes) == 0 {
		return nil
	}

	result := make([]string, 0, len(prefixes))
	seen := make(map[string]struct{}, len(prefixes))
	for _, item := range prefixes {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func normalizeMetricNamespace(name string) string {
	name = strings.TrimSpace(strings.ReplaceAll(name, "-", "_"))
	if name == "" {
		return "awesome_fiber_template"
	}
	return name
}

func InitConfig(projectName, fileName, configFile string, conf interface{}) error {
	v := viper.New()
	configFile = strings.TrimSpace(configFile)

	if configFile != "" {
		v.SetConfigFile(configFile)
		if ext := strings.TrimPrefix(filepath.Ext(configFile), "."); ext != "" {
			v.SetConfigType(ext)
		}
	} else {
		configName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		configType := strings.TrimPrefix(filepath.Ext(fileName), ".")
		if configName == "" {
			configName = fileName
		}
		if configType == "" {
			configType = "yaml"
		}

		v.SetConfigName(configName)
		v.SetConfigType(configType)
		v.AddConfigPath(filepath.Join("/etc", projectName))

		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error loading pwd dir:", err.Error())
		} else {
			v.AddConfigPath(filepath.Join(pwd, "config"))
		}
	}

	applyDefaults(v, projectName)
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	bindKnownEnvKeys(v)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("can not find any %s file: %w", fileName, err)
	}

	fmt.Println("load config:" + v.ConfigFileUsed())

	settings := collectSettings(v)
	raw, err := yaml.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal merged config failed: %w", err)
	}
	if err := yaml.Unmarshal(raw, conf); err != nil {
		return fmt.Errorf("unmarshal merged config failed: %w", err)
	}

	return nil
}

func ensureServerConfig() {
	configOnce.Do(func() {
		opts := currentConfigOptions()
		cfg := new(serverConfig)
		if err := InitConfig(opts.projectName, opts.fileName, opts.configFile, cfg); err != nil {
			configErr = err
			return
		}
		cfg.normalize()
		if err := cfg.validate(); err != nil {
			configErr = err
			return
		}
		configuration = cfg
	})
}

func applyDefaults(v *viper.Viper, projectName string) {
	v.SetDefault("cluster_id", 1)
	v.SetDefault("server.app_id", common.COMMON_PROJECT_NAME)
	v.SetDefault("server.app_name", common.COMMON_PROJECT_NAME)
	v.SetDefault("server.app_version", "v1.0.0")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.ip_address", "127.0.0.1")
	v.SetDefault("server.port", "1437")
	v.SetDefault("server.memory_limit", 1)
	v.SetDefault("server.gc_percent", 1000)
	v.SetDefault("server.network", "tcp")
	v.SetDefault("server.enable_prefork", false)
	v.SetDefault("server.is_full_stack", false)
	v.SetDefault("database.db_type", "sqlite")
	v.SetDefault("database.auto_migrate", true)
	v.SetDefault("database.sqlite.path", "./data/app.db")
	v.SetDefault("database.postgres.db_host", "127.0.0.1")
	v.SetDefault("database.postgres.db_port", "5432")
	v.SetDefault("database.postgres.db_name", "postgres")
	v.SetDefault("database.postgres.db_username", "postgres")
	v.SetDefault("database.postgres.db_password", "123456")
	v.SetDefault("database.mysql.db_host", "127.0.0.1")
	v.SetDefault("database.mysql.db_port", "3306")
	v.SetDefault("database.mysql.db_name", "mysql")
	v.SetDefault("database.mysql.db_username", "root")
	v.SetDefault("database.mysql.db_password", "123456")
	v.SetDefault("prometheus.namespace", normalizeMetricNamespace(projectName))
	v.SetDefault("prometheus.path", "/metrics")
}

func bindKnownEnvKeys(v *viper.Viper) {
	for _, key := range knownConfigKeys {
		_ = v.BindEnv(key.name)
	}
}

func collectSettings(v *viper.Viper) map[string]interface{} {
	settings := make(map[string]interface{})
	for _, key := range knownConfigKeys {
		setNestedValue(settings, key.name, collectValue(v, key))
	}
	return settings
}

func collectValue(v *viper.Viper, key configKey) interface{} {
	switch key.kind {
	case "bool":
		return v.GetBool(key.name)
	case "int":
		return v.GetInt(key.name)
	case "string_slice":
		if raw, ok := os.LookupEnv(envVariableName(key.name)); ok {
			raw = strings.TrimSpace(raw)
			items := strings.Split(raw, ",")
			result := make([]string, 0, len(items))
			for _, item := range items {
				item = strings.TrimSpace(item)
				if item == "" {
					continue
				}
				result = append(result, item)
			}
			return result
		}
		return v.GetStringSlice(key.name)
	default:
		return v.GetString(key.name)
	}
}

func envVariableName(key string) string {
	replacer := strings.NewReplacer(".", "_", "-", "_")
	return "APP_" + strings.ToUpper(replacer.Replace(key))
}

func setNestedValue(target map[string]interface{}, key string, value interface{}) {
	parts := strings.Split(key, ".")
	current := target
	for index, part := range parts {
		if index == len(parts)-1 {
			current[part] = value
			return
		}

		next, ok := current[part].(map[string]interface{})
		if !ok {
			next = make(map[string]interface{})
			current[part] = next
		}
		current = next
	}
}

func GetServerConfig() *serverConfig {
	ensureServerConfig()
	if configErr != nil {
		panic(configErr)
	}
	return configuration
}

func currentConfigOptions() configLoaderOptions {
	configOptionsMu.Lock()
	defer configOptionsMu.Unlock()
	return configOptions
}
