import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Search, Clock, ExternalLink, ChevronDown, ChevronRight } from "lucide-react";
import type { MergedEvent } from "./types";

interface TimelinePanelProps {
  mergedEvents: MergedEvent[];
  onOpenInLogs?: (event: MergedEvent) => void;
}

interface TimelineItemProps {
  event: MergedEvent;
  onOpenInLogs?: (event: MergedEvent) => void;
}

function TimelineItem({ event, onOpenInLogs }: TimelineItemProps) {
  const [expanded, setExpanded] = useState(false);
  
  // 安全解析时间 - 尝试多个可能的字段
  const timeFields = ['_time', 'time', 'timestamp', 'event_time', '@timestamp', 'datetime', 'ts'];
  let eventTime = '';
  for (const field of timeFields) {
    if (event[field]) {
      eventTime = String(event[field]);
      break;
    }
  }
  // 如果都没找到，用当前时间
  if (!eventTime) {
    eventTime = new Date().toISOString();
  }
  const timeObj = new Date(eventTime);
  const isValidTime = !isNaN(timeObj.getTime());
  
  // Debug: 可以通过 console.log(event) 查看实际数据结构
  
  const eventAction = event.activity_name || event.action || event.event_type;
  const severity = event.severity || "info";
  
  let actionColor = "text-muted-foreground";
  if (severity.toLowerCase() === "high" || severity.toLowerCase() === "critical") actionColor = "text-red-500 font-semibold";
  else if (severity.toLowerCase() === "medium") actionColor = "text-amber-500 font-semibold";

  // 摘要：只显示关键字段
  const summary = [
    event.src_ip && `src: ${event.src_ip}`,
    event.dst_ip && `dst: ${event.dst_ip}`,
    event.hostname && `host: ${event.hostname}`,
    event.username && `user: ${event.username}`,
    event.process_name && `process: ${event.process_name}`,
  ].filter(Boolean).join(" | ");

  const rawContent = event.raw_data || JSON.stringify(event, null, 2);
  const isLongContent = rawContent.length > 200;

  return (
    <div className="border-b border-border/50 hover:bg-muted/20 transition-colors">
      {/* 头部：时间 + 规则源 + 操作按钮 */}
      <div className="flex items-start gap-3 p-3">
        {/* 折叠按钮 */}
        <Button
          variant="ghost"
          size="sm"
          className="h-6 w-6 p-0 flex-shrink-0 mt-1"
          onClick={() => setExpanded(!expanded)}
        >
          {expanded ? (
            <ChevronDown className="w-4 h-4" />
          ) : (
            <ChevronRight className="w-4 h-4" />
          )}
        </Button>

        {/* 时间 */}
        <div className="w-[140px] flex-shrink-0 font-mono text-xs text-muted-foreground">
          {isValidTime ? timeObj.toLocaleString("en-GB", {
            year: "numeric",
            month: "2-digit",
            day: "2-digit",
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
            hour12: false,
          }) : "N/A"}
        </div>

        {/* 规则源 */}
        <Badge
          variant="outline"
          className="bg-background border-primary/20 text-primary text-[10px] font-medium whitespace-nowrap overflow-hidden text-ellipsis max-w-[120px] flex-shrink-0"
          title={event._source_template}
        >
          {event._source_template}
        </Badge>

        {/* 操作：事件类型 + 摘要 */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            {eventAction && (
              <span className={actionColor}>[{eventAction}]</span>
            )}
          </div>
          {!expanded && summary && (
            <div className="text-xs text-muted-foreground truncate mt-0.5">
              {summary}
            </div>
          )}
        </div>

        {/* 跳转按钮 */}
        {onOpenInLogs && (
          <Button
            variant="ghost"
            size="sm"
            className="h-6 px-2 text-[10px] flex-shrink-0"
            onClick={() => onOpenInLogs(event)}
          >
            <ExternalLink className="w-3 h-3 mr-1" />
            Query
          </Button>
        )}
      </div>

      {/* 展开内容 */}
      {expanded && (
        <div className="pl-12 pr-3 pb-3">
          {summary && (
            <div className="text-xs text-muted-foreground mb-2 bg-muted/30 p-2 rounded">
              {summary}
            </div>
          )}
          <pre className="text-[11px] font-mono text-muted-foreground bg-muted p-3 rounded overflow-x-auto max-h-60">
            {rawContent}
          </pre>
        </div>
      )}
    </div>
  );
}

export function TimelinePanel({ mergedEvents, onOpenInLogs }: TimelinePanelProps) {
  // 按时间排序
  const sortedEvents = [...mergedEvents].sort((a, b) => {
    const timeA = new Date(a._time || a.time || a.timestamp || 0).getTime();
    const timeB = new Date(b._time || b.time || b.timestamp || 0).getTime();
    return timeB - timeA;
  });

  return (
    <Card className="flex-1 shadow-sm flex flex-col overflow-hidden border-t-4 border-t-primary/20">
      <CardHeader className="pb-2 bg-muted/10 border-b flex-none py-3">
        <div className="flex justify-between items-center">
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <Clock className="w-4 h-4 text-muted-foreground" />
            Unified Event Timeline
          </CardTitle>
          {mergedEvents.length > 0 && (
            <Badge variant="outline" className="text-[10px] font-mono">
              Total: {mergedEvents.length}
            </Badge>
          )}
        </div>
      </CardHeader>
      <CardContent className="flex-1 p-0 overflow-hidden relative">
        {mergedEvents.length > 0 ? (
          <ScrollArea className="h-full w-full">
            <div className="divide-y divide-border/50">
              {sortedEvents.map((ev, idx) => (
                <TimelineItem key={idx} event={ev} onOpenInLogs={onOpenInLogs} />
              ))}
            </div>
          </ScrollArea>
        ) : (
          <div className="absolute inset-0 flex flex-col items-center justify-center text-muted-foreground/60 space-y-4">
            <div className="p-4 rounded-full bg-muted/10 ring-1 ring-border/50">
              <Search className="w-8 h-8 opacity-50" />
            </div>
            <div className="text-center">
              <p className="text-sm font-medium text-muted-foreground">Canvas is empty</p>
              <p className="text-xs mt-1 max-w-[250px] mx-auto">
                Select directives above and click execute to build the timeline correlation.
              </p>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
