package automation

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"strings"

	"github.com/expr-lang/expr"
)

// RunHTTPRequest 执行标准 HTTP 请求动作
func RunHTTPRequest(config map[string]interface{}) StepResult {
	url, _ := config["url"].(string)
	method, _ := config["method"].(string)
	if method == "" {
		method = "GET"
	}

	var bodyReader io.Reader
	if bodyStr, ok := config["body"].(string); ok && bodyStr != "" {
		bodyReader = bytes.NewBufferString(bodyStr)
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return StepResult{Status: "failed", Error: err.Error()}
	}

	if headers, ok := config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if s, ok := v.(string); ok {
				req.Header.Set(k, s)
			}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return StepResult{Status: "failed", Error: err.Error()}
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	return StepResult{
		Status: "success",
		Output: map[string]interface{}{"status_code": resp.StatusCode, "body": string(respBody)},
	}
}

// RunSendEmail 执行 SMTP 发送邮件动作
func RunSendEmail(config map[string]interface{}) StepResult {
	host, _ := config["host"].(string)
	port := 25
	if p, ok := config["port"].(float64); ok {
		port = int(p)
	}
	username, _ := config["username"].(string)
	password, _ := config["password"].(string) // 如果为空，则不使用认证
	to, _ := config["to"].(string)
	subject, _ := config["subject"].(string)
	content, _ := config["content"].(string)

	addr := fmt.Sprintf("%s:%d", host, port)
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+ // [新增]
		"Content-Type: text/html; charset=UTF-8\r\n"+ // [关键：改为 text/html]
		"\r\n"+
		"%s", to, subject, content))

	// 1. 建立 TCP 连接
	c, err := smtp.Dial(addr)
	if err != nil {
		return StepResult{Status: "failed", Error: "Dial error: " + err.Error()}
	}
	defer c.Close()

	// 2. 关键修复：配置跳过 TLS 验证
	// 解决 image_a2101e.png 中的 x509 证书校验问题
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	if err = c.StartTLS(tlsConfig); err != nil {
		// 如果服务器不支持 TLS，这里报错可以忽略，或者根据实际情况降级到普通连接
		// 针对内网 SMTP，有些服务器可能完全不提供 StartTLS
		fmt.Printf("Warning: StartTLS failed: %v\n", err)
	}

	// 3. 关键修复：支持无认证模式
	if password != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if err = c.Auth(auth); err != nil {
			return StepResult{Status: "failed", Error: "Auth error: " + err.Error()}
		}
	}

	// 4. 发送邮件流程
	if err = c.Mail(username); err != nil {
		return StepResult{Status: "failed", Error: err.Error()}
	}
	for _, addr := range strings.Split(to, ",") {
		if err = c.Rcpt(strings.TrimSpace(addr)); err != nil {
			return StepResult{Status: "failed", Error: err.Error()}
		}
	}
	w, err := c.Data()
	if err != nil {
		return StepResult{Status: "failed", Error: err.Error()}
	}
	_, err = w.Write(msg)
	if err != nil {
		return StepResult{Status: "failed", Error: err.Error()}
	}
	err = w.Close()
	if err != nil {
		return StepResult{Status: "failed", Error: err.Error()}
	}

	return StepResult{Status: "success", Output: "Email sent successfully"}
}

// RunExpression 执行纯 Expr 表达式节点 (Pro-Code)
func RunExpression(config map[string]interface{}, ctx *ExecutionContext) StepResult {
	exprStr, _ := config["expression"].(string)
	env := createExprEnv(ctx)

	program, err := expr.Compile(exprStr, expr.Env(env))
	if err != nil {
		return StepResult{Status: "failed", Error: "Compile error: " + err.Error()}
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return StepResult{Status: "failed", Error: "Run error: " + err.Error()}
	}

	return StepResult{Status: "success", Output: output}
}

// RunCondition 执行条件分支判断
func RunCondition(config map[string]interface{}, ctx *ExecutionContext) StepResult {
	// 如果配置了 expression 优先走 Pro-Code 逻辑
	if exprStr, ok := config["expression"].(string); ok && exprStr != "" {
		return RunExpression(config, ctx)
	}

	// 否则走 Low-Code 列表判断 (省略重复逻辑，建议前端统一传 expression)
	return StepResult{Status: "failed", Error: "No valid expression found"}
}
