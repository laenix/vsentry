import * as monaco from 'monaco-editor';
import { registerExprLanguage } from './expr-lang';

// expr 常用内置函数列表
const EXPR_FUNCTIONS = [
  { label: 'len', insertText: 'len($1)', detail: 'Returns the length of array or string', insertTextRules: 4 },
  { label: 'sprintf', insertText: 'sprintf("%s", $1)', detail: 'Format string using placeholders', insertTextRules: 4 },
  { label: 'contains', insertText: 'contains($1, "$2")', detail: 'Check if string contains substring', insertTextRules: 4 },
  { label: 'filter', insertText: 'filter($1, { .# $2 })', detail: 'Filter array elements', insertTextRules: 4 },
  { label: 'map', insertText: 'map($1, { .# $2 })', detail: 'Transform array elements', insertTextRules: 4 },
  { label: 'now', insertText: 'now()', detail: 'Current timestamp', insertTextRules: 4 },
];

/**
 * 注册 Automation 表达式自动补全
 * @param monacoInstance Monaco 实例
 * @param contextData 动态上下文数据 (来自 Test Run 的 PlaybookExecution.logs)
 */
export const registerAutomationCompletion = (
  monacoInstance: typeof monaco,
  contextData?: Record<string, any>
): monaco.IDisposable => {

  // 1. 注册语言高亮定义 (如标识符颜色、操作符等)
  registerExprLanguage(monacoInstance);

  const langId = 'expr'; 

  return monacoInstance.languages.registerCompletionItemProvider(langId, {
    triggerCharacters: ['.'],

    provideCompletionItems: (model, position) => {
      const word = model.getWordUntilPosition(position);
      const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn,
      };

      const textUntilPosition = model.getValueInRange({
        startLineNumber: position.lineNumber,
        startColumn: 1,
        endLineNumber: position.lineNumber,
        endColumn: position.column,
      });

      // --- 场景 1: 顶层关键字 (incident, steps, functions) ---
      if (!textUntilPosition.includes('.')) {
        const keywords = [
          {
            label: 'incident',
            kind: monacoInstance.languages.CompletionItemKind.Variable,
            insertText: 'incident',
            documentation: 'The global incident or trigger context',
            range,
          },
          {
            label: 'steps',
            kind: monacoInstance.languages.CompletionItemKind.Variable,
            insertText: 'steps',
            documentation: 'Reference results from previous nodes',
            range,
          },
        ];

        const functions = EXPR_FUNCTIONS.map(fn => ({
          ...fn,
          kind: monacoInstance.languages.CompletionItemKind.Function,
          range
        }));

        return { suggestions: [...keywords, ...functions] };
      }

      // --- 场景 2: 动态属性补全 ---

      // 2.1 Incident 补全：映射到 ID 为 "1" 的节点或第一个节点
      if (textUntilPosition.endsWith('incident.')) {
        // 自动识别触发器节点：优先找 ID "trigger"，否则找 logs 里的首个 key
        const triggerKey = contextData ? (Object.keys(contextData).find(k => k === "trigger" || !k.includes('_')) || "trigger") : "trigger";
        const triggerOutput = contextData?.[triggerKey]?.output;
        
        if (triggerOutput && typeof triggerOutput === 'object') {
          return {
            suggestions: Object.keys(triggerOutput).map(key => ({
              label: key,
              kind: monacoInstance.languages.CompletionItemKind.Field,
              insertText: key,
              detail: `Mock/Real Value: ${JSON.stringify(triggerOutput[key])}`,
              range
            }))
          };
        }

        // 默认静态提示 (如果 Test Run 尚未执行)
        return {
          suggestions: [
            { label: 'id', kind: monacoInstance.languages.CompletionItemKind.Field, insertText: 'id', range },
            { label: 'name', kind: monacoInstance.languages.CompletionItemKind.Field, insertText: 'name', range },
            { label: 'severity', kind: monacoInstance.languages.CompletionItemKind.Field, insertText: 'severity', range },
          ]
        };
      }

      // 2.2 Steps 动态补全 (核心：深度探测 JSON 结构)
      if (textUntilPosition.startsWith('steps.')) {
        const pathString = textUntilPosition.slice(6); 
        const parts = pathString.split('.');
        const isTriggeredByDot = textUntilPosition.endsWith('.');
        
        // 2.2.1 提示 Node ID
        if (parts.length === 1 && !isTriggeredByDot) {
          const nodeIds = contextData ? Object.keys(contextData) : [];
          return {
            suggestions: nodeIds.map(nodeId => ({
              label: nodeId,
              kind: monacoInstance.languages.CompletionItemKind.Class,
              insertText: nodeId, // ✅ 如果后端改成了字母 ID，这里会自动适配
              detail: `Execution Status: ${contextData?.[nodeId]?.status}`,
              range
            }))
          };
        }

        // 2.2.2 提示子属性 (output, body, etc.)
        if (contextData) {
          const pathParts = isTriggeredByDot ? parts.slice(0, -1) : parts.slice(0, -1);
          let currentObj: any = contextData;
          
          for (const part of pathParts) {
            if (currentObj && typeof currentObj === 'object' && part in currentObj) {
              currentObj = currentObj[part];
            } else {
              currentObj = null;
              break;
            }
          }

          if (currentObj && typeof currentObj === 'object') {
            return {
              suggestions: Object.keys(currentObj).map(key => {
                const val = currentObj[key];
                return {
                  label: key,
                  kind: monacoInstance.languages.CompletionItemKind.Field,
                  insertText: key,
                  detail: typeof val === 'object' ? 'Object' : String(val),
                  range
                };
              })
            };
          }
        }
      }

      return { suggestions: [] };
    },
  });
};