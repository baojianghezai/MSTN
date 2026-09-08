package initialize

import (
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// ValidateProductionSecurity prevents a deployment from silently starting with
// example credentials or request-body logging enabled.
func ValidateProductionSecurity() error {
	if !strings.EqualFold(global.GVA_CONFIG.App.Env, "prod") {
		return nil
	}

	signingKey := strings.TrimSpace(global.GVA_CONFIG.JWT.SigningKey)
	if len(signingKey) < 32 || strings.Contains(strings.ToUpper(signingKey), "CHANGE_ME") {
		return fmt.Errorf("production jwt.signing-key must be a unique secret of at least 32 characters")
	}
	if global.GVA_CONFIG.Mysql.Password == "" || global.GVA_CONFIG.Mysql.Password == "123456" || strings.Contains(strings.ToUpper(global.GVA_CONFIG.Mysql.Password), "CHANGE_ME") {
		return fmt.Errorf("production mysql.password must be set through the private production configuration")
	}
	if !global.GVA_CONFIG.System.UseRedis || strings.TrimSpace(global.GVA_CONFIG.Redis.Password) == "" || strings.Contains(strings.ToUpper(global.GVA_CONFIG.Redis.Password), "CHANGE_ME") {
		return fmt.Errorf("production requires Redis with a non-example password for session and rate-limit security")
	}
	if global.GVA_CONFIG.Zap.AccessReqBody || global.GVA_CONFIG.Zap.AccessRespData || global.GVA_CONFIG.Zap.AccessReqHeaders {
		return fmt.Errorf("production access logging must not record request bodies, response bodies, or request headers")
	}
	return nil
}
