package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"text/template"
	"time"

	"vm-monitoring-system/internal/models"
)

// NotificationService 通知服务
type NotificationService struct {
	enabled    bool
	smtpConfig *SMTPConfig
	smsConfig  *SMSConfig
	httpClient *http.Client
}

// SMTPConfig SMTP配置
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	TLS      bool
}

// SMSConfig SMS配置
type SMSConfig struct {
	Provider    string // aliyun, tencent, twilio
	AccessKey   string
	SecretKey   string
	AppID       string
	SignName    string
	TemplateCode string
}

// NotificationResult 通知结果
type NotificationResult struct {
	Method    string    `json:"method"`
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// NewNotificationService 创建通知服务
func NewNotificationService() *NotificationService {
	return &NotificationService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetupSMTP 配置SMTP
func (s *NotificationService) SetupSMTP(config *SMTPConfig) {
	s.smtpConfig = config
}

// SetupSMS 配置SMS
func (s *NotificationService) SetupSMS(config *SMSConfig) {
	s.smsConfig = config
}

// Enable 启用通知服务
func (s *NotificationService) Enable() {
	s.enabled = true
}

// Disable 禁用通知服务
func (s *NotificationService) Disable() {
	s.enabled = false
}

// IsEnabled 检查是否启用
func (s *NotificationService) IsEnabled() bool {
	return s.enabled
}

// SendAlert 发送告警通知
func (s *NotificationService) SendAlert(ctx context.Context, alert models.AlertRecord, config models.JSONMap) []NotificationResult {
	if !s.enabled {
		return nil
	}

	results := []NotificationResult{}

	// 解析通知配置
	var notificationConfig models.NotificationConfig
	if configData, err := json.Marshal(config); err == nil {
		if err := json.Unmarshal(configData, &notificationConfig); err != nil {
			log.Printf("解析通知配置失败: %v", err)
			results = append(results, NotificationResult{
				Method:    "config",
				Success:   false,
				Message:   "配置解析失败: " + err.Error(),
				Timestamp: time.Now(),
			})
			return results
		}
	} else {
		log.Printf("序列化配置失败: %v", err)
		results = append(results, NotificationResult{
			Method:    "config",
			Success:   false,
			Message:   "配置序列化失败: " + err.Error(),
			Timestamp: time.Now(),
		})
		return results
	}

	// 如果配置为空，使用默认配置
	if len(notificationConfig.Methods) == 0 {
		notificationConfig.Methods = []string{"inApp"}
	}

	// 根据配置的方法发送通知
	for _, method := range notificationConfig.Methods {
		var result NotificationResult

		switch method {
		case "email":
			result = s.sendEmailNotification(ctx, alert, notificationConfig.Email)
		case "sms":
			result = s.sendSMSNotification(ctx, alert, notificationConfig.SMS)
		case "webhook":
			result = s.sendWebhookNotification(ctx, alert, notificationConfig.Webhook)
		case "inApp":
			result = s.sendInAppNotification(ctx, alert, notificationConfig.InApp)
		default:
			result = NotificationResult{
				Method:    method,
				Success:   false,
				Message:   "未知的通知方法",
				Timestamp: time.Now(),
			}
		}

		results = append(results, result)
	}

	// 更新通知状态
	s.updateNotificationStatus(alert.ID, results)

	return results
}

// sendEmailNotification 发送邮件通知
func (s *NotificationService) sendEmailNotification(ctx context.Context, alert models.AlertRecord, emailConfig *struct {
	Enabled     bool     `json:"enabled"`
	Recipients  []string `json:"recipients"`
	CC          []string `json:"cc,omitempty"`
	Template    string   `json:"template,omitempty"`
}) NotificationResult {
	if s.smtpConfig == nil || emailConfig == nil || !emailConfig.Enabled {
		return NotificationResult{
			Method:    "email",
			Success:   false,
			Message:   "邮件配置未启用或不存在",
			Timestamp: time.Now(),
		}
	}

	if len(emailConfig.Recipients) == 0 {
		return NotificationResult{
			Method:    "email",
			Success:   false,
			Message:   "没有配置收件人",
			Timestamp: time.Now(),
		}
	}

	// 构建邮件内容
	subject := fmt.Sprintf("【%s】VM监控告警: %s", getSeverityLabel(alert.Severity), alert.RuleName)
	body := s.buildEmailBody(alert)

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", s.smtpConfig.Host, s.smtpConfig.Port)
	auth := smtp.PlainAuth("", s.smtpConfig.Username, s.smtpConfig.Password, s.smtpConfig.Host)

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		joinAddresses(emailConfig.Recipients),
		subject,
		body,
	))

	err := smtp.SendMail(addr, auth, s.smtpConfig.From, emailConfig.Recipients, msg)
	if err != nil {
		log.Printf("发送邮件失败: %v", err)
		return NotificationResult{
			Method:    "email",
			Success:   false,
			Message:   err.Error(),
			Timestamp: time.Now(),
		}
	}

	log.Printf("邮件通知已发送: %s -> %v", alert.RuleName, emailConfig.Recipients)
	return NotificationResult{
		Method:    "email",
		Success:   true,
		Timestamp: time.Now(),
	}
}

// sendSMSNotification 发送短信通知
func (s *NotificationService) sendSMSNotification(ctx context.Context, alert models.AlertRecord, smsConfig *struct {
	Enabled      bool     `json:"enabled"`
	PhoneNumbers []string `json:"phoneNumbers"`
	Template     string   `json:"template,omitempty"`
}) NotificationResult {
	if s.smsConfig == nil || smsConfig == nil || !smsConfig.Enabled {
		return NotificationResult{
			Method:    "sms",
			Success:   false,
			Message:   "短信配置未启用或不存在",
			Timestamp: time.Now(),
		}
	}

	if len(smsConfig.PhoneNumbers) == 0 {
		return NotificationResult{
			Method:    "sms",
			Success:   false,
			Message:   "没有配置手机号",
			Timestamp: time.Now(),
		}
	}

	// 构建短信内容（暂时未使用）
	_ = fmt.Sprintf("【VM监控】%s告警: %s, VM: %s, 当前值: %.2f, 阈值: %.2f",
		getSeverityLabel(alert.Severity),
		alert.RuleName,
		getStringValue(alert.VMName),
		alert.TriggerValue,
		alert.Threshold,
	)

	// TODO: 接入具体的短信服务商API
	// 阿里云SMS、腾讯云SMS、Twilio等
	
	log.Printf("短信通知(模拟): %s -> %v", alert.RuleName, smsConfig.PhoneNumbers)
	return NotificationResult{
		Method:    "sms",
		Success:   true,
		Message:   "短信通知已发送(模拟)",
		Timestamp: time.Now(),
	}
}

// sendWebhookNotification 发送Webhook通知
func (s *NotificationService) sendWebhookNotification(ctx context.Context, alert models.AlertRecord, webhookConfig *struct {
	Enabled bool              `json:"enabled"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Secret  string            `json:"secret,omitempty"`
}) NotificationResult {
	if webhookConfig == nil || !webhookConfig.Enabled {
		return NotificationResult{
			Method:    "webhook",
			Success:   false,
			Message:   "Webhook配置未启用或不存在",
			Timestamp: time.Now(),
		}
	}

	if webhookConfig.URL == "" {
		return NotificationResult{
			Method:    "webhook",
			Success:   false,
			Message:   "Webhook URL未配置",
			Timestamp: time.Now(),
		}
	}

	// 构建Webhook数据
	payload := map[string]interface{}{
		"id":           alert.ID,
		"ruleName":     alert.RuleName,
		"severity":     alert.Severity,
		"metric":       alert.Metric,
		"triggerValue": alert.TriggerValue,
		"threshold":    alert.Threshold,
		"vmName":       getStringValue(alert.VMName),
		"triggeredAt":  alert.TriggeredAt,
		"timestamp":    time.Now(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return NotificationResult{
			Method:    "webhook",
			Success:   false,
			Message:   fmt.Sprintf("序列化数据失败: %v", err),
			Timestamp: time.Now(),
		}
	}

	// 发送HTTP请求
	method := webhookConfig.Method
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequestWithContext(ctx, method, webhookConfig.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return NotificationResult{
			Method:    "webhook",
			Success:   false,
			Message:   fmt.Sprintf("创建请求失败: %v", err),
			Timestamp: time.Now(),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range webhookConfig.Headers {
		req.Header.Set(key, value)
	}

	// 添加签名（如果配置了密钥）
	if webhookConfig.Secret != "" {
		signature := s.generateSignature(jsonData, webhookConfig.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
		req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return NotificationResult{
			Method:    "webhook",
			Success:   false,
			Message:   fmt.Sprintf("发送请求失败: %v", err),
			Timestamp: time.Now(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Webhook通知已发送: %s -> %s", alert.RuleName, webhookConfig.URL)
		return NotificationResult{
			Method:    "webhook",
			Success:   true,
			Timestamp: time.Now(),
		}
	}

	return NotificationResult{
		Method:    "webhook",
		Success:   false,
		Message:   fmt.Sprintf("Webhook返回非成功状态码: %d", resp.StatusCode),
		Timestamp: time.Now(),
	}
}

// sendInAppNotification 发送应用内通知
func (s *NotificationService) sendInAppNotification(ctx context.Context, alert models.AlertRecord, inAppConfig *struct {
	Enabled bool     `json:"enabled"`
	Users   []string `json:"users,omitempty"`
}) NotificationResult {
	// 应用内通知直接写入通知表或推送到WebSocket
	// TODO: 实现WebSocket推送或通知表写入
	
	log.Printf("应用内通知: %s - %s", alert.RuleName, getStringValue(alert.VMName))
	return NotificationResult{
		Method:    "inApp",
		Success:   true,
		Timestamp: time.Now(),
	}
}

// buildEmailBody 构建邮件正文
func (s *NotificationService) buildEmailBody(alert models.AlertRecord) string {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: {{.HeaderColor}}; color: white; padding: 20px; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 20px; border: 1px solid #ddd; }
        .footer { background: #eee; padding: 10px; text-align: center; font-size: 12px; color: #666; }
        .metric { background: white; padding: 15px; margin: 10px 0; border-left: 4px solid {{.HeaderColor}}; }
        .label { font-weight: bold; color: #333; }
        .value { color: #666; }
        .critical { background: #f44336; }
        .high { background: #ff9800; }
        .medium { background: #2196f3; }
        .low { background: #4caf50; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header {{.SeverityClass}}">
            <h2>🚨 VM监控告警通知</h2>
            <p>告警级别: {{.SeverityLabel}}</p>
        </div>
        <div class="content">
            <h3>{{.RuleName}}</h3>
            
            <div class="metric">
                <p><span class="label">虚拟机:</span> <span class="value">{{.VMName}}</span></p>
                <p><span class="label">集群:</span> <span class="value">{{.ClusterName}}</span></p>
                <p><span class="label">指标:</span> <span class="value">{{.Metric}}</span></p>
                <p><span class="label">触发值:</span> <span class="value" style="color: {{.ValueColor}}; font-weight: bold;">{{.TriggerValue}}</span></p>
                <p><span class="label">阈值:</span> <span class="value">{{.Threshold}}</span></p>
                <p><span class="label">条件:</span> <span class="value">{{.Condition}}</span></p>
                <p><span class="label">触发时间:</span> <span class="value">{{.TriggeredAt}}</span></p>
            </div>
            
            <p>请尽快登录系统查看详情并处理此告警。</p>
            
            <a href="{{.DetailURL}}" style="display: inline-block; background: {{.HeaderColor}}; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">查看详情</a>
        </div>
        <div class="footer">
            <p>此邮件由 VM监控系统 自动发送</p>
            <p>{{.Timestamp}}</p>
        </div>
    </div>
</body>
</html>`

	data := map[string]string{
		"RuleName":     alert.RuleName,
		"VMName":       getStringValue(alert.VMName),
		"ClusterName":  getStringValue(alert.ClusterID),
		"Metric":       alert.Metric,
		"TriggerValue": fmt.Sprintf("%.2f", alert.TriggerValue),
		"Threshold":    fmt.Sprintf("%.2f", alert.Threshold),
		"Condition":    getStringValue(alert.ConditionStr),
		"TriggeredAt":  alert.TriggeredAt.Format("2006-01-02 15:04:05"),
		"SeverityLabel": getSeverityLabel(alert.Severity),
		"SeverityClass": getSeverityClass(alert.Severity),
		"HeaderColor":  getSeverityColor(alert.Severity),
		"ValueColor":   getSeverityColor(alert.Severity),
		"DetailURL":    fmt.Sprintf("https://your-domain.com/alerts/%s", alert.ID),
		"Timestamp":    time.Now().Format("2006-01-02 15:04:05"),
	}

	t := template.Must(template.New("email").Parse(tmpl))
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Sprintf("<h1>告警: %s</h1><p>虚拟机: %s</p><p>指标 %s 触发告警，当前值: %.2f，阈值: %.2f</p>",
			alert.RuleName, getStringValue(alert.VMName), alert.Metric, alert.TriggerValue, alert.Threshold)
	}

	return buf.String()
}

// generateSignature 生成Webhook签名 (HMAC-SHA256)
func (s *NotificationService) generateSignature(data []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// verifySignature 验证Webhook签名
func (s *NotificationService) verifySignature(data []byte, signature string, secret string) bool {
	expectedSignature := s.generateSignature(data, secret)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// updateNotificationStatus 更新通知状态到数据库
func (s *NotificationService) updateNotificationStatus(alertID interface{}, results []NotificationResult) {
	// TODO: 将通知结果保存到数据库
	// 更新alert_records表的notification_status字段
	log.Printf("通知结果: alertID=%v, results=%+v", alertID, results)
}

// 辅助函数

func getSeverityLabel(severity string) string {
	switch severity {
	case "critical":
		return "严重"
	case "high":
		return "高"
	case "medium":
		return "中"
	case "low":
		return "低"
	default:
		return severity
	}
}

func getSeverityClass(severity string) string {
	switch severity {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	case "low":
		return "low"
	default:
		return "medium"
	}
}

func getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return "#f44336"
	case "high":
		return "#ff9800"
	case "medium":
		return "#2196f3"
	case "low":
		return "#4caf50"
	default:
		return "#2196f3"
	}
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func joinAddresses(addresses []string) string {
	result := ""
	for i, addr := range addresses {
		if i > 0 {
			result += ", "
		}
		result += addr
	}
	return result
}
