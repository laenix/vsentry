import React, { useState } from "react"
import { format, subMinutes, subHours, subDays } from "date-fns"
import { Calendar as CalendarIcon, Clock, ArrowRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"

const QUICK_RANGES = [
  { label: "Last 5 minutes", value: "5m" },
  { label: "Last 30 minutes", value: "30m" },
  { label: "Last 1 hour", value: "1h" },
  { label: "Last 12 hours", value: "12h" },
  { label: "Last 24 hours", value: "24h" },
  { label: "Last 2 days", value: "2d" },
  { label: "Last 7 days", value: "7d" },
  { label: "Last 30 days", value: "30d" },
  { label: "Last 1 year", value: "1y" },
]
export function TimeRangePicker({ value, onChange }: { value: string, onChange: (v: string) => void }) {
  const [open, setOpen] = useState(false) // 控制弹窗开关
  const [startDate, setStartDate] = useState<Date | undefined>(subMinutes(new Date(), 5))
  const [endDate, setEndDate] = useState<Date | undefined>(new Date())

  const currentLabel = QUICK_RANGES.find(r => r.value === value)?.label || value

  // 处理快捷选择
  const handleQuickSelect = (val: string) => {
    onChange(val)
    setOpen(false) // 点击后自动关闭
  }

  // 处理绝对时间应用
  const handleApplyCustom = () => {
  if (startDate && endDate) {
    // VictoriaLogs 识别的语法是 _time:[YYYY-MM-DDTHH:MM:SSZ, YYYY-MM-DDTHH:MM:SSZ]
    // 使用 formatISO 确保生成 2026-02-11T18:00:00Z 这种标准格式
    const startISO = startDate.toISOString()
    const endISO = endDate.toISOString()
    
    // 生成 LogsQL 语法
    const customValue = `_time:[${startISO}, ${endISO}]`
    
    onChange(customValue) // 将这个值传给 LogsPage 的 timeRange 状态
    setOpen(false)
  }
}

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button 
          variant="outline" 
          size="sm" 
          className="h-7 px-2 font-normal justify-start gap-2 bg-background border-border/60 hover:bg-accent transition-all"
        >
          <Clock className="h-3 w-3 text-emerald-500" />
          <span className="text-[11px] font-mono font-medium">{currentLabel}</span>
        </Button>
      </PopoverTrigger>
      
      <PopoverContent className="w-[520px] p-0 flex flex-row divide-x shadow-xl border-border/40" align="start">
        {/* 左侧：快捷选择 - 点击即走 */}
        <div className="w-[180px] p-1.5 flex flex-col bg-muted/10">
          <div className="px-2 py-1.5 text-[10px] font-bold text-muted-foreground uppercase tracking-tighter opacity-60">
            Quick Ranges
          </div>
          <div className="flex-1 overflow-y-auto pr-1">
            {QUICK_RANGES.map((range) => (
              <button
                key={range.value}
                onClick={() => handleQuickSelect(range.value)}
                className={cn(
                  "w-full text-left px-2 py-1.5 text-[11px] rounded transition-all mb-0.5",
                  value === range.value 
                    ? "bg-primary text-primary-foreground font-bold" 
                    : "hover:bg-primary/10 text-foreground/80"
                )}
              >
                {range.label}
              </button>
            ))}
          </div>
        </div>

        {/* 右侧：绝对时间选择 - 仿 vmui 布局 */}
        <div className="flex-1 p-3 flex flex-col gap-3 bg-background">
          <div className="text-[10px] font-bold text-muted-foreground uppercase tracking-tighter opacity-60">
            Absolute Time Range
          </div>
          
          <div className="grid grid-cols-2 gap-2">
            <div className="space-y-1">
              <label className="text-[9px] font-bold text-muted-foreground px-1">FROM</label>
              <div className="flex items-center gap-2 border rounded p-1.5 bg-muted/20 text-[10px] font-mono">
                 {startDate ? format(startDate, "yyyy-MM-dd HH:mm") : "Select date"}
              </div>
            </div>
            <div className="space-y-1">
              <label className="text-[9px] font-bold text-muted-foreground px-1">TO</label>
              <div className="flex items-center gap-2 border rounded p-1.5 bg-muted/20 text-[10px] font-mono">
                 {endDate ? format(endDate, "yyyy-MM-dd HH:mm") : "Now"}
              </div>
            </div>
          </div>

          <div className="border rounded-md p-1 bg-card">
            <Calendar
              mode="range"
              selected={{ from: startDate, to: endDate }}
              onSelect={(range) => {
                setStartDate(range?.from)
                setEndDate(range?.to)
              }}
              numberOfMonths={1}
              className="p-0"
              classNames={{
                day_selected: "bg-primary text-primary-foreground hover:bg-primary hover:text-primary-foreground focus:bg-primary focus:text-primary-foreground",
                day_today: "bg-accent text-accent-foreground",
              }}
            />
          </div>

          <div className="flex justify-end gap-2 mt-auto pt-2 border-t">
            <Button 
              variant="ghost" 
              size="sm" 
              className="h-7 text-[10px] font-bold uppercase"
              onClick={() => setOpen(false)}
            >
              Cancel
            </Button>
            <Button 
              size="sm" 
              className="h-7 text-[10px] font-bold uppercase bg-emerald-600 hover:bg-emerald-700"
              onClick={handleApplyCustom}
            >
              Apply Range
            </Button>
          </div>
        </div>
      </PopoverContent>
    </Popover>
  )
}