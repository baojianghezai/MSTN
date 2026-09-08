package hrc

import (
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type CategoryApi struct{}

func (a *CategoryApi) GetJobsTree(c *gin.Context) {
	data, err := hrcService.ServiceGroupApp.CategoryService.GetJobsTree(c.Request.Context())
	if err != nil {
		Fail(c, CodeParamError, err.Error())
	}
	OKWithData(c, data)
}
