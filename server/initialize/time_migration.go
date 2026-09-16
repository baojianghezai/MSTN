package initialize

import (
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/gorm"
)

// migrateInt64ToDatetime converts int64 unix-timestamp columns to proper datetime
// BEFORE GORM AutoMigrate attempts ALTER COLUMN.
func migrateInt64ToDatetime(db *gorm.DB) error {
	// Verified from INFORMATION_SCHEMA: all bigint/int time columns still need conversion
	conversions := map[string][]string{
		"ms_members_info":               {"birthday"},
		"ms_members_bind":               {"bindingtime"},
		"ms_members_msgtip":             {"update_time"},
		"ms_members_log":                {"log_addtime"},
		"ms_members_appeal":             {"addtime"},
		"ms_resume":                     {"addtime", "refreshtime", "word_resume_addtime"},
		"ms_jobs_tmp":                   {"addtime", "refreshtime"},
		"ms_pms":                        {"addtime"},
		"ms_personal_jobs_apply":        {"apply_addtime", "reply_time"},
		"ms_company_interview":          {"interview_time", "interview_addtime"},
		"ms_company_profile":            {"addtime", "refreshtime"},
		"ms_company_cancellation_apply": {"finishtime"},
		"ms_company_favorites":          {"addtime"},
		"ms_resume_download":            {"downloaded_at"},
		"ms_order":                      {"created_at", "paid_at", "closed_at", "payment_started_at"},
		"ms_wxpay_log":                  {"addtime"},
		"ms_config":                     {"add_time"},
		"ms_page":                       {"add_time"},
		"ms_navigation":                 {"add_time"},
		"ms_video_interview":            {"interview_time", "addtime"},
		"ms_job_promotion":              {"created_at"},
		"ms_data_cleanup_log":           {"cutoff_at", "created_at"},
		"ms_payment_notify_log":         {"created_at"},
		"ms_oauth":                      {"create_time"},
		"ms_unbind_mobile":              {"add_time"},
	}

	for table, columns := range conversions {
		for _, col := range columns {
			if err := convertColumnInPlace(db, table, col); err != nil {
				global.GVA_LOG.Warn(fmt.Sprintf("skip %s.%s: %v", table, col, err))
			}
		}
	}

	return nil
}

func convertColumnInPlace(db *gorm.DB, table, column string) error {
	// Check if column exists and is numeric
	var colType string
	row := db.Raw(
		"SELECT DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		table, column,
	).Row()
	if err := row.Scan(&colType); err != nil {
		return nil
	}

	// Already datetime — skip
	if strings.Contains(colType, "datetime") || strings.Contains(colType, "timestamp") {
		return nil
	}

	// Check if there are any non-zero, non-null values
	var count int64
	db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE `%s` IS NOT NULL AND `%s` != 0", table, column, column)).Scan(&count)
	if count == 0 {
		return nil
	}

	// 3-step column replacement via temp column
	tmpCol := column + "_bak"
	db.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", table, tmpCol))

	if err := db.Exec(fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` DATETIME(3) NULL", table, tmpCol)).Error; err != nil {
		return fmt.Errorf("add temp column: %w", err)
	}

	if err := db.Exec(fmt.Sprintf(
		"UPDATE `%s` SET `%s` = FROM_UNIXTIME(`%s`) WHERE `%s` IS NOT NULL AND `%s` != 0",
		table, tmpCol, column, column, column,
	)).Error; err != nil {
		db.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", table, tmpCol))
		return fmt.Errorf("populate temp column: %w", err)
	}

	if err := db.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", table, column)).Error; err != nil {
		db.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", table, tmpCol))
		return fmt.Errorf("drop old column: %w", err)
	}

	if err := db.Exec(fmt.Sprintf(
		"ALTER TABLE `%s` CHANGE COLUMN `%s` `%s` DATETIME(3) NULL",
		table, tmpCol, column,
	)).Error; err != nil {
		return fmt.Errorf("rename temp column: %w", err)
	}

	global.GVA_LOG.Info(fmt.Sprintf("Migrated %s.%s: %s → datetime(3)", table, column, colType))
	return nil
}
