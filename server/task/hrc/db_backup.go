package hrc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/task"
)

// RegisterDBBackup 注册数据库备份任务
// 逻辑：mysqldump 备份，保留最近 7 天
func RegisterDBBackup() {
	task.Register("HRC_DBBackup", "每日数据库备份（mysqldump，保留最近7天）", dbBackup)
}

func dbBackup(ctx context.Context, _ json.RawMessage) error {
	cfg := global.GVA_CONFIG.Mysql
	if cfg.Path == "" {
		return fmt.Errorf("mysql config not found")
	}

	backupDir := filepath.Join(".", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("msrc_%s.sql", time.Now().Format("20060102_150405"))
	backupPath := filepath.Join(backupDir, filename)

	cmd := exec.CommandContext(ctx, "mysqldump",
		"-h", cfg.Path,
		"-P", cfg.Port,
		"-u", cfg.Username,
		"-p"+cfg.Password,
		"--databases", cfg.Dbname,
		"--result-file", backupPath,
	)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mysqldump failed: %v, output: %s", err, string(output))
	}

	// 清理 7 天前的备份
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}
	cutoff := time.Now().AddDate(0, 0, -7)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(backupDir, entry.Name()))
		}
	}

	return nil
}
