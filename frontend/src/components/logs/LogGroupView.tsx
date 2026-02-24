import React, { useState, useMemo, useRef, useEffect } from "react"
import { useVirtualizer } from "@tanstack/react-virtual"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Badge } from "@/components/ui/badge"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Checkbox } from "@/components/ui/checkbox"
import { Layers, Columns, ListFilter } from "lucide-react"
import type { VLResult } from "@/lib/api/vl-client"
import { cn } from "@/lib/utils"
import { LogEntryItem } from "./LogEntryItem"

export function LogGroupView({ data }: { data: VLResult[] }) {
  const [selectedStream, setSelectedStream] = useState<string | null>(null)
  const detailParentRef = useRef<HTMLDivElement>(null)

  // 1. 提取所有可用字段 (用于下拉筛选)
  const allFields = useMemo(() => {
    if (!data || data.length === 0) return []
    const keys = new Set<string>()
    // 采样前20条，避免性能问题
    data.slice(0, 20).forEach(log => Object.keys(log).forEach(k => keys.add(k)))
    return Array.from(keys).sort()
  }, [data])

  // 2. 显示列状态
  const [visibleColumns, setVisibleColumns] = useState<string[]>(["_time", "_msg"])

  // --- 核心修改: 智能列检测逻辑 (与 LogTableView 逻辑保持一致) ---
  useEffect(() => {
    if (!data || data.length === 0) return

    const firstRow = data[0]
    const keys = Object.keys(firstRow)
    const isRawLog = keys.includes("_time")

    if (isRawLog) {
      // 日志模式: 默认显示时间、消息
      // 检查当前 visibleColumns 是否有效，如果无效则重置
      const hasValid = visibleColumns.some(c => keys.includes(c))
      if (!hasValid || visibleColumns.length === 0) {
        setVisibleColumns(["_time", "_msg"])
      }
    } else {
      // 聚合模式: 显示数据中存在的所有字段 (如 count, DestPort)
      const currentKeysStr = visibleColumns.sort().join(",")
      const newKeysStr = keys.sort().join(",")
      
      // 只有当字段结构发生变化时才更新，防止死循环
      if (currentKeysStr !== newKeysStr) {
         setVisibleColumns(Object.keys(firstRow))
      }
    }
  }, [data])

  // 3. 分组逻辑
  const groups = useMemo(() => {
    const map: Record<string, VLResult[]> = {}
    data.forEach(log => {
      // 聚合结果通常没有 _stream 字段，我们将其归类为 "{}"
      const s = log._stream || "{}" 
      if (!map[s]) map[s] = []
      map[s].push(log)
    })
    return Object.entries(map)
      .map(([stream, logs]) => ({ stream, logs, count: logs.length }))
      .sort((a, b) => b.count - a.count)
  }, [data])

  // 自动选择第一个分组
  useEffect(() => {
    // 只有当没有选中，或者选中的流在当前数据中不存在时，才自动选择第一个
    const exists = groups.find(g => g.stream === selectedStream)
    if ((!selectedStream || !exists) && groups.length > 0) {
      setSelectedStream(groups[0].stream)
    }
  }, [groups, selectedStream])

  const activeGroup = groups.find(g => g.stream === selectedStream)

  // 4. 虚拟滚动配置
  const rowVirtualizer = useVirtualizer({
    count: activeGroup?.logs.length || 0,
    getScrollElement: () => detailParentRef.current,
    estimateSize: () => 40, // 预估高度
    overscan: 10,
    measureElement: (el) => el.getBoundingClientRect().height, // 支持动态高度
  })

  // 切换流时重置滚动位置
  useEffect(() => {
    rowVirtualizer.scrollToOffset(0)
  }, [selectedStream])

  if (!data || data.length === 0) {
     return <div className="flex h-full items-center justify-center text-xs text-muted-foreground">No data available</div>
  }

  return (
    <div className="flex h-full divide-x border-t bg-background overflow-hidden font-sans">
      
      {/* --- 左侧 Stream 列表 --- */}
      <div className="w-[280px] flex flex-col bg-muted/5 flex-none border-r">
        <div className="p-2.5 border-b bg-muted/10 text-[10px] font-bold uppercase text-muted-foreground/70 tracking-widest flex items-center">
          <Layers className="w-3 h-3 mr-2 opacity-70" />
          Streams ({groups.length})
        </div>
        <ScrollArea className="flex-1">
          <div className="p-1.5 space-y-1">
            {groups.map((group) => (
              <div
                key={group.stream}
                onClick={() => setSelectedStream(group.stream)}
                className={cn(
                  "p-2 rounded-md cursor-pointer border transition-all select-none",
                  selectedStream === group.stream
                    ? "bg-primary/10 border-primary/20 text-primary shadow-sm"
                    : "hover:bg-muted border-transparent text-muted-foreground"
                )}
              >
                <div className="flex justify-between items-center mb-1">
                  {/* 对聚合结果 ("{}") 进行特殊命名展示 */}
                  <Badge variant="outline" className="text-[9px] px-1 h-4 font-mono bg-background/50">
                    {group.count}
                  </Badge>
                </div>
                <div className="text-[10px] font-mono break-all leading-tight opacity-90 line-clamp-2" title={group.stream}>
                  {group.stream === "{}" ? "(Aggregated Results)" : group.stream}
                </div>
              </div>
            ))}
          </div>
        </ScrollArea>
      </div>

      {/* --- 右侧日志详情 --- */}
      <div className="flex-1 flex flex-col min-w-0 bg-background overflow-hidden">
        
        {/* 工具栏: 显示当前流名称 + 字段切换 */}
        <div className="p-1 border-b bg-muted/10 flex items-center justify-between px-4 h-9 shrink-0">
           <div className="flex items-center gap-2 min-w-0">
             <ListFilter className="w-3 h-3 text-primary" />
             <span className="text-[10px] font-mono text-primary truncate italic max-w-[400px]" title={selectedStream || ""}>
               {selectedStream === "{}" ? "Aggregated Results" : selectedStream}
             </span>
           </div>

           <Popover>
             <PopoverTrigger asChild>
               <button className="flex items-center gap-1.5 text-[9px] font-bold uppercase text-muted-foreground hover:text-foreground border rounded px-2 py-1 bg-background transition-colors hover:bg-muted/50">
                 <Columns className="w-3 h-3" /> Fields
               </button>
             </PopoverTrigger>
             <PopoverContent className="w-64 p-0" align="end">
               <div className="p-2 border-b bg-muted/10 text-[10px] font-bold text-muted-foreground uppercase">
                 Toggle Display Fields
               </div>
               <div className="max-h-[300px] overflow-y-auto p-1 space-y-0.5">
                 {allFields.map(f => (
                   <div 
                     key={f} 
                     className="flex items-center gap-2 p-1.5 hover:bg-muted/50 rounded cursor-pointer transition-colors"
                     onClick={() => {
                        setVisibleColumns(prev => prev.includes(f) ? prev.filter(x => x !== f) : [...prev, f])
                     }}
                   >
                     <Checkbox 
                        id={`gf-${f}`}
                        checked={visibleColumns.includes(f)}
                        className="w-3.5 h-3.5"
                     />
                     <label htmlFor={`gf-${f}`} className="text-[11px] font-mono cursor-pointer truncate flex-1">{f}</label>
                   </div>
                 ))}
               </div>
             </PopoverContent>
           </Popover>
        </div>

        {/* 虚拟滚动列表 */}
        <div ref={detailParentRef} className="flex-1 overflow-auto scrollbar-thin scrollbar-thumb-muted-foreground/20">
           <div
             className="relative w-full"
             style={{ height: `${rowVirtualizer.getTotalSize()}px` }}
           >
             {rowVirtualizer.getVirtualItems().map((virtualRow) => (
               <div
                 key={virtualRow.key}
                 data-index={virtualRow.index}
                 ref={rowVirtualizer.measureElement}
                 className="absolute top-0 left-0 w-full"
                 style={{ transform: `translateY(${virtualRow.start}px)` }}
               >
                 {/* 确保 LogEntryItem 能接收 visibleColumns 并动态渲染 */}
                 <LogEntryItem 
                   log={activeGroup!.logs[virtualRow.index]} 
                   visibleColumns={visibleColumns} 
                 />
               </div>
             ))}
           </div>
        </div>
      </div>
    </div>
  )
}