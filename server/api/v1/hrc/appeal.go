package hrc

import (
	middlewarehrc "github.com/flipped-aurora/gin-vue-admin/server/middleware/hrc"
	hrcService "github.com/flipped-aurora/gin-vue-admin/server/service/hrc"
	"github.com/gin-gonic/gin"
)

type AppealApi struct{}

// Submit 提交账号申诉（公开，已登录自动带 uid，未登录按手机号匹配）
// @Tags HrcAppeal
// @Summary 提交账号申诉
// @Accept application/json
// @Produce json
// @Param data body AppealSubmitRequest true "申诉参数"
// @Success 200 {object} Response
// @Router /api/v1/appeal [post]
func (a *AppealApi) Submit(c *gin.Context) {
	var req AppealSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, CodeParamError, "参数错误")
		return
	}
	uid := optionalMemberUID(c)
	id, err := hrcService.ServiceGroupApp.AppealService.SubmitAppeal(uid, req.RealName, req.Mobile, req.Email, req.Description)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	OKWithData(c, gin.H{"id": id})
}

// Status 申诉进度查询（按手机号）
// @Tags HrcAppeal
// @Summary 申诉进度查询
// @Produce json
// @Param mobile query string true "手机号"
// @Success 200 {object} Response{data=[]AppealStatusItem}
// @Router /api/v1/appeal/status [get]
func (a *AppealApi) Status(c *gin.Context) {
	mobile := c.Query("mobile")
	list, err := hrcService.ServiceGroupApp.AppealService.GetAppealStatus(mobile)
	if err != nil {
		Fail(c, CodeParamError, err.Error())
		return
	}
	items := make([]AppealStatusItem, 0, len(list))
	for _, v := range list {
		items = append(items, AppealStatusItem{
			ID:          v.ID,
			UID:         v.UID,
			RealName:    v.RealName,
			Mobile:      v.Mobile,
			Email:       v.Email,
			Description: v.Description,
			AddTime:     v.AddTime,
			Status:      v.Status,
			StatusCN:    appealStatusCN(v.Status),
		})
	}
	OKWithData(c, items)
}

// optionalMemberUID 若请求携带有效会员 token 则返回其 uid，否则返回 0（未登录）
func optionalMemberUID(c *gin.Context) uint64 {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return 0
	}
	token := auth
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	if claims, err := middlewarehrc.DefaultMemberJWT().ParseToken(token); err == nil {
		return claims.UID
	}
	return 0
}

func appealStatusCN(status int8) string {
	switch status {
	case 1:
		return "已处理"
	case 2:
		return "已驳回"
	default:
		return "待处理"
	}
}
