import React, { useState, useMemo } from "react"
import { ArrowDown01, Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"

// 你自定义的配置
const LIMIT_OPTIONS = [
  { label: "100", value: "100" },
  { label: "1,000", value: "1000" }, // 默认推荐
  { label: "2,000", value: "2000" },
  { label: "5,000", value: "5000" },
  { label: "10,000", value: "10000" }, // 新增的万级数据
]

interface LimitSelectorProps {
  value: string
  onChange: (value: string) => void
}

export function LimitSelector({ value, onChange }: LimitSelectorProps) {
  const [open, setOpen] = useState(false)

  const handleSelect = (val: string) => {
    onChange(val)
    setOpen(false)
  }

  // ✅ 核心修复：优先查找对应的 Label 用于显示
  // 这样选中 "1000" 时，按钮上会显示 "1,000" (带逗号)
  const displayLabel = useMemo(() => {
    const selected = LIMIT_OPTIONS.find(opt => opt.value === value)
    return selected ? selected.label : value
  }, [value])

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button 
          variant="outline" 
          size="sm" 
          className={cn(
            // ✅ 调整宽度：min-w-[80px] 以容纳 "10,000 rows"
            "h-7 px-2 font-normal justify-start gap-2 bg-background border-border/60 hover:bg-accent transition-all min-w-[85px]",
            open && "border-primary/50 ring-1 ring-primary/20"
          )}
        >
          <ArrowDown01 className="h-3.5 w-3.5 text-orange-500/80 shrink-0" />
          
          {/* 显示区域 */}
          <div className="flex items-baseline gap-1 text-[11px] font-mono font-medium">
             <span>{displayLabel}</span>
             <span className="text-muted-foreground/50 scale-90">rows</span>
          </div>
        </Button>
      </PopoverTrigger>
      
      <PopoverContent className="w-[140px] p-1" align="start">
        <div className="flex flex-col gap-0.5">
          <div className="px-2 py-1.5 text-[9px] font-bold text-muted-foreground uppercase tracking-wider opacity-70">
            Max Results
          </div>
          {LIMIT_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              onClick={() => handleSelect(opt.value)}
              className={cn(
                "w-full text-left px-2 py-1.5 text-[11px] font-mono rounded-sm transition-colors flex items-center justify-between group",
                value === opt.value 
                  ? "bg-primary/10 text-primary font-bold" 
                  : "hover:bg-muted text-foreground/80"
              )}
            >
              {/* 下拉列表中也显示带格式的 Label */}
              <span>{opt.label}</span>
              {value === opt.value && <Check className="w-3 h-3 opacity-100" />}
            </button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}