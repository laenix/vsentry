import * as monaco from 'monaco-editor';

export const languageID = 'expr';

export const exprLanguageDef: monaco.languages.IMonarchLanguage = {
  defaultToken: '',
  tokenPostfix: '.expr',

  keywords: [
    'all', 'any', 'filter', 'map', 'none', 'one', 'count',
    'len', 'type', 'abs', 'int', 'float', 'string', 'trim', 
    'upper', 'lower', 'split', 'replace', 'now', 'duration', 'date',
    'first', 'last', 'get', 'contains', 'matches', 'startsWith', 'endsWith',
    'in', 'not', 'and', 'or', 'matches', 'true', 'false', 'nil', 'null'
  ],

  operators: [
    '=', '==', '!=', '>', '<', '>=', '<=', '+', '-', '*', '/', '%', '??', '?', ':'
  ],

  symbols: /[=><!~?:&|+\-*\/\^%]+/,

  tokenizer: {
    root: [
      // Identifiers and keywords
      [/[a-zA-Z_]\w*/, {
        cases: {
          '@keywords': 'keyword',
          '@default': 'identifier'
        }
      }],

      // Whitespace
      { include: '@whitespace' },

      // Delimiters and operators
      [/[{}()\[\]]/, '@brackets'],
      [/@symbols/, {
        cases: {
          '@operators': 'operator',
          '@default': ''
        }
      }],

      // Numbers
      [/\d*\.\d+([eE][\-+]?\d+)?/, 'number.float'],
      [/\d+/, 'number'],

      // Strings
      [/"([^"\\]|\\.)*$/, 'string.invalid'],  // non-teminated string
      [/"/, { token: 'string.quote', bracket: '@open', next: '@string' }],
      [/'/, { token: 'string.quote', bracket: '@open', next: '@stringSingle' }],
      
      // Variables (Context variables like incident.id)
      [/incident\.[a-zA-Z_]\w*/, 'variable.predefined'],
      [/steps\.[a-zA-Z_]\w*/, 'variable.predefined'],
    ],

    string: [
      [/[^\\"]+/, 'string'],
      [/\\./, 'string.escape'],
      [/"/, { token: 'string.quote', bracket: '@close', next: '@pop' }]
    ],

    stringSingle: [
      [/[^\\']+/, 'string'],
      [/\\./, 'string.escape'],
      [/'/, { token: 'string.quote', bracket: '@close', next: '@pop' }]
    ],

    whitespace: [
      [/[ \t\r\n]+/, 'white'],
      [/\/\/.*$/, 'comment'],
    ],
  },
};

export const registerExprLanguage = (monacoInstance: typeof monaco) => {
  if (!monacoInstance.languages.getLanguages().some(l => l.id === languageID)) {
    monacoInstance.languages.register({ id: languageID });
    monacoInstance.languages.setMonarchTokensProvider(languageID, exprLanguageDef);
    
    // 设置一些基本的补全配置（括号匹配等）
    monacoInstance.languages.setLanguageConfiguration(languageID, {
      brackets: [
        ['{', '}'],
        ['[', ']'],
        ['(', ')'],
      ],
      autoClosingPairs: [
        { open: '{', close: '}' },
        { open: '[', close: ']' },
        { open: '(', close: ')' },
        { open: '"', close: '"' },
        { open: "'", close: "'" },
      ]
    });
  }
};