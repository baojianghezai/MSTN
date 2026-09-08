package hrc

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type JobsApi struct{}

// List 职位列表（公开，09 §4.1 #22）
// @Tags HrcJobs
// @Summary 职位列表
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param keyword query string false "关键字（职位名/企业名）"
// @Param trade query int false "行业"
// @Param category query int false "二级分类"
// @Param district query string false "地区"
// @Param education query int false "学历"
// @Param experience query int false "经验"
// @Param minwage query int false "最低薪资"
// @Param maxwage query int false "最高薪资"
// @Param order query string false "排序：last/addtime/salary/stick"
// @Success 200 {object} Response{data=PageData}
// @Router /api/v1/jobs [get]
func (a *JobsApi) List(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	f := hrcService.JobsSearchFilter{
		Keyword:  c.Query("keyword"),
		District: c.Query("district"),
		Order:    c.DefaultQuery("order", "last"),
	}
	f.Trade = queryUint16(c, "trade")
	f.Category = queryUint16(c, "category")
	f.Education = queryUint16(c, "education")
	f.Experience = queryUint16(c, "experience")
	f.MinWage, _ = strconv.Atoi(c.Query("minwage"))
	f.MaxWage, _ = strconv.Atoi(c.Query("maxwage"))

	list, total, err := hrcService.ServiceGroupApp.JobsSearchService.Search(c.Request.Context(), f, pageInfo)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, PageData{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize})
}

// Detail 职位详情（公开，09 §4.1 #23，click 自增）
// @Tags HrcJobs
// @Summary 职位详情
// @Produce json
// @Param id path int true "职位 id"
// @Success 200 {object} Response{data=hrcService.JobPublicDetail}
// @Router /api/v1/jobs/{id} [get]
func (a *JobsApi) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	detail, err := hrcService.ServiceGroupApp.JobsSearchService.Detail(c.Request.Context(), id)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, detail)
}

// HotWords 热门搜索词（公开，09 §4.1 #24）
// @Tags HrcJobs
// @Summary 热门搜索词
// @Produce json
// @Success 200 {object} Response{data=[]string}
// @Router /api/v1/jobs/hot-words [get]
func (a *JobsApi) HotWords(c *gin.Context) {
	words, err := hrcService.ServiceGroupApp.JobsSearchService.HotWords(c.Request.Context())
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, words)
}

// Filters 筛选元数据（公开，09 §4.1 #25：分类聚合，复用 #21）
// @Tags HrcJobs
// @Summary 筛选元数据
// @Produce json
// @Success 200 {object} Response{data=object}
// @Router /api/v1/jobs/filters [get]
func (a *JobsApi) Filters(c *gin.Context) {
	categories, err := hrcService.ServiceGroupApp.CmsService.GetCategories(c.Request.Context())
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, categories)
}

// queryUint16 从 query 取 uint16（非法/空返回 0）
func queryUint16(c *gin.Context, key string) uint16 {
	v, _ := strconv.ParseUint(c.Query(key), 10, 16)
	return uint16(v)
}
