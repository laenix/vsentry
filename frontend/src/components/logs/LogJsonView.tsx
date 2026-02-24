import React, { useRef } from "react"
import { useVirtualizer } from "@tanstack/react-virtual"
import { Copy, Terminal, Braces } from "lucide-react"
import type { VLResult } from "@/lib/api/vl-client"
import { Button } from "@/components/ui/button"

interface LogJsonViewProps {
  data: VLResult[]
}

export function LogJsonView({ data }: LogJsonViewProps) {
  const parentRef = useRef<HTMLDivElement>(null)

  // 1. 虚拟滚动配置：支持动态高度测量
  // JSON 块的高度差异巨大，必须使用 measureElement 实时计算
  const rowVirtualizer = useVirtualizer({
    count: data.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 180, // 给一个合理的初始预估高度
    overscan: 5, // 预渲染 5 项，防止滚动过快白屏
    measureElement: (el) => el.getBoundingClientRect().height,
  })

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text)
    // 如果有 toast 组件：toast.success("JSON copied to clipboard")
  }

  if (data.length === 0) {
    return (
      <div className="flex h-full flex-col items-center justify-center text-muted-foreground/50 gap-3">
        <Braces className="w-10 h-10 opacity-20" />
        <span className="text-xs font-mono">No JSON data available</span>
      </div>
    )
  }

  return (
    <div 
      ref={parentRef} 
      className="h-full w-full overflow-auto bg-background scrollbar-thin p-2"
    >
      <div
        className="relative w-full"
        style={{ height: `${rowVirtualizer.getTotalSize()}px` }}
      >
        {rowVirtualizer.getVirtualItems().map((virtualRow) => {
          const log = data[virtualRow.index]
          return (
            <div
              key={virtualRow.key}
              data-index={virtualRow.index} // ✅ 必须：供 measureElement 识别
              ref={rowVirtualizer.measureElement} // ✅ 必须：绑定测量 ref
              className="absolute top-0 left-0 w-full px-2 py-1.5"
              style={{
                transform: `translateY(${virtualRow.start}px)`,
              }}
            >
              <div className="relative group border border-border/60 rounded-md bg-muted/10 hover:bg-muted/20 transition-colors shadow-sm">
                
                {/* 装饰性头部：显示行号和时间 */}
                <div className="flex items-center justify-between px-3 py-1.5 border-b border-border/40 bg-muted/20 rounded-t-md select-none">
                  <div className="flex items-center gap-2">
                    <Terminal className="w-3 h-3 text-emerald-600/70" />
                    <span className="text-[10px] font-bold font-mono text-muted-foreground/70">
                      Row {virtualRow.index + 1}
                    </span>
                  </div>
                  <span className="text-[9px] font-mono text-muted-foreground/40">
                    {log._time || "timestamp_missing"}
                  </span>
                </div>

                {/* JSON 内容区 */}
                <pre className="p-3 text-[10px] font-mono whitespace-pre-wrap break-all leading-relaxed text-foreground/90 overflow-hidden">
                  {JSON.stringify(log, null, 2)}
                </pre>

                {/* 悬浮复制按钮 */}
                <Button
                  variant="secondary"
                  size="sm"
                  className="absolute top-1.5 right-2 h-6 px-2 text-[9px] opacity-0 group-hover:opacity-100 transition-all bg-background border shadow-sm hover:bg-primary hover:text-primary-foreground"
                  onClick={() => handleCopy(JSON.stringify(log, null, 2))}
                >
                  <Copy className="w-3 h-3 mr-1.5" /> Copy JSON
                </Button>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}