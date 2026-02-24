import * as monaco from 'monaco-editor';

// 1. 定义 LogsQL 核心关键字
const LOGSQL_KEYWORDS = [
  '_time', '_stream', '_msg', 'AND', 'OR', 'NOT', 'IN', 'ANY', 'EXACT'
];

// 2. 定义管道函数及常用过滤函数
const PIPE_FUNCTIONS = [
  'stats', 'limit', 'sort', 'fields', 'filter', 'copy', 'delete', 'rename', 'extract'
];

// 3. 统计聚合函数
const AGGREGATE_FUNCTIONS = [
  'count', 'count_empty', 'sum', 'min', 'max', 'avg', 'median', 'p50', 'p90', 'p95', 'p99'
];

/**
 * 注册 LogsQL 语言及其补全提供者
 */
export function registerLogsQL(monacoInstance: typeof monaco) {
  const languageId = 'logsql';

  // 注册语言 ID
  monacoInstance.languages.register({ id: languageId });

  // 设置语法高亮 (Tokenization)
  monacoInstance.languages.setMonarchTokensProvider(languageId, {
    tokenizer: {
      root: [
        [/[_a-zA-Z][_a-zA-Z0-9]*/, {
          cases: {
            '@LOGSQL_KEYWORDS': 'keyword',
            '@PIPE_FUNCTIONS': 'predefined',
            '@default': 'identifier'
          }
        }],
        [/[{}()\[\]]/, 'delimiter'],
        [/".*?"/, 'string'],
        [/\d+/, 'number'],
        [/\|/, 'operator'], // 管道符
      ]
    },
    LOGSQL_KEYWORDS,
    PIPE_FUNCTIONS
  });

  // 注册自动补全提供者
  monacoInstance.languages.registerCompletionItemProvider(languageId, {
    triggerCharacters: [' ', '|', ':', '"', '.'],
    provideCompletionItems: (model, position) => {
      const word = model.getWordUntilPosition(position);
      const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn
      };

      const suggestions: monaco.languages.CompletionItem[] = [
        // 关键字补全
        ...LOGSQL_KEYWORDS.map(k => ({
          label: k,
          kind: monacoInstance.languages.CompletionItemKind.Keyword,
          insertText: k,
          range
        })),
        // 管道函数补全
        ...PIPE_FUNCTIONS.map(f => ({
          label: f,
          kind: monacoInstance.languages.CompletionItemKind.Function,
          insertText: f,
          range,
          detail: 'LogsQL Pipe Function'
        })),
        // 聚合函数补全
        ...AGGREGATE_FUNCTIONS.map(a => ({
          label: a,
          kind: monacoInstance.languages.CompletionItemKind.Snippet,
          insertText: `${a}(\${1:field})`,
          insertTextRules: monacoInstance.languages.CompletionItemInsertTextRule.InsertAsSnippet,
          range,
          detail: 'Aggregate Function'
        }))
      ];

      return { suggestions };
    }
  });

  // 设置语言配置 (括号匹配等)
  monacoInstance.languages.setLanguageConfiguration(languageId, {
    surroundingPairs: [
      { open: '{', close: '}' },
      { open: '[', close: ']' },
      { open: '(', close: ')' },
      { open: '"', close: '"' },
    ],
    autoClosingPairs: [
      { open: '{', close: '}' },
      { open: '[', close: ']' },
      { open: '(', close: ')' },
      { open: '"', close: '"' },
    ]
  });
}