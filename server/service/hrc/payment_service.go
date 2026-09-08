package hrc

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/config"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	hrcModel "github.com/flipped-aurora/gin-vue-admin/server/model/hrc"
	"github.com/smartwalle/alipay/v3"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"gorm.io/gorm"
)

const (
	PaymentProviderWechatNative = "wechat_native"
	PaymentProviderAlipayPage   = "alipay_page"
	paymentExpiration           = 30 * time.Minute
)

var (
	ErrPaymentNotConfigured = errors.New("支付渠道尚未配置")
	ErrPaymentProvider      = errors.New("不支持的支付渠道")
	ErrPaymentOwnership     = errors.New("订单不属于当前企业")
	ErrPaymentInvalidNotify = errors.New("支付回调校验失败")
)

type PaymentStart struct {
	Provider    string `json:"provider"`
	QRCodeURL   string `json:"qrCodeUrl,omitempty"`
	RedirectURL string `json:"redirectUrl,omitempty"`
	ExpiresAt   int64  `json:"expiresAt"`
}

type PaymentService struct{}

func (s *PaymentService) Start(ctx context.Context, uid, orderID uint64, provider string) (*PaymentStart, error) {
	var order hrcModel.Order
	if err := global.GVA_DB.WithContext(ctx).First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.UID != uid {
		return nil, ErrPaymentOwnership
	}
	if order.IsPaid == 2 {
		return nil, ErrOrderPaid
	}
	if order.IsPaid == 3 {
		return nil, ErrOrderClosed
	}
	if order.Amount <= 0 {
		return nil, ErrSetmealInvalid
	}
	if err := validateNotifyURL(global.GVA_CONFIG.Payment.NotifyBaseURL); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(paymentExpiration)
	var result *PaymentStart
	var err error
	switch provider {
	case PaymentProviderWechatNative:
		result, err = s.startWechat(ctx, &order, expiresAt)
	case PaymentProviderAlipayPage:
		result, err = s.startAlipay(&order, expiresAt)
	default:
		return nil, ErrPaymentProvider
	}
	if err != nil {
		return nil, err
	}
	if err := global.GVA_DB.WithContext(ctx).Model(&hrcModel.Order{}).Where("id = ? AND is_paid = ?", order.ID, 1).Updates(map[string]interface{}{
		"payment": provider, "payment_started_at": time.Now().Unix(),
	}).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PaymentService) startWechat(ctx context.Context, order *hrcModel.Order, expiresAt time.Time) (*PaymentStart, error) {
	cfg := global.GVA_CONFIG.Payment.Wechat
	if !cfg.Enabled || cfg.AppID == "" || cfg.MchID == "" || cfg.CertificateSerialNo == "" || cfg.PrivateKeyPath == "" || cfg.APIv3Key == "" {
		return nil, ErrPaymentNotConfigured
	}
	privateKey, err := utils.LoadPrivateKeyWithPath(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载微信支付商户私钥失败: %w", err)
	}
	client, err := core.NewClient(ctx, option.WithWechatPayAutoAuthCipher(cfg.MchID, cfg.CertificateSerialNo, privateKey, cfg.APIv3Key))
	if err != nil {
		return nil, fmt.Errorf("初始化微信支付客户端失败: %w", err)
	}
	service := native.NativeApiService{Client: client}
	description := "招聘服务-" + order.SetmealName
	notifyURL := global.GVA_CONFIG.Payment.NotifyURL("wechat")
	response, _, err := service.Prepay(ctx, native.PrepayRequest{
		Appid:       &cfg.AppID,
		Mchid:       &cfg.MchID,
		Description: &description,
		OutTradeNo:  &order.OID,
		TimeExpire:  &expiresAt,
		NotifyUrl:   &notifyURL,
		Amount:      &native.Amount{Total: &order.Amount},
	})
	if err != nil {
		return nil, fmt.Errorf("微信支付下单失败: %w", err)
	}
	if response == nil || response.CodeUrl == nil || *response.CodeUrl == "" {
		return nil, errors.New("微信支付未返回二维码链接")
	}
	return &PaymentStart{Provider: PaymentProviderWechatNative, QRCodeURL: *response.CodeUrl, ExpiresAt: expiresAt.Unix()}, nil
}

func (s *PaymentService) startAlipay(order *hrcModel.Order, expiresAt time.Time) (*PaymentStart, error) {
	cfg := global.GVA_CONFIG.Payment.Alipay
	if !cfg.Enabled || cfg.AppID == "" || cfg.PrivateKeyPath == "" || cfg.PublicKeyPath == "" {
		return nil, ErrPaymentNotConfigured
	}
	client, err := newAlipayClient(cfg)
	if err != nil {
		return nil, err
	}
	pageURL, err := client.TradePagePay(alipay.TradePagePay{Trade: alipay.Trade{
		Subject:        "招聘服务-" + order.SetmealName,
		OutTradeNo:     order.OID,
		TotalAmount:    centsToYuan(order.Amount),
		ProductCode:    "FAST_INSTANT_TRADE_PAY",
		NotifyURL:      global.GVA_CONFIG.Payment.NotifyURL("alipay"),
		ReturnURL:      cfg.ReturnURL,
		TimeoutExpress: "30m",
	}})
	if err != nil {
		return nil, fmt.Errorf("支付宝下单失败: %w", err)
	}
	return &PaymentStart{Provider: PaymentProviderAlipayPage, RedirectURL: pageURL.String(), ExpiresAt: expiresAt.Unix()}, nil
}

func (s *PaymentService) HandleWechatNotify(ctx context.Context, req *http.Request) error {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	req.Body = io.NopCloser(bytes.NewReader(body))

	cfg := global.GVA_CONFIG.Payment.Wechat
	if !cfg.Enabled || cfg.AppID == "" || cfg.MchID == "" || cfg.PlatformCertPath == "" || cfg.APIv3Key == "" {
		return ErrPaymentNotConfigured
	}
	certificate, err := utils.LoadCertificateWithPath(cfg.PlatformCertPath)
	if err != nil {
		return fmt.Errorf("加载微信支付平台证书失败: %w", err)
	}
	verifier := verifiers.NewSHA256WithRSAVerifier(core.NewCertificateMapWithList([]*x509.Certificate{certificate}))
	handler, err := notify.NewRSANotifyHandler(cfg.APIv3Key, verifier)
	if err != nil {
		return err
	}
	transaction := new(payments.Transaction)
	if _, err := handler.ParseNotifyRequest(ctx, req, transaction); err != nil {
		s.logNotify(ctx, PaymentProviderWechatNative, 0, "", "", body, false, err.Error())
		return fmt.Errorf("%w: %v", ErrPaymentInvalidNotify, err)
	}
	if transaction.OutTradeNo == nil || transaction.TransactionId == nil || transaction.Amount == nil || transaction.Amount.Total == nil || transaction.TradeState == nil || transaction.Appid == nil || transaction.Mchid == nil ||
		*transaction.TradeState != "SUCCESS" || *transaction.Appid != cfg.AppID || *transaction.Mchid != cfg.MchID {
		s.logNotify(ctx, PaymentProviderWechatNative, 0, valueOf(transaction.OutTradeNo), valueOf(transaction.TransactionId), body, true, "通知字段不合法")
		return ErrPaymentInvalidNotify
	}
	return s.confirmNotify(ctx, PaymentProviderWechatNative, valueOf(transaction.OutTradeNo), *transaction.Amount.Total, valueOf(transaction.TransactionId), body)
}

func (s *PaymentService) HandleAlipayNotify(ctx context.Context, values url.Values) error {
	cfg := global.GVA_CONFIG.Payment.Alipay
	if !cfg.Enabled || cfg.AppID == "" || cfg.PrivateKeyPath == "" || cfg.PublicKeyPath == "" {
		return ErrPaymentNotConfigured
	}
	client, err := newAlipayClient(cfg)
	if err != nil {
		return err
	}
	payload := []byte(values.Encode())
	if err := client.VerifySign(ctx, values); err != nil {
		s.logNotify(ctx, PaymentProviderAlipayPage, 0, values.Get("out_trade_no"), values.Get("trade_no"), payload, false, err.Error())
		return fmt.Errorf("%w: %v", ErrPaymentInvalidNotify, err)
	}
	status := alipay.TradeStatus(values.Get("trade_status"))
	if (status != alipay.TradeStatusSuccess && status != alipay.TradeStatusFinished) || values.Get("app_id") != cfg.AppID || values.Get("out_trade_no") == "" || values.Get("trade_no") == "" ||
		(cfg.SellerID != "" && values.Get("seller_id") != cfg.SellerID) {
		s.logNotify(ctx, PaymentProviderAlipayPage, 0, values.Get("out_trade_no"), values.Get("trade_no"), payload, true, "通知字段不合法")
		return ErrPaymentInvalidNotify
	}
	return s.confirmAlipayNotify(ctx, values.Get("out_trade_no"), values.Get("total_amount"), values.Get("trade_no"), payload)
}

func (s *PaymentService) confirmNotify(ctx context.Context, provider, outTradeNo string, amount int64, transactionID string, payload []byte) error {
	var order hrcModel.Order
	if err := global.GVA_DB.WithContext(ctx).Where("oid = ?", outTradeNo).First(&order).Error; err != nil {
		s.logNotify(ctx, provider, 0, outTradeNo, transactionID, payload, true, err.Error())
		return ErrPaymentInvalidNotify
	}
	if err := ServiceGroupApp.OrderService.ConfirmGatewayPaid(ctx, order.ID, amount, provider, transactionID); err != nil {
		s.logNotify(ctx, provider, order.ID, outTradeNo, transactionID, payload, true, err.Error())
		return err
	}
	s.logNotify(ctx, provider, order.ID, outTradeNo, transactionID, payload, true, "")
	return nil
}

func (s *PaymentService) confirmAlipayNotify(ctx context.Context, outTradeNo, totalAmount, transactionID string, payload []byte) error {
	var order hrcModel.Order
	if err := global.GVA_DB.WithContext(ctx).Where("oid = ?", outTradeNo).First(&order).Error; err != nil {
		s.logNotify(ctx, PaymentProviderAlipayPage, 0, outTradeNo, transactionID, payload, true, err.Error())
		return ErrPaymentInvalidNotify
	}
	if totalAmount != centsToYuan(order.Amount) {
		s.logNotify(ctx, PaymentProviderAlipayPage, order.ID, outTradeNo, transactionID, payload, true, "实收金额与订单金额不一致")
		return ErrPaymentInvalidNotify
	}
	return s.confirmNotify(ctx, PaymentProviderAlipayPage, outTradeNo, order.Amount, transactionID, payload)
}

func (s *PaymentService) logNotify(ctx context.Context, provider string, orderID uint64, outTradeNo, transactionID string, payload []byte, verified bool, message string) {
	digest := sha256.Sum256(payload)
	if len(message) > 255 {
		message = message[:255]
	}
	_ = global.GVA_DB.WithContext(ctx).Create(&hrcModel.PaymentNotifyLog{
		Provider: provider, OrderID: orderID, OutTradeNo: outTradeNo, TransactionID: transactionID,
		PayloadSHA256: fmt.Sprintf("%x", digest), Verified: verified, Message: message, CreatedAt: time.Now().Unix(),
	}).Error
}

func newAlipayClient(cfg config.AlipayPayment) (*alipay.Client, error) {
	privateKey, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载支付宝应用私钥失败: %w", err)
	}
	publicKey, err := os.ReadFile(cfg.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("加载支付宝公钥失败: %w", err)
	}
	client, err := alipay.New(cfg.AppID, string(privateKey), !cfg.Sandbox)
	if err != nil {
		return nil, fmt.Errorf("初始化支付宝客户端失败: %w", err)
	}
	if err := client.LoadAliPayPublicKey(string(publicKey)); err != nil {
		return nil, fmt.Errorf("加载支付宝公钥失败: %w", err)
	}
	return client, nil
}

func validateNotifyURL(baseURL string) error {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return errors.New("支付回调地址必须配置为公网 HTTPS 地址")
	}
	return nil
}

func centsToYuan(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

func valueOf(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
