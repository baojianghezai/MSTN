# 微信支付与支付宝配置

本项目的企业套餐支持微信支付 Native 扫码支付和支付宝电脑网站支付。支付成功后的权益开通只依赖异步通知验签，浏览器跳转不会直接开通套餐。

## 上线前准备

1. 为 API 服务配置公网 HTTPS 域名，并将其填入 `payment.notify-base-url`，例如 `https://api.example.com`。
2. 在微信支付商户平台配置 APIv3 密钥、商户私钥和平台证书。回调地址固定为 `https://api.example.com/api/v1/pay/notify/wechat`。
3. 在支付宝开放平台配置应用私钥、支付宝公钥和异步通知地址 `https://api.example.com/api/v1/pay/notify/alipay`。
4. 在配置文件或 Secret 挂载的覆盖文件中填写 `payment` 节；不要把真实私钥、APIv3 密钥提交到仓库。

```yaml
payment:
  notify-base-url: https://api.example.com
  wechat:
    enabled: true
    app-id: wx_app_id
    mch-id: merchant_id
    certificate-serial-no: merchant_certificate_serial
    private-key-path: /run/secrets/wechat/apiclient_key.pem
    platform-cert-path: /run/secrets/wechat/platform_cert.pem
    api-v3-key: replace_with_32_byte_api_v3_key
  alipay:
    enabled: true
    sandbox: false
    app-id: alipay_app_id
    private-key-path: /run/secrets/alipay/app_private_key.pem
    public-key-path: /run/secrets/alipay/alipay_public_key.pem
    seller-id: optional_seller_id
    return-url: https://www.example.com/company/plan
```

## 行为与验收

- 企业端创建订单后，可选择微信或支付宝发起支付。
- 微信返回二维码链接，由企业端展示二维码；支付宝返回支付页跳转链接。
- 回调会验证签名、订单号、金额、微信应用/商户号或支付宝应用/收款账号，并按网关交易号幂等发放权益。
- 回调日志只保存载荷 SHA-256 摘要和校验结果，不保存回调正文。
