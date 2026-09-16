package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	"github.com/gin-gonic/gin"
)

type ResumeRouter struct{}

// InitResumeRouter 个人简历路由（会员 JWT + utype=1；09 §4.2 #48-#52/#54/#55/#57/#58）
// #53 刷新 / #56 复制 / #59 照片 / #60 发邮箱后置（04 §2.8 后注 M3 收尾边界）
func (r *ResumeRouter) InitResumeRouter(Router *gin.RouterGroup) {
	resumeGroup := Router.Group("personal").Use(middlewarehrc.MemberAuth()).Use(middlewarehrc.UtypeAuth(1))
	{
		resumeGroup.POST("resumes", hrcResumeApi.CreateResume)
		resumeGroup.GET("resumes", hrcResumeApi.List)
		resumeGroup.GET("resumes/:id", hrcResumeApi.GetResume)
		resumeGroup.PUT("resumes/:id", hrcResumeApi.UpdateResume)
		resumeGroup.DELETE("resumes/:id", hrcResumeApi.DeleteResume)
		resumeGroup.PUT("resumes/:id/display", hrcResumeApi.SetDisplay)
		resumeGroup.PUT("resumes/:id/default", hrcResumeApi.SetDefault)
		resumeGroup.GET("resumes/:id/completeness", hrcResumeApi.GetCompleteness)
		resumeGroup.POST("resumes/:id/outward", hrcResumeApi.UploadOutward)
		resumeGroup.DELETE("resumes/:id/outward", hrcResumeApi.DeleteOutward)
	}
}
