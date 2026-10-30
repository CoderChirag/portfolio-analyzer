package init_mod

import "coderchirag/folio-tracker/lib/env"

type envCfg struct {
	dbHost          string
	dbPort          string
	dbRootUser      string
	dbRootPassword  string
	dbFolioUser     string
	dbFolioPassword string
	dbName          string
	migrationsDir   string
}

func newEnvCfg() (*envCfg, error) {
	host, err := env.Getenv("DB_HOST")
	if err != nil {
		return nil, err
	}
	port := env.GetenvDefault("DB_PORT", "3306")
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

	cfg := &envCfg{
		dbHost:          host,
		dbPort:          port,
		dbRootUser:      rootUser,
		dbRootPassword:  rootPass,
		dbFolioUser:     folioUser,
		dbFolioPassword: folioPass,
		dbName:          dbName,
		migrationsDir:   migrationsDir,
	}
	return cfg, nil
}
