package initialize

import (
	"os"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	"github.com/flipped-aurora/gin-vue-admin/server/model/media"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/logger"

	"gorm.io/gorm"
)

func Gorm() *gorm.DB {
	switch global.GVA_CONFIG.System.DbType {
	case "mysql":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Mysql.Dbname
		return GormMysql()
	case "pgsql":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Pgsql.Dbname
		return GormPgSql()
	case "oracle":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Oracle.Dbname
		return GormOracle()
	case "mssql":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Mssql.Dbname
		return GormMssql()
	case "sqlite":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Sqlite.Dbname
		return GormSqlite()
	default:
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Mysql.Dbname
		return GormMysql()
	}
}

func RegisterTables() {
	if global.GVA_CONFIG.System.DisableAutoMigrate {
		logger.Bg().Mod("system").Info("auto-migrate is disabled, skipping table registration")
		ensureJobCatalogData()
		ensureLogViewerMetadata()
		return
	}

	db := global.GVA_DB

	// Migrate int64 unix timestamps to datetime BEFORE AutoMigrate
	// otherwise ALTER COLUMN fails with "Incorrect datetime value"
	if err := migrateInt64ToDatetime(db); err != nil {
		logger.Bg().Mod("system").Err(err).Error("int64→datetime migration failed")
		os.Exit(1)
	}

	err := db.AutoMigrate(

		system.SysApi{},
		system.SysIgnoreApi{},
		system.SysUser{},
		system.SysBaseMenu{},
		system.JwtBlacklist{},
		system.SysAuthority{},
		system.SysDepartment{},
		system.SysPosition{},
		system.SysDataAccessLog{},
		system.SysAuthorityDepartment{},
		system.SysDictionary{},
		system.SysOperationRecord{},
		system.SysAutoCodeHistory{},
		system.SysDictionaryDetail{},
		system.SysBaseMenuParameter{},
		system.SysBaseMenuBtn{},
		system.SysAuthorityBtn{},
		system.SysAutoCodePackage{},
		system.SysExportTemplate{},
		system.Condition{},
		system.JoinTemplate{},
		system.SysParams{},
		system.SysSecurityConfig{},
		system.SysVersion{},
		system.SysError{},
		system.SysApiToken{},
		system.SysLoginLog{},
		system.SysTimedTask{},
		system.SysTimedTaskLog{},

		example.ExaCustomer{},
		media.MediaUpload{},
		media.MediaUploadChunk{},
		media.FileUploadAndDownload{},
		media.AttachmentCategory{},
	)
	if err != nil {
		logger.Bg().Mod("system").Err(err).Error("register table failed")
		os.Exit(1)
	}

	err = bizModel()

	if err != nil {
		logger.Bg().Mod("system").Err(err).Error("register biz_table failed")
		os.Exit(1)
	}
	ensureLogViewerMetadata()
	logger.Bg().Mod("system").Info("register table success")
}

func ensureLogViewerMetadata() {
	if err := EnsureLogViewerData(); err != nil {
		logger.Bg().Mod("log-viewer").Err(err).Warn("log viewer metadata seed skipped")
	}
}

// ensureJobCatalogData is safe after schema migration: it only inserts
// missing catalogue rows and never changes existing operations data.
func ensureJobCatalogData() {
	if err := seedCategories(global.GVA_DB); err != nil {
		logger.Bg().Mod("hrc").Err(err).Warn("job title catalogue seed skipped")
		return
	}
	if err := seedJobCategories(global.GVA_DB); err != nil {
		logger.Bg().Mod("hrc").Err(err).Warn("job category catalogue seed skipped")
	}
}
