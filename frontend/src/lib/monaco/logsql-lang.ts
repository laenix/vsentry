import * as monaco from 'monaco-editor';

export const logSQLLanguageDef: monaco.languages.IMonarchLanguage = {
  defaultToken: '',
  tokenPostfix: '.logsql',
  ignoreCase: true,

  keywords: [
    '_stream', '_time', '_msg', 'error', 'warn', 'info', 'debug',
    'stats', 'uniq', 'sort', 'limit', 'offset', 'fields',
    'filter', 'union', 'or', 'and', 'not', 'in', 'by'
  ],

  operators: [
    '|', '=', '!=', ':=', '>', '<', '>=', '<=', '+', '-', '*', '/', '%'
  ],

  tokenizer: {
    root: [
      // 管道符高亮 (亮橙色)
      [/\|/, 'keyword.flow'],
      
      // 字段名 (浅蓝色)
      [/[a-zA-Z_][\w$]*/, {
        cases: {
          '@keywords': 'keyword',
          '@default': 'identifier'
        }
      }],

      // 数字
      [/\d+/, 'number'],

      // 字符串
      [/"/, { token: 'string.quote', bracket: '@open', next: '@string' }],
      [/`/, { token: 'string.quote', bracket: '@open', next: '@backtick' }],
    ],

    string: [
      [/[^\\"]+/, 'string'],
      [/"/, { token: 'string.quote', bracket: '@close', next: '@pop' }]
    ],

    backtick: [
      [/[^\\`]+/, 'string'],
      [/`/, { token: 'string.quote', bracket: '@close', next: '@pop' }]
    ]
  }
};

export const defineLogSQLTheme = (monacoInstance: typeof monaco) => {
  monacoInstance.editor.defineTheme('logsql-dark', {
    base: 'vs-dark',
    inherit: true,
    rules: [
      { token: 'keyword.flow', foreground: 'FF9900', fontStyle: 'bold' },
      { token: 'keyword', foreground: 'C586C0' },
      { token: 'identifier', foreground: '9CDCFE' },
      { token: 'string', foreground: 'CE9178' },
      { token: 'number', foreground: 'B5CEA8' },
    ],
    colors: {
      'editor.background': '#1e1e1e',
    }
  });
};