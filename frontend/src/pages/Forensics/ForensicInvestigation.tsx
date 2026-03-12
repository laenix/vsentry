import { useEffect, useState, useMemo } from "react";
import { forensicsService } from "@/services/forensics";
import { ruleService, type DetectionRule } from "@/services/rules";
import { useTabStore } from "@/stores/tab-store";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Loader2, Play, FlaskConical, Clock, ArrowLeft, ExternalLink, Search, X, ChevronDown, ChevronUp, Copy, Check } from "lucide-react";
import { toast } from "sonner";

interface ForensicInvestigationProps {
  tabData?: {
    case_id: number;
    file_id: number;
    file_type: string;
    file_name: string;
  };
}

interface ForensicResult {
  rule_id: number;
  rule_name: string;
  rule_query: string; // Rule的实际Query语句 - : string;
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
  
  // SearchSum展开Status - [searchQuery, setSearchQuery] = useState("");
  const [expandedResults, setExpandedResults] = useState<Set<number>>(new Set());
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

  // LoadForensicsRule - (() => {
    const fetchRules = async () => {
      try {
        const res = await ruleService.list();
        if (res.code === 200) {
          let list: DetectionRule[] = [];
          if (Array.isArray(res.data)) list = res.data;
          else if (res.data?.rules) list = res.data.rules;
          
          const forensicRules = list.filter(r => (r.type || "alert") === "forensic" && r.enabled);
          setRules(forensicRules.length > 0 ? forensicRules : list.filter(r => r.enabled));
        }
      } catch (e) {
        console.error("Fetch error:", e);
      }
    };
    fetchRules();
  }, []);

  // LoadFileInfo - (() => {
    const fetchFileInfo = async () => {
      if (!tabData?.case_id) return;
      try {
        const res = await forensicsService.getTask(tabData.case_id);
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
        tabData!.case_id,
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
        throw new Error(res.msg || "ExecuteFailed");
      }
    } catch (e: any) {
      toast.error("Analysis failed", { description: e.message });
    } finally {
      setExecuting(false);
    }
  };

  const handleGoBack = () => {
    if (tabData?.case_id) {
      addTab('forensics', 'Forensics', {});
    }
  };

  const handleViewInLogs = (result: ForensicResult) => {
    //   使用 | 分隔QueryCondition，rule_query 是实际的Query语句
    let query = `env:forensics`;
    query += ` | task_id:${tabData?.case_id}`;
    query += ` | forensic_file_id:${tabData?.file_id}`;
    // AddRule的Query语句 - (result.rule_query) {
      query += ` | ${result.rule_query}`;
    }
    addTab('logs', `Logs: ${tabData?.file_name}`, { query });
  };

  // Filter结果 - filteredResults = useMemo(() => {
    if (!searchQuery.trim()) return results;
    const query = searchQuery.toLowerCase();
    return results.filter(r => 
      r.rule_name.toLowerCase().includes(query) ||
      r.severity.toLowerCase().includes(query) ||
      JSON.stringify(r.matched_data).toLowerCase().includes(query)
    );
  }, [results, searchQuery]);

  // 切换展开 - toggleExpand = (idx: number) => {
    setExpandedResults(prev => {
      const next = new Set(prev);
      if (next.has(idx)) next.delete(idx);
      else next.add(idx);
      return next;
    });
  };

  // 复制 - copyResult = (result: ForensicResult, idx: number) => {
    const jsonStr = JSON.stringify(result.matched_data, null, 2);
    if (navigator.clipboard?.writeText) {
      navigator.clipboard.writeText(jsonStr).then(() => {
        setCopiedIndex(idx);
        toast.success("Copied to clipboard");
        setTimeout(() => setCopiedIndex(null), 2000);
      });
    } else {
      const textarea = document.createElement('textarea');
      textarea.value = jsonStr;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      setCopiedIndex(idx);
      toast.success("Copied to clipboard");
      setTimeout(() => setCopiedIndex(null), 2000);
    }
  };

  //   FormatTime - 数字格式，带年份
  const formatTime = (timeStr: string) => {
    if (!timeStr) return "Unknown";
    const date = new Date(timeStr);
    if (isNaN(date.getTime())) return timeStr;
    return date.toLocaleString('en-CA', { 
      year: 'numeric', month: '2-digit', day: '2-digit', 
      hour: '2-digit', minute: '2-digit', second: '2-digit',
      hour12: false
    });
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
              {fileInfo.event_count} Event
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
        <Card className="flex-1 flex flex-col border-t-4 border-t-purple-500">
          <CardHeader className="py-3 border-b bg-muted/10">
            <div className="flex justify-between items-center gap-4">
              <CardTitle className="text-sm">Analysis Results</CardTitle>
              
              {/* Search框 */}
              <div className="flex items-center gap-2">
                <div className="relative">
                  <Search className="w-3 h-3 absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground" />
                  <Input 
                    placeholder="Search results..." 
                    className="h-7 text-xs pl-7 w-[180px]"
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
                {results.length > 0 && (
                  <Badge variant="outline" className="text-[10px]">
                    {searchQuery ? `${filteredResults.length} / ${results.length}` : results.length}
                  </Badge>
                )}
              </div>
            </div>
          </CardHeader>
          <CardContent className="flex-1 p-0 overflow-hidden">
            <ScrollArea className="h-full">
              {filteredResults.length === 0 ? (
                <div className="h-full flex flex-col items-center justify-center text-muted-foreground gap-3">
                  <FlaskConical className="w-10 h-10 opacity-20" />
                  <p className="text-sm">
                    {searchQuery ? "No matching results" : "Select rules and run analysis"}
                  </p>
                </div>
              ) : (
                <div className="py-4 px-6">
                  {/* Time线 */}
                  <div className="relative">
                    {/* 垂直线 */}
                    <div className="absolute left-4 top-0 bottom-0 w-0.5 bg-gradient-to-b from-purple-500/30 via-purple-500/10 to-transparent" />
                    
                    {filteredResults.map((result, idx) => {
                      const isExpanded = expandedResults.has(idx);
                      let severityColor = "bg-blue-500";
                      if (result.severity?.toLowerCase() === "critical" || result.severity?.toLowerCase() === "high") {
                        severityColor = "bg-red-500";
                      } else if (result.severity?.toLowerCase() === "medium") {
                        severityColor = "bg-amber-500";
                      } else if (result.severity?.toLowerCase() === "low") {
                        severityColor = "bg-green-500";
                      }

                      return (
                        <div key={idx} className="relative flex gap-4 pb-6 last:pb-0 group">
                          {/* Time点圆圈 */}
                          <div className={`relative z-10 w-2 h-2 rounded-full mt-2 shrink-0 ${severityColor} ring-4 ring-background group-hover:scale-125 transition-transform`} />
                          
                          {/* 结果卡片 */}
                          <div className="flex-1 min-w-0">
                            {/* 头部 */}
                            <div className="flex items-center gap-2 mb-1">
                              <span className="text-xs font-mono text-muted-foreground whitespace-nowrap">
                                {formatTime(result._time)}
                              </span>
                              <Badge variant="outline" className={`${
                                result.severity === 'critical' ? 'border-red-500 text-red-500' :
                                result.severity === 'high' ? 'border-orange-500 text-orange-500' :
                                result.severity === 'medium' ? 'border-amber-500 text-amber-500' :
                                'border-blue-500 text-blue-500'
                              }`}>
                                {result.severity}
                              </Badge>
                              <span className="font-medium text-sm">{result.rule_name}</span>
                              <Badge variant="secondary" className="text-[10px]">
                                {result.count} matches
                              </Badge>
                            </div>

                            {/* 内容 */}
                            <div className="bg-card border rounded-md overflow-hidden shadow-sm hover:border-purple-500/20 transition-colors">
                              <div 
                                className="p-2 px-3 flex items-center justify-between cursor-pointer bg-muted/20 hover:bg-muted/30"
                                onClick={() => toggleExpand(idx)}
                              >
                                <div className="flex items-center gap-2 overflow-hidden">
                                  {isExpanded ? (
                                    <ChevronUp className="w-4 h-4 text-muted-foreground shrink-0" />
                                  ) : (
                                    <ChevronDown className="w-4 h-4 text-muted-foreground shrink-0" />
                                  )}
                                  <span className="text-xs text-muted-foreground truncate">
                                    {result.matched_data?.length > 0 
                                      ? JSON.stringify(result.matched_data[0]).substring(0, 80) + "..." 
                                      : "No matched data"}
                                  </span>
                                </div>
                                <div className="flex items-center gap-1 shrink-0">
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-6 px-2 text-[10px]"
                                    onClick={(e) => { e.stopPropagation(); copyResult(result, idx); }}
                                  >
                                    {copiedIndex === idx ? <Check className="w-3 h-3 text-green-500" /> : <Copy className="w-3 h-3" />}
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-6 px-2 text-[10px] text-blue-600 hover:text-blue-700"
                                    onClick={(e) => { e.stopPropagation(); handleViewInLogs(result); }}
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
                                    {JSON.stringify(result.matched_data, null, 2)}
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
              )}
            </ScrollArea>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
