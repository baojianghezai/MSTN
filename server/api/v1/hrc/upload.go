package hrc

import (
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type UploadApi struct{}

// Upload 通用文件上传（图片/附件/富文本插图），会员鉴权
// @Tags HrcUpload
// @Summary 通用文件上传
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Success 200 {object} Response
// @Router /api/v1/upload [post]
func (a *UploadApi) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Fail(c, CodeParamError, "接收文件失败")
		return
	}
	url, err := hrcService.ServiceGroupApp.UploadService.UploadFile(c.Request.Context(), file)
	if err != nil {
		Fail(c, CodeParamError, "上传失败")
		return
	}
	OKWithData(c, gin.H{"url": url})
}
