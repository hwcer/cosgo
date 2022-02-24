package cosmo

// Plugin GORM plugin interface
type Plugin interface {
	Name() string
	Initialize(*DB) error
}

type ExecuteHandle func(db *DB) error
