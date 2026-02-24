import { useRef, useEffect } from 'react';
import Editor, { useMonaco } from '@monaco-editor/react';
import type { OnMount }from '@monaco-editor/react';
import { logSQLLanguageDef, defineLogSQLTheme } from '@/lib/monaco/logsql-lang';
import { registerLogsQL } from '@/lib/monaco/completion';

interface LogSQLEditorProps {
  value: string;
  onChange: (value: string | undefined) => void;
  onRun?: () => void;
}

export function LogSQLEditor({ value, onChange, onRun }: LogSQLEditorProps) {
  const monaco = useMonaco();
  const editorRef = useRef<any>(null);

  useEffect(() => {
    if (monaco) {
      if (!monaco.languages.getLanguages().some(l => l.id === 'logsql')) {
        monaco.languages.register({ id: 'logsql' });
        monaco.languages.setMonarchTokensProvider('logsql', logSQLLanguageDef);
        defineLogSQLTheme(monaco);
        registerLogsQL(monaco);
      }
    }
  }, [monaco]);

  const handleEditorDidMount: OnMount = (editor, monaco) => {
    editorRef.current = editor;
    // 绑定 Ctrl+Enter 运行
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
      onRun?.();
    });
  };

  return (
    <div className="h-full w-full overflow-hidden rounded-md border border-border bg-[#1e1e1e]">
      <Editor
        height="100%"
        defaultLanguage="logsql"
        theme="logsql-dark"
        value={value}
        onChange={onChange}
        onMount={handleEditorDidMount}
        options={{
          fontFamily: "'Fira Code', monospace",
          fontSize: 14,
          minimap: { enabled: false },
          scrollBeyondLastLine: false,
          automaticLayout: true,
          padding: { top: 16 }
        }}
      />
    </div>
  )
}