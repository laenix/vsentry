import React, { useState } from "react"
import { ChevronDown, ChevronRight, Copy, Terminal } from "lucide-react"
import { cn } from "@/lib/utils"

interface LogEntryItemProps {
  log: any
  visibleColumns: string[]
}

export function LogEntryItem({ log, visibleColumns }: LogEntryItemProps) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <div className={cn(
      "border-b border-border/40 bg-background transition-colors hover:bg-muted/20 group",
      isOpen && "bg-muted/5"
    )}>
      {/* === 未展开状态 (Collapsed) === */}
      <div 
        className="flex items-center gap-0 px-2 h-8 cursor-pointer w-full select-none"
        onClick={() => setIsOpen(!isOpen)}
      >
        {/* 1. 箭头图标 */}
        <div className="text-muted-foreground/40 shrink-0 w-6 flex justify-center">
          {isOpen ? <ChevronDown className="w-3.5 h-3.5" /> : <ChevronRight className="w-3.5 h-3.5" />}
        </div>
        
        {/* 2. 内容预览区 - 核心复刻 vmui
            - whitespace-nowrap: 核心！禁止换行
            - overflow-hidden: 超出父容器直接切断
        */}
        <div className="flex items-center gap-6 flex-1 overflow-hidden whitespace-nowrap min-w-0">
          {visibleColumns.map(col => (
            // shrink-0 确保字段不会因为宽度不够被挤扁，而是保持原样展示，直到被切断
            <div key={col} className="flex items-baseline gap-2 shrink-0">
              {/* Key: 灰色、大写、粗体 */}
              <span className="text-[10px] font-black text-muted-foreground/50 uppercase font-mono tracking-tight shrink-0">
                {col}:
              </span>
              
              {/* Value: 黑色、等宽。不再限制宽度，让它自然延伸 */}
              <span className="text-[11px] font-mono text-foreground/90 font-medium">
                {String(log[col] ?? "")}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* === 展开状态 (Expanded) === 
          需求：只展示 JSON，不重复展示字段列表
      */}
      {isOpen && (
        <div className="px-8 pb-3 pt-1 animate-in fade-in slide-in-from-top-1 duration-150 cursor-auto">
          {/* JSON 代码块 */}
          <div className="bg-muted/30 rounded border border-border/50 p-3 relative group/json shadow-sm">
            {/* 顶部标签 */}
            <div className="absolute top-2 left-3 text-[9px] font-bold text-muted-foreground/40 uppercase flex items-center gap-1 select-none">
              <Terminal className="w-3 h-3" /> JSON Source
            </div>

            {/* 内容 */}
            <pre className="text-[10px] font-mono whitespace-pre-wrap break-all text-muted-foreground leading-relaxed pt-5 pl-1">
              {JSON.stringify(log, null, 2)}
            </pre>

            {/* 复制按钮 */}
            <button 
              onClick={(e) => {
                e.stopPropagation();
                navigator.clipboard.writeText(JSON.stringify(log, null, 2));
              }}
              className="absolute top-2 right-2 opacity-0 group-hover/json:opacity-100 bg-background border px-2 py-1 rounded text-[9px] hover:bg-accent transition-all flex items-center gap-1 shadow-sm cursor-pointer z-10"
            >
              <Copy className="w-3 h-3" /> Copy
            </button>
          </div>
        </div>
      )}
    </div>
  )
}