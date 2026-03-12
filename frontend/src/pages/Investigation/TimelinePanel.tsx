import { useState, useMemo } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Search, Clock, ExternalLink, ChevronDown, ChevronUp, Copy, Check, X } from "lucide-react";
import type { MergedEvent } from "./types";
import { toast } from "sonner";
import { useTabStore } from "@/stores/tab-store";

interface TimelinePanelProps {
  mergedEvents: MergedEvent[];
}

// Time轴样式Group件 - function TimelinePanel({ mergedEvents }: TimelinePanelProps) {
  const [expandedEvents, setExpandedEvents] = useState<Set<number>>(new Set());
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  
  const { addTab } = useTabStore();

  // FilterEvent - filteredEvents = useMemo(() => {
    if (!searchQuery.trim()) return mergedEvents;
    const query = searchQuery.toLowerCase();
    return mergedEvents.filter(ev => {
      const jsonStr = JSON.stringify(ev).toLowerCase();
      return jsonStr.includes(query);
    });
  }, [mergedEvents, searchQuery]);

  const toggleExpand = (index: number) => {
    setExpandedEvents(prev => {
      const next = new Set(prev);
      if (next.has(index)) {
        next.delete(index);
      } else {
        next.add(index);
      }
      return next;
    });
  };

  // 复制Event - 到剪贴板
  const copyEvent = (event: MergedEvent, index: number) => {
    const jsonStr = JSON.stringify(event, null, 2);
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(jsonStr).then(() => {
        setCopiedIndex(index);
        toast.success("Copied to clipboard");
        setTimeout(() => setCopiedIndex(null), 2000);
      }).catch(() => {
        toast.error("Failed to copy");
      });
    } else {
      // Degrade方案 - textarea = document.createElement('textarea');
      textarea.value = jsonStr;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      setCopiedIndex(index);
      toast.success("Copied to clipboard");
      setTimeout(() => setCopiedIndex(null), 2000);
    }
  };

  //   FormatTime - 数字格式，带年份
  const formatTime = (timeStr: any) => {
    if (!timeStr) return "Unknown";
    const date = new Date(timeStr);
    if (isNaN(date.getTime())) return String(timeStr);
    return date.toLocaleString('en-CA', { 
      year: 'numeric', month: '2-digit', day: '2-digit', 
      hour: '2-digit', minute: '2-digit', second: '2-digit',
      hour12: false
    });
  };

  // 跳转到 - Query (使用 tab-store)
  const openInLogsQuery = (event: MergedEvent) => {
    let query = "";
    
    // 优先使用Rule的Query语句 - (event._rule_query) {
      query = event._rule_query;
    } else {
      //   如果没有RuleQuery语句，则使用原有的逻辑构建Query
      // 尝试GetTime - timeStr = event._time;
      if (timeStr) {
        const eventTime = new Date(timeStr);
        if (!isNaN(eventTime.getTime())) {
          //   ±5 分钟Time窗口
          const start = new Date(eventTime.getTime() - 5 * 60 * 1000).toISOString();
          const end = new Date(eventTime.getTime() + 5 * 60 * 1000).toISOString();
          query = `_time:[${start}, ${end}]`;
        }
      }
      
      // 尝试从EventMedium提取关键字段构建Query - (event["observer.hostname"]) {
        query += ` | observer.hostname:"${event["observer.hostname"]}"`;
      } else if (event.hostname) {
        query += ` | hostname:"${event.hostname}"`;
      }
      
      if (event["src_endpoint.ip"]) {
        query += ` | src_endpoint.ip:"${event["src_endpoint.ip"]}"`;
      } else if (event.src_ip) {
        query += ` | src_ip:"${event.src_ip}"`;
      }
    }

    if (!query) {
      query = "*"; //   没有有效字段则Query所有
    }

    // 使用 - -store 跳转
    addTab('logs', 'Logs Query', { q: query });
    toast.success("Opened in Logs Query");
  };

  return (
    <Card className="flex-1 shadow-sm flex flex-col overflow-hidden border-t-4 border-t-primary/20">
      <CardHeader className="pb-2 bg-muted/10 border-b flex-none py-3">
        <div className="flex justify-between items-center gap-4">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <Clock className="w-4 h-4 text-muted-foreground" />
            Unified Event Timeline
          </CardTitle>
          
          {/* Search框 */}
          <div className="flex items-center gap-2">
            <div className="relative">
              <Search className="w-3 h-3 absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <Input 
                placeholder="Search events..." 
                className="h-7 text-xs pl-7 w-[200px]"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
              {searchQuery && (
                <button 
                  className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  onClick={() => setSearchQuery("")}
                >
                  <X className="w-3 h-3" />
                </button>
              )}
            </div>
            {mergedEvents.length > 0 && (
              <Badge variant="outline" className="text-[10px] font-mono">
                {searchQuery ? `${filteredEvents.length} / ${mergedEvents.length}` : mergedEvents.length}
              </Badge>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex-1 p-0 overflow-hidden relative">
        <ScrollArea className="h-full w-full">
          {filteredEvents.length > 0 ? (
            <div className="py-4 px-6">
              {/* Time线 */}
              <div className="relative">
                {/* 垂直线 */}
                <div className="absolute left-4 top-0 bottom-0 w-0.5 bg-gradient-to-b from-primary/30 via-primary/10 to-transparent" />
                
                {filteredEvents.map((ev, idx) => {
                  //   找到原始索引，用于正确Handle展开Status
                  const originalIdx = mergedEvents.indexOf(ev);
                  const eventAction = ev.activity_name || ev.action || ev.event_type;
                  const severity = ev.severity || "info";
                  const isExpanded = expandedEvents.has(originalIdx);
                  
                  // 颜色映射 - severityColor = "bg-blue-500";
                  if (severity.toLowerCase() === "critical" || severity.toLowerCase() === "high") {
                    severityColor = "bg-red-500";
                  } else if (severity.toLowerCase() === "medium") {
                    severityColor = "bg-amber-500";
                  } else if (severity.toLowerCase() === "low") {
                    severityColor = "bg-green-500";
                  }

                  return (
                    <div key={originalIdx} className="relative flex gap-4 pb-6 last:pb-0 group">
                      {/* Time点圆圈 */}
                      <div className={`relative z-10 w-2 h-2 rounded-full mt-2 shrink-0 ${severityColor} ring-4 ring-background group-hover:scale-125 transition-transform`} />
                      
                      {/* Event卡片 */}
                      <div className="flex-1 min-w-0">
                        {/* Time戳 + 来源 */}
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-xs font-mono text-muted-foreground whitespace-nowrap">
                            {formatTime(ev._time)}
                          </span>
                          <Badge variant="outline" className="bg-background border-primary/20 text-primary text-[10px] font-medium">
                            {ev._source_template}
                          </Badge>
                          {eventAction && (
                            <span className={`text-xs font-medium ${
                              severity.toLowerCase() === "high" || severity.toLowerCase() === "critical" 
                                ? "text-red-500" 
                                : severity.toLowerCase() === "medium"
                                ? "text-amber-500"
                                : "text-muted-foreground"
                            }`}>
                              [{eventAction}]
                            </span>
                          )}
                        </div>

                        {/* Event内容 */}
                        <div className="bg-card border rounded-md overflow-hidden shadow-sm hover:border-primary/20 transition-colors">
                          {/* 内容头部 - 可点击展开 */}
                          <div 
                            className="p-2 px-3 flex items-center justify-between cursor-pointer bg-muted/20 hover:bg-muted/30"
                            onClick={() => toggleExpand(originalIdx)}
                          >
                            <div className="flex items-center gap-2 overflow-hidden">
                              {isExpanded ? (
                                <ChevronUp className="w-4 h-4 text-muted-foreground shrink-0" />
                              ) : (
                                <ChevronDown className="w-4 h-4 text-muted-foreground shrink-0" />
                              )}
                              <span className="text-xs font-mono text-muted-foreground truncate">
                                {ev.raw_data || JSON.stringify(ev).substring(0, 100) + "..."}
                              </span>
                            </div>
                            <div className="flex items-center gap-1 shrink-0">
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 px-2 text-[10px]"
                                onClick={(e) => { e.stopPropagation(); copyEvent(ev, originalIdx); }}
                              >
                                {copiedIndex === originalIdx ? <Check className="w-3 h-3 text-green-500" /> : <Copy className="w-3 h-3" />}
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 px-2 text-[10px] text-blue-600 hover:text-blue-700"
                                onClick={(e) => { e.stopPropagation(); openInLogsQuery(ev); }}
                              >
                                <ExternalLink className="w-3 h-3 mr-1" />
                                Logs
                              </Button>
                            </div>
                          </div>

                          {/* 展开的Detail */}
                          {isExpanded && (
                            <div className="border-t bg-muted/5 p-3">
                              <pre className="text-xs font-mono text-muted-foreground whitespace-pre-wrap break-all max-h-64 overflow-auto">
                                {JSON.stringify(ev, null, 2)}
                              </pre>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          ) : (
            <div className="absolute inset-0 flex flex-col items-center justify-center text-muted-foreground/60 space-y-4">
              <div className="p-4 rounded-full bg-muted/10 ring-1 ring-border/50">
                <Search className="w-8 h-8 opacity-50" />
              </div>
              <div className="text-center">
                {searchQuery ? (
                  <>
                    <p className="text-sm font-medium text-muted-foreground">No matching events</p>
                    <p className="text-xs mt-1 max-w-[250px] mx-auto">
                      Try a different search term.
                    </p>
                  </>
                ) : (
                  <>
                    <p className="text-sm font-medium text-muted-foreground">Canvas is empty</p>
                    <p className="text-xs mt-1 max-w-[250px] mx-auto">
                      Select directives above and click execute to build the timeline correlation.
                    </p>
                  </>
                )}
              </div>
            </div>
          )}
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
