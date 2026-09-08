package initialize

import (
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

func TestValidateProductionSecurity(t *testing.T) {
	original := global.GVA_CONFIG
	t.Cleanup(func() { global.GVA_CONFIG = original })

	global.GVA_CONFIG = config.Server{
		App:    config.App{Env: "prod"},
		JWT:    config.JWT{SigningKey: strings.Repeat("a", 32)},
		Mysql:  config.Mysql{GeneralDB: config.GeneralDB{Password: "database-secret"}},
		Redis:  config.Redis{Password: "redis-secret"},
		System: config.System{UseRedis: true},
	}
	if err := ValidateProductionSecurity(); err != nil {
		t.Fatalf("valid production configuration rejected: %v", err)
	}

	global.GVA_CONFIG.Zap.AccessReqBody = true
	if err := ValidateProductionSecurity(); err == nil {
		t.Fatal("expected sensitive access logging to be rejected")
	}
}
