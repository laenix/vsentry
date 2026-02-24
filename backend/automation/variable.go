package automation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
)

// createExprEnv 构造 expr 执行环境，解决结构体字段访问权限问题
func createExprEnv(ctx *ExecutionContext) map[string]interface{} {
	stepsMap := make(map[string]interface{})
	for id, result := range ctx.Steps {
		stepsMap[id] = map[string]interface{}{
			"status": result.Status,
			"output": result.Output,
			"error":  result.Error,
		}
	}

	// 构造基础环境映射
	env := map[string]interface{}{
		"incident": ctx.Global["incident"],
		"steps":    stepsMap,
		"env":      ctx.Global, // 包含手动触发注入的 context

		// [核心修复] 注册常用工具函数
		"sprintf": fmt.Sprintf,
		"len":     func(v interface{}) int { return 0 }, // 可根据需要补充
		"to_table_markdown": func(data interface{}, columns ...string) string {
			// 将接口转换为 slice
			list, ok := data.([]interface{})
			if !ok || len(list) == 0 {
				return "No Data"
			}

			var sb strings.Builder
			// 1. 构造表头
			sb.WriteString("| " + strings.Join(columns, " | ") + " |\n")
			sb.WriteString("| " + strings.Repeat(" --- |", len(columns)) + "\n")

			// 2. 构造行
			for _, item := range list {
				row, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				var rowValues []string
				for _, col := range columns {
					val := row[col]
					rowValues = append(rowValues, fmt.Sprintf("%v", val))
				}
				sb.WriteString("| " + strings.Join(rowValues, " | ") + " |\n")
			}
			return sb.String()
		},

		"to_table_html": func(data interface{}, columns ...string) string {
			// 1. 类型安全检查
			list, ok := data.([]interface{})
			if !ok || len(list) == 0 {
				return "<p style='color: gray;'>No Data Available</p>"
			}

			var sb strings.Builder
			// 2. 写入简单的内联样式 (邮件客户端对外部 CSS 支持较差)
			sb.WriteString("<table border='1' cellpadding='5' cellspacing='0' style='border-collapse: collapse; width: 100%; font-family: sans-serif; font-size: 14px;'>")

			// 3. 构造表头
			sb.WriteString("<tr style='background-color: #f2f2f2; text-align: left;'>")
			for _, col := range columns {
				sb.WriteString(fmt.Sprintf("<th style='padding: 8px; border: 1px solid #ddd;'>%s</th>", col))
			}
			sb.WriteString("</tr>")

			// 4. 构造行数据
			for _, item := range list {
				row, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				sb.WriteString("<tr>")
				for _, col := range columns {
					val := row[col]
					// 处理空值或格式化
					valStr := fmt.Sprintf("%v", val)
					if val == nil {
						valStr = "-"
					}
					sb.WriteString(fmt.Sprintf("<td style='padding: 8px; border: 1px solid #ddd;'>%s</td>", valStr))
				}
				sb.WriteString("</tr>")
			}
			sb.WriteString("</table>")

			return sb.String()
		},
	}
	return env
}

// ResolveVariables 递归解析 Config 中的 {{...}} 变量
func ResolveVariables(config map[string]interface{}, ctx *ExecutionContext) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	env := createExprEnv(ctx) // 统一调用环境构造

	for k, v := range config {
		strVal, ok := v.(string)
		if !ok {
			result[k] = v
			continue
		}

		re := regexp.MustCompile(`\{\{(.*?)\}\}`)
		matches := re.FindAllStringSubmatch(strVal, -1)

		if len(matches) == 0 {
			result[k] = v
			continue
		}

		newStr := strVal
		for _, match := range matches {
			fullMatch := match[0]
			expression := strings.TrimSpace(match[1])

			program, err := expr.Compile(expression, expr.Env(env))
			if err != nil {
				return nil, fmt.Errorf("compile error in %s: %v", k, err)
			}

			output, err := expr.Run(program, env)
			if err != nil {
				return nil, fmt.Errorf("run error in %s: %v", k, err)
			}

			newStr = strings.Replace(newStr, fullMatch, fmt.Sprintf("%v", output), 1)
		}
		result[k] = newStr
	}
	return result, nil
}
