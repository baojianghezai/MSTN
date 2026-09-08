package core

import (
	"context"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

func RunServer() {
	if global.GVA_CONFIG.System.UseRedis {
		initialize.Redis()
		if global.GVA_CONFIG.System.UseMultipoint {
			initialize.RedisList()
		}
	}

	// 初始化通用缓存（必须在 Redis 之后：有 Redis 用 Redis，否则用内存）
	initialize.InitGvaCache()

	if global.GVA_CONFIG.System.UseMongo {
		if err := initialize.Mongo.Initialization(); err != nil {
			zap.L().Error(fmt.Sprintf("%+v", err))
		}
	}

	if global.GVA_DB != nil {
		system.LoadAll(context.Background())
		(&system.SecurityConfigService{}).LoadAll(context.Background())
		// 幂等注册 hrc 业务菜单（不依赖「菜单管理」界面手工配置）
		initialize.EnsureHrcMenus()
	}

	Router := initialize.Routers()
	address := fmt.Sprintf(":%d", global.GVA_CONFIG.System.Addr)

	fmt.Printf(`
	欢迎使用 名硕人才网-后台管理
	当前版本:%s
	默认自动化文档地址:http://127.0.0.1%s/swagger/index.html
	默认前端文件运行地址:http://127.0.0.1:8080
`, global.Version, address)

	initServer(address, Router, 10*time.Minute, 10*time.Minute)
}
