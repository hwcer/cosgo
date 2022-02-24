package cosmo

import (
	"context"
	"gorm.io/gorm/logger"
	"time"
)

// Session session config when create session with Session() method
type Session struct {
	DBName string
	//DryRun                   bool
	//PrepareStmt              bool
	NewDB     bool
	SkipHooks bool
	//SkipDefaultTransaction   bool
	//DisableNestedTransaction bool
	//AllowGlobalUpdate        bool
	//FullSaveAssociations     bool
	//QueryFields              bool
	Context context.Context
	Logger  logger.Interface
	NowTime func() time.Time
	//CreateBatchSize          int
}
