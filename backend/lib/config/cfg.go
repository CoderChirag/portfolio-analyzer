package config

import (
	"coderchirag/folio-tracker/lib/env"
	"fmt"
	"strconv"
)

type Config struct {
	DbHost          string
	DbPort          string
	DbRootUser      string
	DbRootPassword  string
	DbFolioUser     string
	DbFolioPassword string
	DbName          string
	DbPoolSize      int
	MigrationsDir   string
	Port            int
}

func Load() (*Config, error) {
	host, err := env.Getenv("DB_HOST")
	if err != nil {
		return nil, err
	}
	dbPort := env.GetenvDefault("DB_PORT", "3306")
	rootUser, err := env.Getenv("DB_ROOT_USER")
	if err != nil {
		return nil, err
	}
	rootPass, err := env.Getenv("DB_ROOT_PASSWORD")
	if err != nil {
		return nil, err
	}
	folioUser, err := env.Getenv("DB_FOLIO_USER")
	if err != nil {
		return nil, err
	}
	folioPass, err := env.Getenv("DB_FOLIO_PASSWORD")
	if err != nil {
		return nil, err
	}
	dbName := env.GetenvDefault("DB_NAME", "folio")
	migrationsDir, err := env.Getenv("MIGRATIONS_DIR")
	if err != nil {
		return nil, err
	}
	ps, err := env.Getenv("DB_POOL_SIZE")
	if err != nil {
		return nil, err
	}
	poolSize, err := strconv.Atoi(ps)
	if err != nil {
		return nil, fmt.Errorf("invalid poolsize: %w", err)
	}

	port, err := strconv.Atoi(env.GetenvDefault("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	config := &Config{
		DbHost:          host,
		DbPort:          dbPort,
		DbRootUser:      rootUser,
		DbRootPassword:  rootPass,
		DbFolioUser:     folioUser,
		DbFolioPassword: folioPass,
		DbName:          dbName,
		DbPoolSize:      poolSize,
		MigrationsDir:   migrationsDir,
		Port:            port,
	}
	return config, nil
}
