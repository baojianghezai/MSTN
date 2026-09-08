package hrc

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/utils/upload"
)

// UploadService 通用文件上传（09 §4.1 编号 20）
type UploadService struct{}

// UploadFile 上传单文件，返回可访问 URL 路径（local/OSS 由配置 system.oss-type 决定）
func (s *UploadService) UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error) {
	maxSize := global.GVA_CONFIG.Media.MaxFileSize
	if maxSize <= 0 {
		maxSize = 10 << 20
	}
	if err := upload.ValidateRecruitmentFile(file, maxSize); err != nil {
		return "", errors.New("uploaded file failed security validation")
	}
	url, _, err := upload.NewOss().UploadFile(ctx, file)
	return url, err
}
