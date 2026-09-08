package hrc

import "github.com/gin-gonic/gin"

type SetmealRouter struct{}

func (r *SetmealRouter) InitSetmealRouter(router *gin.RouterGroup) {
	router.GET("setmeals", hrcSetmealApi.PublicList)
}
