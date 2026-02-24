import React, { useEffect, useRef } from 'react';
import Editor, { useMonaco } from '@monaco-editor/react';
import { registerAutomationCompletion } from '@/lib/monaco/automation-completion';

interface MonacoInputProps {
  value: string;
  onChange: (value: string) => void;
  height?: string;
  language?: string;
  placeholder?: string;
  // ✅ 新增：接收动态上下文数据 (来自 Test Run 的结果)
  contextData?: Record<string, any>;
}

export function MonacoInput({ 
  value, 
  onChange, 
  height = "32px", 
  language = "expr", 
  contextData 
}: MonacoInputProps) {
  const monaco = useMonaco();
  const editorRef = useRef<any>(null);

  // ✅ 核心：当 monaco 实例或 contextData 变化时，重新注册补全逻辑
  useEffect(() => {
    if (monaco) {
      // 注册并获取销毁函数 (Disposable)
      const disposable = registerAutomationCompletion(monaco, contextData);
      
      // 清理函数：组件卸载或数据更新时，销毁旧的补全注册，防止重复
      return () => {
        disposable.dispose();
      };
    }
  }, [monaco, contextData]);

  const handleEditorDidMount = (editor: any, monaco: any) => {
    editorRef.current = editor;
    
    editor.updateOptions({
      minimap: { enabled: false },
      lineNumbers: 'off',
      glyphMargin: false,
      folding: false,
      lineDecorationsWidth: 0,
      lineNumbersMinChars: 0,
      renderLineHighlight: 'none',
      overviewRulerLanes: 0,
      hideCursorInOverviewRuler: true,
      scrollbar: { vertical: 'hidden', horizontal: 'hidden' },
      overviewRulerBorder: false,
      contextmenu: false,
      fontFamily: "'Fira Code', monospace",
      fontSize: 12,
      wordWrap: 'off',
      padding: { top: 6, bottom: 6 },
      // 关键：让补全菜单浮动在 body 上，避免被侧边栏遮挡
      fixedOverflowWidgets: true 
    });
  };

  return (
    <div className="border rounded-md overflow-hidden focus-within:ring-1 focus-within:ring-primary focus-within:border-primary transition-all bg-background">
      <Editor
        height={height}
        language={language}
        value={value}
        onChange={(val) => onChange(val || '')}
        onMount={handleEditorDidMount}
        theme="vs-light" 
        options={{
          scrollBeyondLastLine: false,
          automaticLayout: true,
          tabSize: 2,
          fixedOverflowWidgets: true,
          // 优化：单行模式下禁用回车换行
          acceptSuggestionOnEnter: "on",
        }}
      />
    </div>
  );
}