package stack

import (
	"fmt"
	"strings"
)

const (
	OptionFiberVersion = "fiber_version"
	OptionCLIStyle     = "cli_style"
	OptionLogger       = "logger"
	OptionDB           = "db"
	OptionDataAccess   = "data_access"

	FiberV2 = "v2"
	FiberV3 = "v3"

	CLINative = "native"
	CLICobra  = "cobra"

	LoggerZap  = "zap"
	LoggerSlog = "slog"

	DBSQLite = "sqlite"
	DBPgSQL  = "pgsql"
	DBMySQL  = "mysql"

	DBPostgresKind = "postgres"

	DataAccessStdlib = "stdlib"
	DataAccessSQLX   = "sqlx"
	DataAccessSQLC   = "sqlc"
)

func DefaultFiberVersion() string {
	return FiberV3
}

func DefaultCLIStyle() string {
	return CLICobra
}

func DefaultLogger() string {
	return LoggerZap
}

func DefaultDB() string {
	return DBSQLite
}

func DefaultDataAccess() string {
	return DataAccessStdlib
}

func NormalizeOptions(options map[string]string) map[string]string {
	normalized := make(map[string]string, len(options)+12)
	for key, value := range options {
		normalized[key] = value
	}

	if strings.EqualFold(strings.TrimSpace(options["_normalized"]), "true") {
		normalized["_normalized"] = "true"
	} else {
		if _, ok := options[OptionLogger]; ok {
			normalized["_explicit_"+OptionLogger] = "true"
		}
		if _, ok := options[OptionDB]; ok {
			normalized["_explicit_"+OptionDB] = "true"
		}
		if _, ok := options[OptionDataAccess]; ok {
			normalized["_explicit_"+OptionDataAccess] = "true"
		}
		normalized["_normalized"] = "true"
	}

	normalized[OptionFiberVersion] = FiberVersion(normalized)
	normalized[OptionCLIStyle] = CLIStyle(normalized)
	normalized[OptionLogger] = Logger(normalized)
	normalized[OptionDB] = DB(normalized)
	normalized[OptionDataAccess] = DataAccess(normalized)
	normalized["fiber_module"] = FiberModule(normalized)
	normalized["fiber_dependency"] = FiberDependency(normalized)
	normalized["default_stack"] = DefaultStackLabel()
	normalized["default_logger"] = DefaultLogger()
	normalized["default_database"] = DefaultDB()
	normalized["default_data_access"] = DefaultDataAccess()
	normalized["logger_backend"] = Logger(normalized)
	normalized["db_kind"] = DBKind(normalized)
	normalized["db_type_default"] = DBKind(normalized)
	normalized["data_access_kind"] = DataAccess(normalized)
	return normalized
}

func FiberVersion(options map[string]string) string {
	version := strings.ToLower(strings.TrimSpace(options[OptionFiberVersion]))
	switch version {
	case "", FiberV3:
		return FiberV3
	case FiberV2:
		return FiberV2
	default:
		return version
	}
}

func CLIStyle(options map[string]string) string {
	style := strings.ToLower(strings.TrimSpace(options[OptionCLIStyle]))
	switch style {
	case "", CLICobra:
		return CLICobra
	case CLINative:
		return CLINative
	default:
		return style
	}
}

func Logger(options map[string]string) string {
	value := strings.ToLower(strings.TrimSpace(options[OptionLogger]))
	switch value {
	case "", LoggerZap:
		return LoggerZap
	case LoggerSlog:
		return LoggerSlog
	default:
		return value
	}
}

func DB(options map[string]string) string {
	value := strings.ToLower(strings.TrimSpace(options[OptionDB]))
	switch value {
	case "", DBSQLite:
		return DBSQLite
	case DBPgSQL:
		return DBPgSQL
	case DBMySQL:
		return DBMySQL
	default:
		return value
	}
}

func DBKind(options map[string]string) string {
	switch DB(options) {
	case DBPgSQL:
		return DBPostgresKind
	case DBMySQL:
		return DBMySQL
	default:
		return DBSQLite
	}
}

func DataAccess(options map[string]string) string {
	value := strings.ToLower(strings.TrimSpace(options[OptionDataAccess]))
	switch value {
	case "", DataAccessStdlib:
		return DataAccessStdlib
	case DataAccessSQLX:
		return DataAccessSQLX
	case DataAccessSQLC:
		return DataAccessSQLC
	default:
		return value
	}
}

func ValidateOptions(options map[string]string) error {
	switch FiberVersion(options) {
	case FiberV2, FiberV3:
	default:
		return fmt.Errorf("fiber version %q is not supported", options[OptionFiberVersion])
	}

	switch CLIStyle(options) {
	case CLINative, CLICobra:
	default:
		return fmt.Errorf("cli style %q is not supported", options[OptionCLIStyle])
	}

	switch Logger(options) {
	case LoggerZap, LoggerSlog:
	default:
		return fmt.Errorf("logger %q is not supported", options[OptionLogger])
	}

	switch DB(options) {
	case DBSQLite, DBPgSQL, DBMySQL:
	default:
		return fmt.Errorf("database %q is not supported", options[OptionDB])
	}

	switch DataAccess(options) {
	case DataAccessStdlib, DataAccessSQLX, DataAccessSQLC:
	default:
		return fmt.Errorf("data access %q is not supported", options[OptionDataAccess])
	}

	return nil
}

func BaseName(base string, options map[string]string) string {
	if CLIStyle(options) == CLICobra {
		return base + "-cobra"
	}
	return base
}

func PackName(pack string, options map[string]string) string {
	if FiberVersion(options) == FiberV3 {
		return pack + "-v3"
	}
	return pack
}

func FiberModule(options map[string]string) string {
	if FiberVersion(options) == FiberV2 {
		return "github.com/gofiber/fiber/v2"
	}
	return "github.com/gofiber/fiber/v3"
}

func FiberDependency(options map[string]string) string {
	if FiberVersion(options) == FiberV2 {
		return "github.com/gofiber/fiber/v2 v2.52.13"
	}
	return "github.com/gofiber/fiber/v3 v3.1.0"
}

func DefaultStackLabel() string {
	return "fiber-v3 + cobra + viper"
}

func SupportedFiberVersions() string {
	return FiberV3 + "," + FiberV2
}

func SupportedCLIStyles() string {
	return CLICobra + "," + CLINative
}

func SupportedLoggers() string {
	return LoggerZap + "," + LoggerSlog
}

func SupportedDatabases() string {
	return DBSQLite + "," + DBPgSQL + "," + DBMySQL
}

func SupportedDataAccess() string {
	return DataAccessStdlib + "," + DataAccessSQLX + "," + DataAccessSQLC
}

func RuntimeOverlayPacks(options map[string]string, presetName string) []string {
	if presetName == "extra-light" {
		return []string{}
	}

	return []string{
		"runtime-logger-" + Logger(options),
		"runtime-data-" + DataAccess(options),
	}
}
