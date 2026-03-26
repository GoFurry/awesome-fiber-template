package env

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/GoFurry/awesome-go-template/fiber/v3/common"
	"gopkg.in/yaml.v2"
)

func init() {
	InitServerConfig(common.COMMON_PROJECT_NAME)
}

var configuration = new(serverConfig)

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

type PrometheusConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Url            string   `yaml:"url"`
	AuthName       string   `yaml:"auth_name"`
	AuthPass       string   `yaml:"auth_pass"`
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
	Enabled  bool                 `yaml:"enabled"`
	DBType   string               `yaml:"db_type"`
	SQLite   SQLiteDataBaseConfig `yaml:"sqlite"`
	Postgres SQLDataBaseConfig    `yaml:"postgres"`
	MySQL    SQLDataBaseConfig    `yaml:"mysql"`
	DSN      string               `yaml:"dsn"`
	DBName   string               `yaml:"db_name"`
	DBHost   string               `yaml:"db_host"`
	DBPort   string               `yaml:"db_port"`
	DBUser   string               `yaml:"db_username"`
	DBPass   string               `yaml:"db_password"`
	SQLPath  string               `yaml:"sqlite_path"`
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

func InitServerConfig(projectName string) {
	InitConfig(projectName, "server.yaml", configuration)
	configuration.normalize()
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
	cfg.Prometheus.ServiceMetrics = normalizeMetricPrefixes(cfg.Prometheus.ServiceMetrics)
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

	res := make([]string, 0, len(prefixes))
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
		res = append(res, item)
	}
	return res
}

func InitConfig(projectName string, fileName string, conf interface{}) {
	hit := false

	file := "/etc/" + projectName + "/" + fileName
	if FileExists(file) {
		err := loadYaml(file, conf)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			hit = true
		}
	}

	if !hit {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error loading pwd dir:", err.Error())
		} else {
			filePath := pwd + "/conf/" + fileName
			if FileExists(filePath) {
				err = loadYaml(filePath, conf)
				if err != nil {
					fmt.Println("Error loading "+fileName+" file:", err.Error())
				} else {
					hit = true
				}
			}
		}
	}

	if !hit {
		fmt.Println("can not find any " + fileName + " file")
		panic("can not find any " + fileName + " file")
	}
}

func FileExists(path string) bool {
	fmt.Println("check filepath:" + path)
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func loadYaml(path string, conf interface{}) error {
	fmt.Println("load config:" + path)
	if FileExists(path) {
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(fileBytes, conf)
	}
	return errors.New("can not find config file: " + path)
}

func GetServerConfig() *serverConfig {
	return configuration
}
