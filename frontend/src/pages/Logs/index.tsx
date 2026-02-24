import React, { useState, useEffect, useCallback } from "react";
import { useSearchParams } from "react-router-dom";
import {
  Play, RotateCw, LayoutGrid, Table as TableIcon, Code2, Share2, Loader2
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";

import { runVLQuery, runVLHits } from "@/lib/api/vl-client";
import type { VLResult } from "@/lib/api/vl-client";
import { LogTable } from "@/components/logs/LogTableView";
import { LogGroupView } from "@/components/logs/LogGroupView";
import { LogJsonView } from "@/components/logs/LogJsonView";
import { LogHitsChart } from "@/components/logs/LogHitsChart";
import { LogSQLEditor } from "@/components/editor/LogSQLEditor";
import { TimeRangePicker } from "@/components/logs/TimeRangePicker";
import { LimitSelector } from "@/components/logs/LimitSelector";
import { toast } from "sonner";

// ... calcTimeRange 保持不变 ...
const calcTimeRange = (rangeStr: string) => {
  if (!rangeStr || rangeStr === "all") return { start: undefined, end: undefined };
  const now = new Date();
  const start = new Date();
  const value = parseInt(rangeStr);
  if (rangeStr.endsWith("m")) start.setMinutes(now.getMinutes() - value);
  else if (rangeStr.endsWith("h")) start.setHours(now.getHours() - value);
  else if (rangeStr.endsWith("d")) start.setDate(now.getDate() - value);
  else return { start: undefined, end: undefined };
  return { start: start.toISOString(), end: now.toISOString() };
};

export default function LogsPage() {
  const [searchParams, setSearchParams] = useSearchParams();

  const [query, setQuery] = useState("*");
  const [limit, setLimit] = useState("1000");
  const [timeRange, setTimeRange] = useState("5m");
  const [logs, setLogs] = useState<VLResult[]>([]);
  const [totalHits, setTotalHits] = useState<number>(0);
  const [loading, setLoading] = useState(false);
  const [delay, setDelay] = useState(0);
  const [viewMode, setViewMode] = useState<"group" | "table" | "json">("table");
  const [editorHeight, setEditorHeight] = useState(120);

  // 初始化逻辑 (保持不变)
  useEffect(() => {
    const q = searchParams.get("q");
    const t = searchParams.get("t");
    const l = searchParams.get("l");
    if (q) setQuery(decodeURIComponent(q));
    if (t) setTimeRange(t);
    if (l) setLimit(l);
    if (q) setTimeout(() => handlerRun(decodeURIComponent(q), t || "5m", l || "1000"), 100);
  }, []);

  const handlerRun = useCallback(async (customQuery?: string, customTime?: string, customLimit?: string) => {
    const q = customQuery ?? query;
    const t = customTime ?? timeRange;
    const l = customLimit ?? limit;

    setLoading(true);
    try {
      const { start, end } = calcTimeRange(t);
      // 并行请求：数据 + 总数
      const [data, hits] = await Promise.all([
        runVLQuery(q, l, start, end),
        runVLHits(q, start, end)
      ]);
      setLogs(data);
      setTotalHits(hits);
      setSearchParams({ q: encodeURIComponent(q), t: t, l: l }, { replace: true });
    } catch (error: any) {
      toast.error("Query Failed", { description: error.message });
    } finally {
      setLoading(false);
    }
  }, [query, timeRange, limit, setSearchParams]);

  // 自动刷新逻辑
  useEffect(() => {
    if (delay > 0) {
      const timer = setInterval(() => handlerRun(), delay * 1000);
      return () => clearInterval(timer);
    }
  }, [delay, handlerRun]);

  // 拖拽逻辑
  const startResizing = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    const startY = e.clientY;
    const startHeight = editorHeight;
    const onMouseMove = (ev: MouseEvent) => setEditorHeight(Math.max(60, Math.min(startHeight + (ev.clientY - startY), 600)));
    const onMouseUp = () => {
      document.removeEventListener("mousemove", onMouseMove);
      document.removeEventListener("mouseup", onMouseUp);
    };
    document.addEventListener("mousemove", onMouseMove);
    document.addEventListener("mouseup", onMouseUp);
  }, [editorHeight]);

  const copyShareLink = () => {
    navigator.clipboard.writeText(window.location.href);
    toast.success("Link copied");
  };

  return (
    // 使用 Tabs 包裹整个页面以共享 viewMode 状态，但不使用 TabsContent 这种分割布局
    <Tabs 
      value={viewMode} 
      onValueChange={(v) => setViewMode(v as any)} 
      className="h-full flex flex-col gap-0 bg-background font-sans overflow-hidden"
    >
      
      {/* --- A. 顶部工具栏 (Header) --- */}
      <div className="flex items-center px-3 py-1.5 border-b bg-muted/10 flex-none gap-4 h-[46px]">
        {/* 左侧控制区 */}
        <div className="flex items-center gap-2">
          <Button onClick={() => handlerRun()} disabled={loading} size="sm" className="bg-emerald-600 hover:bg-emerald-700 h-8 px-4 shadow-sm">
            {loading ? <RotateCw className="w-3.5 h-3.5 animate-spin" /> : <Play className="w-3.5 h-3.5" />}
            <span className="ml-2 text-xs font-bold">Execute</span>
          </Button>

          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={copyShareLink} title="Share">
            <Share2 className="w-4 h-4" />
          </Button>

          <div className="h-5 w-[1px] bg-border mx-1" />

          <TimeRangePicker value={timeRange} onChange={setTimeRange} />
          <LimitSelector value={limit} onChange={setLimit} />

          <div className="h-5 w-[1px] bg-border mx-1" />

          <div className="flex items-center gap-2 bg-background border rounded px-2 h-8">
            <span className="text-[10px] text-muted-foreground font-medium uppercase">AUTO:</span>
            <select className="text-xs bg-transparent outline-none font-bold w-12 cursor-pointer" value={delay} onChange={(e) => setDelay(Number(e.target.value))}>
              <option value="0">Off</option>
              <option value="5">5s</option>
              <option value="10">10s</option>
            </select>
          </div>
        </div>

        {/* --- 右侧：视图切换 (完美还原截图布局) --- */}
        <div className="ml-auto flex items-center gap-2">
           {/* 如果是聚合查询(无图表)，在这里显示 Hits 也不错，或者就隐藏 */}
           {(!logs.length || !logs[0]._time) && totalHits > 0 && (
             <Badge variant="outline" className="h-7 bg-background font-mono">
               HITS: {totalHits.toLocaleString()}
             </Badge>
           )}

           <TabsList className="h-8 bg-muted/50 p-0.5">
            <TabsTrigger value="group" className="text-[11px] px-3 h-7 data-[state=active]:bg-background data-[state=active]:shadow-sm">
              <LayoutGrid className="w-3.5 h-3.5 mr-1.5" /> Group
            </TabsTrigger>
            <TabsTrigger value="table" className="text-[11px] px-3 h-7 data-[state=active]:bg-background data-[state=active]:shadow-sm">
              <TableIcon className="w-3.5 h-3.5 mr-1.5" /> Table
            </TabsTrigger>
            <TabsTrigger value="json" className="text-[11px] px-3 h-7 data-[state=active]:bg-background data-[state=active]:shadow-sm">
              <Code2 className="w-3.5 h-3.5 mr-1.5" /> JSON
            </TabsTrigger>
          </TabsList>
        </div>
      </div>

      {/* --- B. SQL 编辑器 --- */}
      <div className="flex-none border-b relative group z-10 flex flex-col bg-[#1e1e1e]" style={{ height: editorHeight }}>
        <div className="flex-1 min-h-0">
          <LogSQLEditor value={query} onChange={(v) => setQuery(v || "")} onRun={() => handlerRun()} />
        </div>
        {/* 拖拽条 */}
        <div onMouseDown={startResizing} className="absolute bottom-0 left-0 right-0 h-1.5 cursor-row-resize hover:bg-primary/20 flex justify-center items-center z-50">
           <div className="w-10 h-0.5 bg-white/10 rounded-full" />
        </div>
      </div>

      {/* --- C. 图表区域 (自动折叠) --- */}
      <div 
        className="flex-none border-b bg-card relative overflow-hidden transition-all duration-300" 
        // 如果第一条数据没有 _time (聚合查询)，则高度设为 0 隐藏
        style={{ height: logs.length > 0 && logs[0]._time ? 180 : 0 }}
      > 
         <LogHitsChart data={logs} />
      </div>

      {/* --- D. 结果展示 --- */}
      <div className="flex-1 overflow-hidden relative bg-card">
        {viewMode === "table" && <LogTable data={logs} onFilterClick={(k, v) => setQuery(prev => `${prev} AND ${k}:"${v}"`)} />}
        {viewMode === "group" && <LogGroupView data={logs} />}
        {viewMode === "json" && <LogJsonView data={logs} />}
      </div>
      
    </Tabs>
  );
}