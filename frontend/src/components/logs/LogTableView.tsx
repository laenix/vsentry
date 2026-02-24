import React, { useState, useMemo, useRef, useEffect } from "react"
import { useVirtualizer } from "@tanstack/react-virtual"
import { Filter, Settings2, GripVertical } from "lucide-react"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Checkbox } from "@/components/ui/checkbox"
import type { VLResult } from "@/lib/api/vl-client"
import { cn } from "@/lib/utils"

interface LogTableProps {
  data: VLResult[]
  onFilterClick?: (key: string, value: any) => void
}

export function LogTable({ data, onFilterClick }: LogTableProps) {
  const parentRef = useRef<HTMLDivElement>(null)

  // 1. 列宽状态 (保留一些默认值，其他动态计算)
  const [columnWidths, setColumnWidths] = useState<Record<string, number>>({
    _time: 220,
    _stream: 250,
    _msg: 600,
  })

  // 2. 当前显示的列 (初始化为空，由 useEffect 决定)
  const [visibleColumns, setVisibleColumns] = useState<string[]>([])

  // --- 核心修改: 智能列检测逻辑 ---
  useEffect(() => {
    if (!data || data.length === 0) return

    const firstRow = data[0]
    const keys = Object.keys(firstRow)
    const isRawLog = keys.includes("_time")

    if (isRawLog) {
      // 场景 A: 原始日志
      // 如果当前显示的列完全没有命中新数据的 key (说明之前可能是聚合视图)，则重置
      // 或者如果是第一次加载
      const hasValidColumns = visibleColumns.some(col => keys.includes(col))
      if (!hasValidColumns || visibleColumns.length === 0) {
        setVisibleColumns(["_time", "_stream", "_msg"])
      }
    } else {
      // 场景 B: 聚合查询 (如 stats by ...)
      // 聚合查询的字段通常是不固定的，所以每次结构变化都重置为所有字段
      // 为了防止无限循环，我们比较 key 是否一致
      const currentKeysStr = visibleColumns.sort().join(",")
      const newKeysStr = keys.sort().join(",")
      
      if (currentKeysStr !== newKeysStr) {
        // 保持原始顺序 (通常 LogSQL 返回的顺序是有意义的，如 Group Key 在前)
        setVisibleColumns(Object.keys(firstRow)) 
      }
    }
  }, [data]) // 依赖 data 变化自动触发

  // 3. 提取所有可用字段 (用于下拉菜单筛选)
  const allFields = useMemo(() => {
    if (!data || data.length === 0) return []
    const keys = new Set<string>()
    // 采样前 20 条数据提取 Key (避免遍历百万条数据)
    data.slice(0, 20).forEach(log => Object.keys(log).forEach(k => keys.add(k)))
    return Array.from(keys).sort()
  }, [data])

  // 4. 拖拽调整列宽逻辑
  const startResizing = (e: React.MouseEvent, col: string) => {
    e.preventDefault()
    const startX = e.pageX
    const startWidth = columnWidths[col] || 150

    const onMouseMove = (moveEvent: MouseEvent) => {
      const newWidth = Math.max(80, startWidth + (moveEvent.pageX - startX))
      setColumnWidths(prev => ({ ...prev, [col]: newWidth }))
    }

    const onMouseUp = () => {
      document.removeEventListener("mousemove", onMouseMove)
      document.removeEventListener("mouseup", onMouseUp)
      document.body.style.cursor = "default"
    }

    document.addEventListener("mousemove", onMouseMove)
    document.addEventListener("mouseup", onMouseUp)
    document.body.style.cursor = "col-resize"
  }

  // 5. 计算表格总宽度
  const totalTableWidth = useMemo(() => {
    return visibleColumns.reduce((acc, col) => acc + (columnWidths[col] || 150), 0)
  }, [visibleColumns, columnWidths])

  // 6. 虚拟滚动配置
  const rowVirtualizer = useVirtualizer({
    count: data.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 32, // 单行高度
    overscan: 10,
  })

  // 如果没有数据
  if (!data || data.length === 0) {
    return <div className="flex h-full items-center justify-center text-xs text-muted-foreground">No data available</div>
  }

  return (
    <div className="flex flex-col h-full bg-background border rounded-md overflow-hidden">
      {/* 顶部工具栏: 列配置 */}
      <div className="flex justify-end p-1 border-b bg-muted/20">
        <Popover>
          <PopoverTrigger asChild>
            <button className="flex items-center gap-1 text-[10px] text-muted-foreground hover:text-foreground px-2 py-1 rounded border hover:bg-muted/50 transition-colors">
              <Settings2 className="w-3 h-3" /> Columns
            </button>
          </PopoverTrigger>
          <PopoverContent className="w-56 p-2" align="end">
            <div className="space-y-2">
              <div className="text-[10px] font-bold text-muted-foreground border-b pb-1 uppercase">Display Fields</div>
              <div className="max-h-60 overflow-y-auto pr-1 space-y-1">
                {allFields.map(field => (
                  <div key={field} className="flex items-center gap-2 p-1 hover:bg-muted/50 rounded">
                    <Checkbox
                      id={`table-col-${field}`}
                      checked={visibleColumns.includes(field)}
                      onCheckedChange={(checked) => {
                        setVisibleColumns(prev => 
                          checked 
                            ? [...prev, field] 
                            : prev.filter(f => f !== field)
                        )
                      }}
                    />
                    <label htmlFor={`table-col-${field}`} className="text-xs font-mono cursor-pointer truncate flex-1" title={field}>
                      {field}
                    </label>
                  </div>
                ))}
              </div>
            </div>
          </PopoverContent>
        </Popover>
      </div>

      {/* 滚动容器 */}
      <div
        ref={parentRef}
        className="flex-1 overflow-auto scrollbar-thin scrollbar-thumb-muted-foreground/20"
      >
        <div style={{ width: totalTableWidth, minWidth: '100%' }}>
          {/* 表头 (Sticky) */}
          <div className="sticky top-0 z-20 bg-muted/90 backdrop-blur-sm border-b flex h-8 shadow-sm">
            {visibleColumns.map((col) => (
              <div
                key={col}
                className="relative px-3 flex items-center text-[10px] font-bold text-muted-foreground uppercase border-r group bg-muted/50"
                style={{ width: columnWidths[col] || 150, flexShrink: 0 }}
              >
                <span className="truncate" title={col}>{col}</span>
                {/* 拖拽手柄 */}
                <div
                  onMouseDown={(e) => startResizing(e, col)}
                  className="absolute right-0 top-0 bottom-0 w-1.5 cursor-col-resize hover:bg-primary/50 transition-colors z-30 flex items-center justify-center opacity-0 group-hover:opacity-100"
                >
                  <GripVertical className="w-2 h-2 text-primary" />
                </div>
              </div>
            ))}
          </div>

          {/* 表体 (Virtual) */}
          <div
            style={{
              height: `${rowVirtualizer.getTotalSize()}px`,
              position: 'relative',
            }}
          >
            {rowVirtualizer.getVirtualItems().map((virtualRow) => {
              const log = data[virtualRow.index]
              return (
                <div
                  key={virtualRow.key}
                  className="absolute left-0 w-full flex border-b hover:bg-primary/5 transition-colors group/row"
                  style={{
                    height: `${virtualRow.size}px`,
                    transform: `translateY(${virtualRow.start}px)`,
                  }}
                >
                  {visibleColumns.map((col) => (
                    <div
                      key={col}
                      className="px-3 flex items-center text-[10px] font-mono border-r last:border-0 truncate"
                      style={{ width: columnWidths[col] || 150, flexShrink: 0 }}
                    >
                      <div className="flex items-center justify-between w-full min-w-0 group/cell">
                        <span className="truncate" title={String(log[col] ?? "")}>
                          {String(log[col] ?? "")}
                        </span>
                        
                        {/* 单元格快捷筛选 (仅鼠标悬停显示) */}
                        {onFilterClick && (
                          <Filter
                            className="w-3 h-3 opacity-0 group-hover/cell:opacity-100 cursor-pointer text-primary shrink-0 hover:scale-110 ml-1 transition-all"
                            onClick={(e) => {
                              e.stopPropagation()
                              onFilterClick(col, log[col])
                            }}
                          />
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}