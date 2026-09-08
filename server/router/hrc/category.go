package hrc

import "github.com/gin-gonic/gin"

type Category struct{}

func (hrc *Category) InitCategoryRouter(Router *gin.RouterGroup) {

	Router.GET("category/jobstree", hrcCategoryApi.GetJobsTree)
}
