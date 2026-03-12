import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { investigationService, type InvestigationDirective, extractParameters } from "@/services/investigation";
import { incidentService } from "@/services/incidents";
import { forensicsService } from "@/services/forensics";
import { toast } from "sonner";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

// 引入共享类型与拆分出的子组件
import type { InvestigationPageProps, MergedEvent } from "./types";
import { ContextPanel } from "./ContextPanel";
import { DirectivesPanel } from "./DirectivesPanel";
import { TimelinePanel } from "./TimelinePanel";

export default function InvestigationPage({ tabData }: InvestigationPageProps) {
  // 1. 路由与初始上下文识别
  const [searchParams] = useSearchParams();
  const urlIncidentId = searchParams.get("incident_id");
  const activeIncidentId = tabData?.incident_id?.toString() || urlIncidentId;
  
  // 取证上下文
  const forensicsCaseId = tabData?.case_id;
  const forensicsFileId = tabData?.file_id;

  // 2. 核心状态池
  const [templates, setTemplates] = useState<InvestigationDirective[]>([]);
  const [selectedTemplates, setSelectedTemplates] = useState<number[]>([]);
  
  // Incident 完整数据与 Alert 切换状态
  const [incidentData, setIncidentData] = useState<any>(null);
  const [forensicsData, setForensicsData] = useState<any>(null);
  const [selectedAlertIdx, setSelectedAlertIndex] = useState<string>("0");

  // 左侧情报面板的变量池
  const [contextVars, setContextVars] = useState<Record<string, string>>(tabData?.params || {});
  const [newVarKey, setNewVarKey] = useState("");
  const [newVarValue, setNewVarValue] = useState("");

  const [loading, setLoading] = useState(false);
  const [mergedEvents, setMergedEvents] = useState<MergedEvent[]>([]);

  // ==================== 初始化与生命周期 ====================

  // 初始化加载调查规则 (从 Rule Center 获取)
  useEffect(() => {
    fetchTemplates();
  }, []);

  const fetchTemplates = async () => {
    try {
      // 新版：从 Rule Center 获取 type="investigation" 的规则
      const res = await investigationService.listRules();
      if (res.code === 200 && res.data?.rules) {
        // 转换 Rule 格式为 Directive 格式，并自动提取 parameters
        const directives: InvestigationDirective[] = res.data.rules
          .filter((r: any) => r.type === "investigation")
          .map((r: any) => ({
            id: r.id,
            name: r.name,
            description: r.description || "",
            logsql: r.query, // query -> logsql
            parameters: JSON.stringify(extractParameters(r.query)), // 从 query 自动提取参数
          }));
        setTemplates(directives);
      }
    } catch (error) {
      console.error("Failed to load investigation rules:", error);
      toast.error("Failed to load investigation rules");
    }
  };

  // 页面加载时抓取 Incident 数据
  useEffect(() => {
    if (activeIncidentId && Object.keys(contextVars).length === 0) {
      loadIncidentContext(activeIncidentId);
    }
  }, [activeIncidentId]);

  // 页面加载时抓取取证数据
  useEffect(() => {
    if (forensicsCaseId && Object.keys(contextVars).length === 0) {
      loadForensicsContext(forensicsCaseId, forensicsFileId);
    }
  }, [forensicsCaseId, forensicsFileId]);

  const loadIncidentContext = async (id: string) => {
    try {
      const res = await incidentService.detail(Number(id));
      if (res.code === 200 && res.data) {
        setIncidentData(res.data);
        applyAlertContext(res.data, 0); // 默认提取第 0 条告警的上下文
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
        
        // 提取文件信息作为上下文
        const file = fileId 
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

  // ==================== 核心逻辑处理器 ====================

  // 根据选中的 Alert 动态解析上下文
  const applyAlertContext = (incident: any, alertIndex: number) => {
    const newVars: Record<string, string> = { incident_id: String(incident.ID || incident.id) };
    
    let baseTime: Date | null = null;

    // 1. 优先提取：尝试从告警原文 (VictoriaLogs JSON) 中提取绝对真理时间
    if (incident.alerts && incident.alerts.length > alertIndex) {
      const alert = incident.alerts[alertIndex];
      if (alert.content) {
        try {
          const contentObj = JSON.parse(alert.content);
          
          // 如果原文有 _time，以此为基准！这是最准的！
          if (contentObj._time) {
            baseTime = new Date(contentObj._time);
          }

          // 展平 JSON (支持深层嵌套字典解析)
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
          
          // OCSF 常用字段别名映射
          if (newVars["observer.hostname"]) newVars["hostname"] = newVars["observer.hostname"];
          if (newVars["src_endpoint.ip"]) newVars["src_ip"] = newVars["src_endpoint.ip"];
          if (newVars["target_user.name"]) newVars["username"] = newVars["target_user.name"];
          else if (newVars["actor.user.name"]) newVars["username"] = newVars["actor.user.name"];
          if (newVars["process.name"]) newVars["process_name"] = newVars["process.name"];
        } catch(e) {
          console.warn("Failed to parse alert JSON content", e);
        }
      }
    }

    // 2. 兜底提取：如果 JSON 里没有时间，再用数据库的 CreatedAt/FirstSeen
    if (!baseTime || isNaN(baseTime.getTime())) {
       const fallbackTimeStr = incident.first_seen || incident.CreatedAt;
       if (fallbackTimeStr) {
           baseTime = new Date(fallbackTimeStr);
       } else {
           baseTime = new Date(); // 最终兜底为当前时间
       }
    }

    // 3. 以 baseTime 为中心，往前推 2 小时，往后推 2 小时
    const start = new Date(baseTime.getTime() - 2 * 3600 * 1000).toISOString();
    const end = new Date(baseTime.getTime() + 2 * 3600 * 1000).toISOString();
    
    // VictoriaLogs 喜欢 2026-03-03T00:40:00Z 这种格式 (去掉毫秒)
    newVars['start_time'] = start.split('.')[0] + 'Z';
    newVars['end_time'] = end.split('.')[0] + 'Z';

    // 覆盖更新左侧面板
    setContextVars(newVars); 
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

  // 并发执行所有的调查指令
  const handleExecuteInvestigation = async () => {
    if (selectedTemplates.length === 0) {
      toast.warning("Please select at least one investigation directive.");
      return;
    }

    setLoading(true);
    setMergedEvents([]); 
    
    let allEvents: MergedEvent[] = [];
    let updatedContext: Record<string, string> = { ...contextVars };

    try {
      const promises = selectedTemplates.map(async (templateId) => {
        const template = templates.find(t => t.id === templateId);
        const reqData = {
          rule_id: templateId, // 使用 rule_id (新版 Rule Center)
          incident_id: activeIncidentId ? parseInt(activeIncidentId) : undefined,
          params: contextVars,
        };

        const res = await investigationService.execute(reqData);
        if (res.code === 200 && res.data) {
          if (res.data.context_used) {
            updatedContext = { ...updatedContext, ...res.data.context_used };
          }
          // 使用后端返回的 rule_name
          return (res.data.events || []).map((ev: any) => ({
            ...ev,
            _time: ev._time || ev.time || ev.timestamp || new Date().toISOString(),
            _source_template: res.data.rule_name || template?.name || "Unknown Rule",
          }));
        }
        return [];
      });

      const resultsArray = await Promise.all(promises);
      resultsArray.forEach(events => { allEvents = [...allEvents, ...events]; });
      
      // 全局时间排序
      allEvents.sort((a, b) => new Date(b._time).getTime() - new Date(a._time).getTime());

      setMergedEvents(allEvents);
      setContextVars(updatedContext); 
      toast.success(`Investigation completed`, { description: `Found ${allEvents.length} correlated events.` });
    } catch (error: any) {
      toast.error("Investigation failed", { description: error.message });
    } finally {
      setLoading(false);
    }
  };

  // ==================== 视图渲染 ====================

  return (
    <div className="p-6 h-full flex flex-col md:flex-row gap-6">
      
      {/* 左侧区域：源头选择器 + 情报面板 */}
      <div className="w-full md:w-80 flex flex-col gap-4 flex-none">
        
        {/* Alert 切换器 (仅在有多个告警时显示) */}
        {incidentData && incidentData.alerts?.length > 1 && (
          <div className="bg-card border rounded-lg p-3 shadow-sm flex flex-col gap-2 border-l-4 border-l-blue-500">
            <Label className="text-xs text-muted-foreground flex justify-between">
              <span>Select Context Source</span>
              <Badge variant="secondary" className="text-[10px]">{incidentData.alerts.length} Alerts</Badge>
            </Label>
            <Select value={selectedAlertIdx} onValueChange={handleAlertChange}>
              <SelectTrigger className="h-8 text-xs font-mono">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {incidentData.alerts.map((al: any, idx: number) => {
                  const t = al.created_at || al.CreatedAt || al._time;
                  return (
                    <SelectItem key={idx} value={String(idx)} className="text-xs font-mono">
                      #{al.id} - {t ? new Date(t).toLocaleTimeString() : 'Unknown'}
                    </SelectItem>
                  )
                })}
              </SelectContent>
            </Select>
          </div>
        )}
        
        <ContextPanel 
          activeIncidentId={activeIncidentId}
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

      {/* 右侧区域：指令管理 + 结果画板 */}
      <div className="flex-1 flex flex-col gap-4 min-w-0 overflow-hidden">
        <DirectivesPanel 
          templates={templates}
          selectedTemplates={selectedTemplates}
          onChangeSelection={setSelectedTemplates} // 传递选中的模板 ID 数组
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