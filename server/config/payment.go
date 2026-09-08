package config

import "strings"

type Payment struct {
	NotifyBaseURL string        `mapstructure:"notify-base-url" json:"notifyBaseUrl" yaml:"notify-base-url"`
	Wechat        WechatPayment `mapstructure:"wechat" json:"wechat" yaml:"wechat"`
	Alipay        AlipayPayment `mapstructure:"alipay" json:"alipay" yaml:"alipay"`
}

type WechatPayment struct {
	Enabled             bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	AppID               string `mapstructure:"app-id" json:"appId" yaml:"app-id"`
	MchID               string `mapstructure:"mch-id" json:"mchId" yaml:"mch-id"`
	CertificateSerialNo string `mapstructure:"certificate-serial-no" json:"certificateSerialNo" yaml:"certificate-serial-no"`
	PrivateKeyPath      string `mapstructure:"private-key-path" json:"privateKeyPath" yaml:"private-key-path"`
	PlatformCertPath    string `mapstructure:"platform-cert-path" json:"platformCertPath" yaml:"platform-cert-path"`
	APIv3Key            string `mapstructure:"api-v3-key" json:"-" yaml:"api-v3-key"`
}

type AlipayPayment struct {
	Enabled        bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Sandbox        bool   `mapstructure:"sandbox" json:"sandbox" yaml:"sandbox"`
	AppID          string `mapstructure:"app-id" json:"appId" yaml:"app-id"`
	PrivateKeyPath string `mapstructure:"private-key-path" json:"privateKeyPath" yaml:"private-key-path"`
	PublicKeyPath  string `mapstructure:"public-key-path" json:"publicKeyPath" yaml:"public-key-path"`
	SellerID       string `mapstructure:"seller-id" json:"sellerId" yaml:"seller-id"`
	ReturnURL      string `mapstructure:"return-url" json:"returnUrl" yaml:"return-url"`
}

func (p Payment) NotifyURL(provider string) string {
	return strings.TrimRight(p.NotifyBaseURL, "/") + "/api/v1/pay/notify/" + provider
}
