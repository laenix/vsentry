import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { investigationService, type InvestigationDirective, extractParameters } from "@/services/investigation";
import { incidentService } from "@/services/incidents";
import { forensicsService } from "@/services/forensics";
import { toast } from "sonner";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

// 引入共享Type与拆分出的子Group件 - type { InvestigationPageProps, MergedEvent } from "./types";
import { ContextPanel } from "./ContextPanel";
import { DirectivesPanel } from "./DirectivesPanel";
import { TimelinePanel } from "./TimelinePanel";

// Time窗口Option - TIME_RANGE_OPTIONS = [
  { value: "1h", label: "±1 hour", hours: 1 },
  { value: "2h", label: "±2 hours", hours: 2 },
  { value: "6h", label: "±6 hours", hours: 6 },
  { value: "12h", label: "±12 hours", hours: 12 },
  { value: "24h", label: "±24 hours", hours: 24 },
  { value: "7d", label: "±7 days", hours: 24 * 7 },
  { value: "unlimited", label: "Unlimited", hours: 0 },
];

export default function InvestigationPage({ tabData }: InvestigationPageProps) {
  //   1. 路由与初始上下文识别
  const [searchParams] = useSearchParams();
  const urlIncidentId = searchParams.get("incident_id");
  const activeIncidentId = tabData?.incident_id?.toString() || urlIncidentId;
  
  // Forensics上下文 - forensicsCaseId = tabData?.case_id;
  const forensicsFileId = tabData?.file_id;

  //   2. 核心Status池
  const [templates, setTemplates] = useState<InvestigationDirective[]>([]);
  const [selectedTemplates, setSelectedTemplates] = useState<number[]>([]);
  
  // Time窗口Config - [timeRange, setTimeRange] = useState<string>("2h");
  
  // Incident - const [incidentData, setIncidentData] = useState<any>(null);
  const [incidents, setIncidents] = useState<any[]>([]); // EventList - [selectedIncidentId, setSelectedIncidentId] = useState<string>(activeIncidentId || ""); // 当前选Medium的Event - [forensicsData, setForensicsData] = useState<any>(null);
  const [selectedAlertIdx, setSelectedAlertIndex] = useState<string>("0");

  // 左侧情报Panel的Variable池 - [contextVars, setContextVars] = useState<Record<string, string>>(tabData?.params || {});
  const [newVarKey, setNewVarKey] = useState("");
  const [newVarValue, setNewVarValue] = useState("");

  const [loading, setLoading] = useState(false);
  const [mergedEvents, setMergedEvents] = useState<MergedEvent[]>([]);

  //   ==================== 初始化与生命周期 ====================

  //   初始化LoadInvestigationRule (从 Rule Center Get)
  useEffect(() => {
    fetchTemplates();
  }, []);

  // LoadEventList - (() => {
    fetchIncidents();
  }, []);

  const fetchIncidents = async () => {
    try {
      const res = await incidentService.list();
      if (res.code === 200 && res.data) {
        setIncidents(res.data);
      }
    } catch (err) {
      console.error("Failed to load incidents:", err);
    }
  };

  const fetchTemplates = async () => {
    try {
      //   New版：从 Rule Center Get type="investigation" 的Rule
      const res = await investigationService.listRules();
      if (res.code === 200 && res.data?.rules) {
        // Convert - 格式为 Directive 格式，并自动提取 parameters
        const directives: InvestigationDirective[] = res.data.rules
          .filter((r: any) => r.type === "investigation")
          .map((r: any) => ({
            id: r.id,
            name: r.name,
            description: r.description || "",
            logsql: r.query, //   query -> logsql
            parameters: JSON.stringify(extractParameters(r.query)), // 从 - 自动提取Parameter
          }));
        setTemplates(directives);
      }
    } catch (error) {
      console.error("Failed to load investigation rules:", error);
      toast.error("Failed to load investigation rules");
    }
  };

  // 页面Load时抓取 - Data
  useEffect(() => {
    //   如果有选Medium的EventID，Load该Event
    if (selectedIncidentId) {
      loadIncidentContext(selectedIncidentId);
    }
  }, [selectedIncidentId]);

  // 页面Load时抓取ForensicsData - (() => {
    if (forensicsCaseId && Object.keys(contextVars).length === 0) {
      loadForensicsContext(forensicsCaseId, forensicsFileId);
    }
  }, [forensicsCaseId, forensicsFileId]);

  // HandleEvent切换 - handleIncidentChange = (id: string) => {
    if (!id || id === "none") {
      setSelectedIncidentId("");
      setIncidentData(null);
      setContextVars({});
      setMergedEvents([]);
      return;
    }
    setSelectedIncidentId(id);
  };

  const loadIncidentContext = async (id: string) => {
    try {
      const res = await incidentService.detail(Number(id));
      if (res.code === 200 && res.data) {
        setIncidentData(res.data);
        applyAlertContext(res.data, 0); // 默认提取第 - 条Alert的上下文
      }
    } catch (err) { 
      console.error("Auto extract failed", err); 
    }
  };

  const loadForensicsContext = async (caseId: number, fileId?: number) => {
    try {
      const res = await forensicsService.getTask(caseId);
      if (res.code === 200 && res.data) {
        setForensicsData(res.data);
        
        // 提取FileInfo作为上下文 - file = fileId 
          ? res.data.files?.find((f: any) => f.id === fileId)
          : res.data.files?.[0];
        
        if (file) {
          const vars: Record<string, string> = {
            case_id: String(caseId),
            file_id: String(file.id),
            file_type: file.file_type,
            file_name: file.original_name,
            event_count: String(file.event_count || 0),
          };
          setContextVars(vars);
        }
      }
    } catch (err) {
      console.error("Load forensics context failed", err);
    }
  };

  //   ==================== 核心逻辑Handle器 ====================

  // GetTime窗口Config - getTimeRangeConfig = () => {
    const option = TIME_RANGE_OPTIONS.find(o => o.value === timeRange);
    return option || TIME_RANGE_OPTIONS[1]; // 默认 - };

  // 根据选Medium的 - 动态Parse上下文
  const applyAlertContext = (incident: any, alertIndex: number) => {
    const newVars: Record<string, string> = { incident_id: String(incident.ID || incident.id) };
    
    let baseTime: Date | null = null;

    //   1. 优先提取：尝试从Alert原文 (VictoriaLogs JSON) Medium提取绝对真理Time
    if (incident.alerts && incident.alerts.length > alertIndex) {
      const alert = incident.alerts[alertIndex];
      if (alert.content) {
        try {
          const contentObj = JSON.parse(alert.content);
          
          // 如果原文有 - ，以此为基准！这是最准的！
          if (contentObj._time) {
            baseTime = new Date(contentObj._time);
          }

          // 展平 - (支持深层嵌套字典Parse)
          const flatten = (obj: any, prefix = '') => {
            for (const key in obj) {
              const val = obj[key];
              const newKey = prefix ? `${prefix}.${key}` : key;
              if (typeof val === 'object' && val !== null && !Array.isArray(val)) { 
                flatten(val, newKey); 
              } else if (val !== null && val !== undefined) { 
                newVars[newKey] = String(val); 
              }
            }
          };
          flatten(contentObj);
          
          // OCSF - if (newVars["observer.hostname"]) newVars["hostname"] = newVars["observer.hostname"];
          if (newVars["src_endpoint.ip"]) newVars["src_ip"] = newVars["src_endpoint.ip"];
          if (newVars["target_user.name"]) newVars["username"] = newVars["target_user.name"];
          else if (newVars["actor.user.name"]) newVars["username"] = newVars["actor.user.name"];
          if (newVars["process.name"]) newVars["process_name"] = newVars["process.name"];
        } catch(e) {
          console.warn("Failed to parse alert JSON content", e);
        }
      }
    }

    //   2. 兜底提取：如果 JSON 里没有Time，再用Database的 CreatedAt/FirstSeen
    if (!baseTime || isNaN(baseTime.getTime())) {
       const fallbackTimeStr = incident.first_seen || incident.CreatedAt;
       if (fallbackTimeStr) {
           baseTime = new Date(fallbackTimeStr);
       } else {
           baseTime = new Date(); //   最终兜底为当前Time
       }
    }

    //   3. 根据Config计算Time窗口
    const config = getTimeRangeConfig();
    if (config.hours === 0) {
      //   Unlimited: 不SettingsTimeLimit
      delete newVars['start_time'];
      delete newVars['end_time'];
    } else {
      const start = new Date(baseTime.getTime() - config.hours * 3600 * 1000).toISOString();
      const end = new Date(baseTime.getTime() + config.hours * 3600 * 1000).toISOString();
      
      // VictoriaLogs - 2026-03-03T00:40:00Z 这种格式 (去掉毫秒)
      newVars['start_time'] = start.split('.')[0] + 'Z';
      newVars['end_time'] = end.split('.')[0] + 'Z';
    }

    // 覆盖Update左侧Panel - (newVars); 
  };

  const handleAlertChange = (val: string) => {
    setSelectedAlertIndex(val);
    if (incidentData) applyAlertContext(incidentData, parseInt(val));
  };

  const handleAddVar = () => {
    if (!newVarKey.trim() || !newVarValue.trim()) return;
    setContextVars(prev => ({ ...prev, [newVarKey.trim()]: newVarValue.trim() }));
    setNewVarKey("");
    setNewVarValue("");
  };

  const handleRemoveVar = (key: string) => {
    const updated = { ...contextVars };
    delete updated[key];
    setContextVars(updated);
  };

  //   并发Execute所有的Investigation指令 (单个Failed不影响其他)
  const handleExecuteInvestigation = async () => {
    if (selectedTemplates.length === 0) {
      toast.warning("Please select at least one investigation directive.");
      return;
    }

    setLoading(true);
    setMergedEvents([]); 
    
    let allEvents: MergedEvent[] = [];
    let updatedContext: Record<string, string> = { ...contextVars };
    let failedRules: string[] = [];

    try {
      //   并发Execute，每个Rule独立HandleFailed
      const resultsArray = await Promise.allSettled(
        selectedTemplates.map(async (templateId) => {
          const template = templates.find(t => t.id === templateId);
          const reqData = {
            rule_id: templateId,
            incident_id: activeIncidentId ? parseInt(activeIncidentId) : undefined,
            params: contextVars,
          };

          const res = await investigationService.execute(reqData);
          if (res.code === 200 && res.data) {
            if (res.data.context_used) {
              updatedContext = { ...updatedContext, ...res.data.context_used };
            }
            //   HandleTime字段 - VictoriaLogs 可能Return两个 _time 字段（ISO Sum Unix timestamp），后者会覆盖前者
            return (res.data.events || []).map((ev: any) => {
              let eventTime = ev._time;
              // 如果 - 是 Unix Time戳（纯数字字符串），尝试Convert
              if (eventTime && !isNaN(Number(eventTime)) && String(eventTime).includes('.')) {
                eventTime = new Date(Number(eventTime) * 1000).toISOString();
              } else if (eventTime && !isNaN(Number(eventTime))) {
                // 毫秒Time戳 - = new Date(Number(eventTime)).toISOString();
              }
              return {
                ...ev,
                _time: eventTime || ev.timestamp || new Date().toISOString(),
                _source_template: res.data.rule_name || template?.name || "Unknown Rule",
                _rule_query: template?.logsql || "", //   AddRuleQuery语句
              };
            });
          }
          // 如果不是 - ，抛出Error
          throw new Error(res.msg || "Execution failed");
        })
      );

      //   Handle结果，区分SuccessSumFailed
      resultsArray.forEach((result, idx) => {
        if (result.status === 'fulfilled') {
          allEvents = [...allEvents, ...(result.value as MergedEvent[])];
        } else {
          const template = templates.find(t => t.id === selectedTemplates[idx]);
          failedRules.push(template?.name || `Rule #${selectedTemplates[idx]}`);
          console.error(`Rule ${selectedTemplates[idx]} failed:`, result.reason);
        }
      });
      
      // 全局TimeSort - .sort((a, b) => new Date(b._time).getTime() - new Date(a._time).getTime());

      setMergedEvents(allEvents);
      setContextVars(updatedContext); 

      // 提示结果 - (failedRules.length > 0) {
        toast.warning(`Investigation completed with ${failedRules.length} failed rule(s)`, { 
          description: `Failed: ${failedRules.join(', ')}` 
        });
      } else {
        toast.success(`Investigation completed`, { description: `Found ${allEvents.length} correlated events.` });
      }
    } catch (error: any) {
      toast.error("Investigation failed", { description: error.message });
    } finally {
      setLoading(false);
    }
  };

  //   ==================== View渲染 ====================

  return (
    <div className="p-6 h-full flex flex-col md:flex-row gap-6">
      
      {/* 左侧区域：只放 Investigation Context */}
      <div className="w-full md:w-80 flex flex-col gap-4 flex-none">
        <ContextPanel 
          activeIncidentId={selectedIncidentId || undefined}
          contextVars={contextVars}
          newVarKey={newVarKey}
          newVarValue={newVarValue}
          setNewVarKey={setNewVarKey}
          setNewVarValue={setNewVarValue}
          handleAddVar={handleAddVar}
          handleRemoveVar={handleRemoveVar}
          forensicsCaseId={forensicsCaseId}
          forensicsFileId={forensicsFileId}
          forensicsFileName={forensicsData?.files?.find((f: any) => f.id === forensicsFileId)?.original_name}
        />
      </div>

      {/* 右侧区域：AlertSelect器 + Time范围 + 指令Manage + 结果画板 */}
      <div className="flex-1 flex flex-col gap-4 min-w-0 overflow-hidden">
        
        {/* 顶部控制栏：Incident + Alert + Time范围 */}
        <div className="flex gap-4 flex-wrap">
          {/* Incident Select器 */}
          <div className="bg-card border rounded-lg p-3 shadow-sm flex flex-col gap-2 border-l-4 border-l-primary flex-1 min-w-[200px]">
            <Label className="text-xs text-muted-foreground flex justify-between">
              <span>Incident</span>
              <Badge variant="secondary" className="text-[10px]">{incidents.length} total</Badge>
            </Label>
            <Select value={selectedIncidentId} onValueChange={handleIncidentChange}>
              <SelectTrigger className="h-8 text-xs">
                <SelectValue placeholder="Select incident..." />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none" className="text-xs">-- Manual Mode --</SelectItem>
                {incidents.map((inc: any) => {
                  const id = inc.ID || inc.id || 0;
                  const name = inc.name || `Incident #${id}`;
                  return (
                    <SelectItem key={id} value={String(id)} className="text-xs">
                      #{id} - {name.substring(0, 25)}{name.length > 25 ? '...' : ''}
                    </SelectItem>
                  );
                })}
              </SelectContent>
            </Select>
          </div>

          {/* Alert Select器 (仅在有多个Alert时显示) */}
          {incidentData && incidentData.alerts?.length > 0 && (
            <div className="bg-card border rounded-lg p-3 shadow-sm flex flex-col gap-2 border-l-4 border-l-blue-500 flex-1 min-w-[200px]">
              <Label className="text-xs text-muted-foreground flex justify-between">
                <span>Alert / Evidence</span>
                <Badge variant="secondary" className="text-[10px]">{incidentData.alerts.length} alerts</Badge>
              </Label>
              <Select value={selectedAlertIdx} onValueChange={handleAlertChange}>
                <SelectTrigger className="h-8 text-xs font-mono">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {incidentData.alerts.map((al: any, idx: number) => {
                    const t = al.created_at || al.CreatedAt || al._time;
                    const timeStr = t ? new Date(t).toLocaleString('en-CA', { 
                      year: 'numeric', month: '2-digit', day: '2-digit', 
                      hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false
                    }) : 'Unknown';
                    return (
                      <SelectItem key={idx} value={String(idx)} className="text-xs font-mono">
                        #{al.id} - {timeStr}
                      </SelectItem>
                    )
                  })}
                </SelectContent>
              </Select>
            </div>
          )}

          {/* Time窗口Select器 */}
          <div className="bg-card border rounded-lg p-3 shadow-sm flex flex-col gap-2 border-l-4 border-l-purple-500 flex-1 min-w-[180px]">
            <Label className="text-xs text-muted-foreground flex justify-between">
              <span>Time Range</span>
              {timeRange !== 'unlimited' && (
                <Badge variant="secondary" className="text-[10px] bg-purple-50 text-purple-600">
                  {TIME_RANGE_OPTIONS.find(o => o.value === timeRange)?.label}
                </Badge>
              )}
            </Label>
            <Select value={timeRange} onValueChange={(val) => {
              setTimeRange(val);
              if (incidentData) {
                applyAlertContext(incidentData, parseInt(selectedAlertIdx));
              }
            }}>
              <SelectTrigger className="h-8 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {TIME_RANGE_OPTIONS.map(opt => (
                  <SelectItem key={opt.value} value={opt.value} className="text-xs">
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <DirectivesPanel 
          templates={templates}
          selectedTemplates={selectedTemplates}
          onChangeSelection={setSelectedTemplates}
          contextVars={contextVars}
          loading={loading}
          onExecute={handleExecuteInvestigation}
          onRefreshTemplates={fetchTemplates}
        />
        <TimelinePanel mergedEvents={mergedEvents} />
      </div>
    </div>
  );
}