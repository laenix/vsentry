import * as monaco from 'monaco-editor';

// 1. 定义 LogsQL 核心关键字
const LOGSQL_KEYWORDS = [
  '_time', '_stream', '_msg', 'AND', 'OR', 'NOT', 'IN', 'ANY', 'EXACT'
];

// 2. 定义管道Function及常用FilterFunction
const PIPE_FUNCTIONS = [
  'stats', 'limit', 'sort', 'fields', 'filter', 'copy', 'delete', 'rename', 'extract'
];

// 3. 统计聚合Function
const AGGREGATE_FUNCTIONS = [
  'count', 'count_empty', 'sum', 'min', 'max', 'avg', 'median', 'p50', 'p90', 'p95', 'p99'
];

/**
 * 注册 LogsQL 语言及其补全提供者
 */
export function registerLogsQL(monacoInstance: typeof monaco) {
  const languageId = 'logsql';

  // Register语言 ID
  monacoInstance.languages.register({ id: languageId });

  // Settings语法High亮 (Tokenization)
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

  // Register自动补全提供者
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
        // 管道Function补全
        ...PIPE_FUNCTIONS.map(f => ({
          label: f,
          kind: monacoInstance.languages.CompletionItemKind.Function,
          insertText: f,
          range,
          detail: 'LogsQL Pipe Function'
        })),
        // 聚合Function补全
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

  // Settings语言配置 (括号匹配等)
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