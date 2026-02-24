import Editor from "@monaco-editor/react";

interface ReadOnlyJsonViewerProps {
  value: string;
  height?: string;
  className?: string;
}

export function ReadOnlyJsonViewer({ value, height = "400px", className }: ReadOnlyJsonViewerProps) {
  // 尝试格式化 JSON，如果不是合法 JSON 则显示原文
  let formattedValue = value;
  try {
    const jsonObj = JSON.parse(value);
    formattedValue = JSON.stringify(jsonObj, null, 2);
  } catch (e) {
    // 保持原样
  }

  return (
    // 关键修复：将 height 传给最外层容器，否则 Editor 的 height="100%" 会塌陷
    <div 
      className={`border rounded-md overflow-hidden bg-[#1e1e1e] ${className || ""}`} 
      style={{ height }}
    >
      <Editor
        height="100%" // 让 Editor 填满父容器
        defaultLanguage="json"
        value={formattedValue}
        theme="vs-dark"
        options={{
          readOnly: true,
          minimap: { enabled: false }, // 关闭缩略图
          scrollBeyondLastLine: false,
          fontSize: 12,
          wordWrap: "on", // 自动换行
          automaticLayout: true, // 关键：自适应容器大小变化
          domReadOnly: true,
          lineNumbers: "off", // 可选：看你想不想要行号，off 比较像纯文本展示
          folding: true,
          renderLineHighlight: "none",
          scrollbar: {
            vertical: "auto",
            horizontal: "hidden"
          }
        }}
      />
    </div>
  );
}