import React, { useMemo } from "react";
import type { VLResult } from "@/lib/api/vl-client";
import { cn } from "@/lib/utils";

interface LogHitsChartProps {
  data: VLResult[]
}

export function LogHitsChart({ data }: LogHitsChartProps) {
  const { buckets, streams, maxTotal, startTime, endTime } = useMemo(() => {
    // ✅ 修复崩溃核心：
    // 1. 如果没有数据，返回空
    // 2. 如果第一条数据没有 _time 字段（说明是 stats/uniq 等聚合查询结果），无法画时间分布图，直接返回空
    if (!data || data.length === 0 || !data[0]._time) {
      return { buckets: [], streams: [], maxTotal: 0, startTime: "", endTime: "" };
    }

    // 过滤掉时间无效的脏数据，防止计算 NaN
    const validData = data.filter(d => d._time && !isNaN(new Date(d._time).getTime()));
    
    if (validData.length === 0) {
      return { buckets: [], streams: [], maxTotal: 0, startTime: "", endTime: "" };
    }

    // --- 以下逻辑使用 validData 而不是 data ---

    const timestamps = validData.map(d => new Date(d._time).getTime()).sort((a, b) => a - b)
    const min = timestamps[0]
    const max = timestamps[timestamps.length - 1]
    const range = max - min || 1
    const bucketCount = 80 // 对标 vmui 的精细度

    // 1. 提取所有流并分配颜色 (对标 vmui)
    const streamSet = new Set<string>()
    validData.forEach(d => streamSet.add(d._stream || "{}"))
    const streams = Array.from(streamSet).map((s, i) => ({
      name: s,
      color: `hsl(${(i * 137.5) % 360}, 60%, 65%)` // 自动生成互补色
    }))

    // 2. 初始化桶
    const buckets = Array.from({ length: bucketCount }).map(() => ({
      counts: {} as Record<string, number>,
      total: 0
    }))

    // 3. 填充数据 (堆叠逻辑)
    validData.forEach(d => {
      const ts = new Date(d._time).getTime()
      // 确保 bIdx 在 0 ~ bucketCount-1 范围内
      const bIdx = Math.min(Math.floor(((ts - min) / range) * bucketCount), bucketCount - 1)
      
      // 双重保险：虽然上面过滤了，但防止极端浮点数误差导致 bIdx 越界
      if (buckets[bIdx]) {
        const sName = d._stream || "{}"
        buckets[bIdx].counts[sName] = (buckets[bIdx].counts[sName] || 0) + 1
        buckets[bIdx].total++
      }
    })

    return {
      buckets,
      streams,
      maxTotal: Math.max(...buckets.map(b => b.total), 1),
      startTime: new Date(min).toLocaleString(),
      endTime: new Date(max).toLocaleString()
    }
  }, [data])

  // 如果没有有效的分桶数据，不渲染任何内容 (比如聚合查询时)
  if (buckets.length === 0) return null

  return (
    <div className="flex flex-col h-full w-full bg-background/50 p-3 pt-6 relative group">
      {/* 1. 顶部指标栏 (对标 vmui) */}
      <div className="absolute top-1 left-3 flex gap-4 text-[9px] text-muted-foreground font-mono">
        <span>Total: {data.length} hits</span>
        <span>Streams: {streams.length}</span>
      </div>

      <div className="flex-1 flex items-end gap-[1px] relative border-l border-b border-border/60">
        {/* 2. 背景网格线 */}
        <div className="absolute inset-0 flex flex-col justify-between pointer-events-none">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="w-full border-t border-border/20 border-dashed h-0" />
          ))}
        </div>

        {/* 3. 纵坐标刻度 */}
        <div className="absolute -left-7 top-0 bottom-0 flex flex-col justify-between text-[8px] text-muted-foreground/40 font-mono">
          <span>{Math.ceil(maxTotal)}</span>
          <span>{Math.ceil(maxTotal / 2)}</span>
          <span>0</span>
        </div>

        {/* 4. 堆叠柱状图主体 */}
        {buckets.map((b, i) => (
          <div key={i} className="flex-1 h-full flex flex-col justify-end group/bar relative">
            {streams.map(s => (
              <div
                key={s.name}
                className="w-full transition-all duration-300 hover:brightness-110"
                style={{
                  height: `${((b.counts[s.name] || 0) / maxTotal) * 100}%`,
                  backgroundColor: s.color
                }}
              />
            ))}
            
            {/* Tooltip */}
            {b.total > 0 && (
              <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover/bar:block z-50 bg-popover/95 border p-2 rounded shadow-xl text-[9px] font-mono min-w-[150px]">
                <div className="border-b mb-1 pb-1 font-bold">Hits: {b.total}</div>
                {streams.filter(s => b.counts[s.name]).map(s => (
                  <div key={s.name} className="flex justify-between gap-4">
                    <span className="truncate opacity-70">{s.name}</span>
                    <span style={{ color: s.color }}>{b.counts[s.name]}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>

      {/* 5. 横坐标时间 */}
      <div className="flex justify-between mt-1 text-[8px] text-muted-foreground/60 font-mono">
        <span>{startTime}</span>
        <span>Timeline (Distribution by _stream)</span>
        <span>{endTime}</span>
      </div>

      {/* 6. 图例区域 (对标 vmui 底部图例) */}
      <div className="mt-3 flex flex-wrap gap-x-4 gap-y-1 border-t pt-2">
        {streams.map(s => {
          // 使用原始 data 计算总数，而不是 validData，保证图例数字匹配表格
          const count = data.filter(d => (d._stream || "{}") === s.name).length
          return (
            <div key={s.name} className="flex items-center gap-1.5 no-shrink">
              <div className="w-2 h-2 rounded-sm" style={{ backgroundColor: s.color }} />
              <span className="text-[9px] font-mono text-muted-foreground truncate max-w-[200px]" title={s.name}>
                {s.name}
              </span>
              <span className="text-[9px] font-bold opacity-40">({count})</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}