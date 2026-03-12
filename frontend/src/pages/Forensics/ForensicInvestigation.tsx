import { useEffect, useState, useMemo } from "react";
import { forensicsService } from "@/services/forensics";
import { ruleService, type DetectionRule } from "@/services/rules";
import { useTabStore } from "@/stores/tab-store";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Loader2, Play, FlaskConical, FileText, Clock, ArrowLeft, ExternalLink } from "lucide-react";
import { toast } from "sonner";
import { TimelinePanel } from "@/pages/Investigation/TimelinePanel";

interface ForensicInvestigationProps {
  tabData?: {
    task_id: number;  // 原来是 case_id，现改为 task_id
    file_id: number;
    file_type: string;
    file_name: string;
    triggered_rule?: string;  // 触发的规则
  };
}

interface ForensicResult {
  rule_id: number;
  rule_name: string;
  severity: string;
  matched_data: Record<string, any>[];
  count: number;
  _time: string;
}

export default function ForensicInvestigationPage({ tabData }: ForensicInvestigationProps) {
  const { addTab } = useTabStore();
  
  const [fileInfo, setFileInfo] = useState<any>(null);
  const [rules, setRules] = useState<DetectionRule[]>([]);
  const [selectedRules, setSelectedRules] = useState<number[]>([]);
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<ForensicResult[]>([]);
  const [executing, setExecuting] = useState(false);

  // 把 ForensicResult 转换成 MergedEvent 格式给 TimelinePanel 使用
  const forensicEvents = useMemo(() => {
    const events: any[] = [];
    results.forEach((result) => {
      result.matched_data.forEach((data: any) => {
        events.push({
          ...data,
          _time: result._time,
          _source_template: result.rule_name,
          severity: result.severity,
          activity_name: data.activity_name || data.action || 'Match',
        });
      });
    });
    return events;
  }, [results]);

  // 加载ForensicsRule
  useEffect(() => {
    const fetchRules = async () => {
      try {
        console.log("Fetching rules...");
        const res = await ruleService.list();
        console.log("API Response received:", res);
        
        if (res.code === 200) {
          let list: DetectionRule[] = [];
          if (Array.isArray(res.data)) list = res.data;
          else if (res.data?.rules) list = res.data.rules;
          
          console.log("All rules:", JSON.stringify(list));
          
          // Debug: 显示所有Enable的Rule
          const enabledRules = list.filter(r => r.enabled);
          console.log("Enabled rules:", JSON.stringify(enabledRules));
          
          // 只显示ForensicsRule
          const forensicRules = list.filter(r => (r.type || "alert") === "forensic" && r.enabled);
          console.log("Forensic rules:", JSON.stringify(forensicRules));
          
          // 如果没有ForensicsRule，显示所有Rule以便Debug
          setRules(forensicRules.length > 0 ? forensicRules : enabledRules);
        } else {
          console.error("API error:", res.msg);
        }
      } catch (e) {
        console.error("Fetch error:", e);
      }
    };
    fetchRules();
  }, []);

  // 加载FileInfo
  useEffect(() => {
    const fetchFileInfo = async () => {
      if (!tabData?.task_id) return;
      try {
        const res = await forensicsService.getTask(tabData.task_id);
        if (res.code === 200 && res.data?.files) {
          const file = res.data.files.find((f: any) => f.id === tabData.file_id);
          if (file) setFileInfo(file);
        }
      } catch (e) {
        console.error(e);
      }
    };
    fetchFileInfo();
  }, [tabData]);

  const handleRuleToggle = (ruleId: number) => {
    setSelectedRules(prev => 
      prev.includes(ruleId) 
        ? prev.filter(id => id !== ruleId)
        : [...prev, ruleId]
    );
  };

  const handleExecute = async () => {
    if (selectedRules.length === 0) {
      toast.warning("Please select at least one forensic rule");
      return;
    }

    setExecuting(true);
    setResults([]);

    try {
      const res = await forensicsService.executeRules(
        tabData!.task_id,
        tabData!.file_id,
        selectedRules
      );

      if (res.code === 200 && res.data) {
        const ruleResults = res.data.map((r: any) => ({
          rule_id: r.rule_id,
          rule_name: r.rule_name,
          severity: r.severity,
          matched_data: r.matched_data || [],
          count: r.count,
          _time: new Date().toISOString(),
        }));
        setResults(ruleResults);
        toast.success(`Analysis complete. Found ${ruleResults.length} results`);
      } else {
        throw new Error(res.msg || "执行失败");
      }
    } catch (e: any) {
      toast.error("Analysis failed", { description: e.message });
    } finally {
      setExecuting(false);
    }
  };

  const handleGoBack = () => {
    // 返回取证案件页面
    if (tabData?.task_id) {
      addTab('forensics', 'Forensics', {});
    }
  };

  const handleViewInLogs = (result: ForensicResult) => {
    // 跳转到日志查询页面，使用 | 分隔符
    const query = `env:forensics | task_id:${tabData?.task_id} | file_id:${tabData?.file_id}`;
    addTab('logs', `Logs: ${tabData?.file_name}`, { query });
  };

  if (!tabData) {
    return (
      <div className="p-6 flex items-center justify-center h-full">
        <p className="text-muted-foreground">Missing analysis parameters</p>
      </div>
    );
  }

  return (
    <div className="p-6 h-full flex flex-col gap-6">
      {/* 顶部导航 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={handleGoBack}>
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <h1 className="text-lg font-bold flex items-center gap-2">
              <FlaskConical className="w-5 h-5 text-purple-500" />
              Forensic Analysis
            </h1>
            <p className="text-sm text-muted-foreground">
              {tabData.file_name}
            </p>
          </div>
        </div>
        
        {fileInfo && (
          <div className="flex items-center gap-4 text-sm">
            <Badge variant="outline">{fileInfo.file_type}</Badge>
            <span className="text-muted-foreground">
              {fileInfo.event_count} events
            </span>
          </div>
        )}
      </div>

      <div className="flex-1 flex gap-6 min-h-0">
        {/* Left: Forensic Rule Selection */}
        <div className="w-80 flex flex-col gap-4">
          <Card className="flex-1 flex flex-col">
            <CardHeader className="py-3 border-b">
              <CardTitle className="text-sm">Forensic Rules</CardTitle>
              <CardDescription className="text-xs">Select forensic rules to execute</CardDescription>
            </CardHeader>
            <CardContent className="flex-1 p-0 overflow-hidden">
              <ScrollArea className="h-full">
                {rules.length === 0 ? (
                  <div className="p-4 text-center text-muted-foreground text-sm">
                    No enabled forensic rules
                  </div>
                ) : (
                  <div className="p-2 space-y-2">
                    {rules.map(rule => {
                      const ruleId = rule.ID || rule.id || 0;
                      const isSelected = selectedRules.includes(ruleId);
                      return (
                        <div
                          key={ruleId}
                          className={`p-3 rounded-lg border cursor-pointer transition-all ${
                            isSelected 
                              ? 'border-purple-500 bg-purple-50 dark:bg-purple-950' 
                              : 'hover:border-purple-300 hover:bg-purple-50/50'
                          }`}
                          onClick={() => handleRuleToggle(ruleId)}
                        >
                          <div className="flex items-center justify-between">
                            <span className="font-medium text-sm">{rule.name}</span>
                            <input 
                              type="checkbox" 
                              checked={isSelected}
                              onChange={() => handleRuleToggle(ruleId)}
                              className="w-4 h-4 accent-purple-500"
                            />
                          </div>
                          {rule.description && (
                            <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                              {rule.description}
                            </p>
                          )}
                        </div>
                      );
                    })}
                  </div>
                )}
              </ScrollArea>
            </CardContent>
          </Card>

          <Button 
            onClick={handleExecute} 
            disabled={executing || selectedRules.length === 0}
            className="w-full bg-purple-600 hover:bg-purple-700"
          >
            {executing ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Analyzing...
              </>
            ) : (
              <>
                <Play className="w-4 h-4 mr-2" />
                Run Analysis ({selectedRules.length})
              </>
            )}
          </Button>
        </div>

        {/* Right: Analysis Results Timeline */}
        <Card className="flex-1 flex flex-col overflow-hidden">
          <CardHeader className="py-3 border-b flex-none">
            <CardTitle className="text-sm">Analysis Results</CardTitle>
            <CardDescription className="text-xs">Forensic rule detection results</CardDescription>
          </CardHeader>
          <CardContent className="flex-1 p-0 overflow-hidden">
            {results.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center text-muted-foreground gap-3">
                <FlaskConical className="w-10 h-10 opacity-20" />
                <p className="text-sm">Select rules and run analysis</p>
              </div>
            ) : (
              <TimelinePanel 
                mergedEvents={forensicEvents} 
                onOpenInLogs={(ev) => {
                  const queryParts: string[] = [];
                  if (ev.src_ip) queryParts.push(`src_ip:${ev.src_ip}`);
                  if (ev.dst_ip) queryParts.push(`dst_ip:${ev.dst_ip}`);
                  if (ev.hostname) queryParts.push(`hostname:${ev.hostname}`);
                  const query = queryParts.length > 0 ? queryParts.join(' | ') : '*';
                  addTab('logs', 'Logs Query', { query });
                }}
              />
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
