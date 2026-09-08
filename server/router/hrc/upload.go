package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type UploadRouter struct{}

// InitUploadRouter 通用文件上传（会员鉴权，09 §4.1 编号 20）
func (r *UploadRouter) InitUploadRouter(Router *gin.RouterGroup) {
	Router.POST("upload", middlewarehrc.MemberAuth(), hrcUploadApi.Upload)
}
