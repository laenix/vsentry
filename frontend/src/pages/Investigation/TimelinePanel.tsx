import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Search, Clock } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { MergedEvent } from "./types";

interface TimelinePanelProps {
  mergedEvents: MergedEvent[];
}

export function TimelinePanel({ mergedEvents }: TimelinePanelProps) {
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
              Total Events: {mergedEvents.length}
            </Badge>
          )}
        </div>
      </CardHeader>
      <CardContent className="flex-1 p-0 overflow-hidden relative">
        <ScrollArea className="h-full w-full">
          {mergedEvents.length > 0 ? (
            <Table>
              <TableHeader className="sticky top-0 bg-background/95 backdrop-blur z-10 border-b shadow-sm">
                <TableRow>
                  <TableHead className="w-[180px] font-semibold">Timestamp</TableHead>
                  <TableHead className="w-[160px] font-semibold">Directive Source</TableHead>
                  <TableHead className="font-semibold">Event Details</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {mergedEvents.map((ev, idx) => {
                  const eventAction = ev.activity_name || ev.action || ev.event_type;
                  const severity = ev.severity || "info";
                  
                  let actionColor = "text-muted-foreground";
                  if (severity.toLowerCase() === "high" || severity.toLowerCase() === "critical") actionColor = "text-red-500 font-semibold";
                  else if (severity.toLowerCase() === "medium") actionColor = "text-amber-500 font-semibold";

                  return (
                    <TableRow key={idx} className="hover:bg-muted/30 transition-colors group">
                      <TableCell className="font-mono text-xs whitespace-nowrap text-muted-foreground align-top pt-3">
                        {new Date(ev._time).toLocaleString(undefined, { 
                          month: 'short', day: '2-digit', 
                          hour: '2-digit', minute: '2-digit', second: '2-digit'
                        })}
                      </TableCell>
                      <TableCell className="align-top pt-3">
                        <Badge variant="outline" className="bg-background border-primary/20 text-primary text-[10px] font-medium whitespace-nowrap overflow-hidden text-ellipsis max-w-[140px]" title={ev._source_template}>
                          {ev._source_template}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-xs max-w-xl align-top pt-3 pb-3">
                        <div className="flex flex-col gap-1">
                          {eventAction && (
                            <span className={actionColor}>[{eventAction}]</span>
                          )}
                          <span className="font-mono text-[11px] text-muted-foreground break-all whitespace-pre-wrap leading-relaxed">
                            {ev.raw_data || JSON.stringify(ev, null, 2)}
                          </span>
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
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
        </ScrollArea>
      </CardContent>
    </Card>
  );
}