import React, { useEffect, useState } from 'react';
import type { Node } from '@xyflow/react';
import { Settings, Info, Calendar, PlayCircle, Mail, Globe, AlertTriangle, Calculator, Shield, Zap } from 'lucide-react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { MonacoInput } from './MonacoInput';
import { KeyValueList } from './KeyValueList';

interface ConfigPanelProps {
  selectedNode: Node | null;
  onUpdateNode: (id: string, data: any) => void;
  // ✅ 核心：接收来自 PlaybookEditor 轮询到的执行结果
  executionContext?: Record<string, any>;
}

export default function ConfigPanel({ selectedNode, onUpdateNode, executionContext }: ConfigPanelProps) {
  const [config, setConfig] = useState<any>({});

  // 当选中节点切换时，同步本地配置状态
  useEffect(() => {
    if (selectedNode) {
      setConfig(selectedNode.data.config || {});
    }
  }, [selectedNode]);

  // 更新配置并实时同步回 React Flow State
  const updateConfig = (key: string, value: any) => {
    const newConfig = { ...config, [key]: value };
    setConfig(newConfig);
    onUpdateNode(selectedNode!.id, {
      ...selectedNode!.data,
      config: newConfig
    });
  };

  if (!selectedNode) {
    return (
      <div className="h-full flex flex-col items-center justify-center text-muted-foreground p-4 text-center">
        <Settings className="w-12 h-12 mb-4 opacity-10" />
        <p className="text-sm font-medium">Select a node to configure</p>
        <p className="text-xs opacity-60">Configure triggers, logic, or actions</p>
      </div>
    );
  }

  const { type, label } = selectedNode.data;

  return (
    <div className="h-full flex flex-col bg-card border-l">
      {/* 1. 顶部状态栏：展示类型与 ID */}
      <div className="p-4 border-b bg-muted/10">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <span className="text-[10px] font-bold uppercase bg-primary/10 text-primary px-2 py-0.5 rounded tracking-tighter">
              {type === 'trigger' ? (config.trigger_type || 'incident_created').replace('_', ' ') : type?.replace('_', ' ')}
            </span>
            <span className="text-xs text-muted-foreground font-mono">#{selectedNode.id}</span>
          </div>
        </div>
        <input
          className="w-full bg-transparent text-lg font-bold focus:outline-none border-b border-transparent focus:border-primary transition-colors py-1"
          placeholder="Enter node name..."
          value={label}
          onChange={(e) => onUpdateNode(selectedNode.id, { ...selectedNode.data, label: e.target.value })}
        />
      </div>

      <ScrollArea className="flex-1">
        <div className="p-4 space-y-6">
          
          {/* ================= 触发器配置 (Trigger) ================= */}
          {type === 'trigger' && (
             <div className="space-y-4">
               <div className="space-y-2">
                 <label className="text-xs font-semibold flex items-center gap-1">
                   <Zap className="w-3 h-3 text-amber-500" /> Trigger Source
                 </label>
                 <select 
                    className="w-full text-xs px-2 py-2 border rounded bg-background focus:ring-1 focus:ring-primary"
                    value={config.trigger_type || 'incident_created'}
                    onChange={(e) => updateConfig('trigger_type', e.target.value)}
                 >
                    <option value="incident_created">Incident Created (Automatic)</option>
                    <option value="manual">Manual / API Trigger (Testing)</option>
                    <option value="schedule">Scheduled (Cron)</option>
                 </select>
               </div>

               {/* Incident Created 模式 */}
               {(config.trigger_type === 'incident_created' || !config.trigger_type) && (
                 <div className="text-xs text-muted-foreground p-3 border border-dashed rounded bg-slate-50 flex gap-2 italic">
                    <Info className="w-4 h-4 shrink-0 text-slate-400" />
                    <p>Starts automatically when a new incident is created. Context: <code>incident.*</code></p>
                 </div>
               )}

               {/* Manual 模式：方便开发调试 */}
               {config.trigger_type === 'manual' && (
                  <div className="space-y-3">
                    <div className="bg-blue-50 text-blue-700 p-3 rounded text-[11px] flex gap-2">
                      <PlayCircle className="w-4 h-4 shrink-0" />
                      <p>Used for testing or manual intervention. Allows execution without a real incident event.</p>
                    </div>
                    <div className="space-y-2">
                       <label className="text-xs font-semibold">Mock Payload (JSON)</label>
                       <MonacoInput 
                          value={config.payload_template || '{\n  "input": "data"\n}'}
                          onChange={(val) => updateConfig('payload_template', val)}
                          height="120px"
                          language="json"
                       />
                    </div>
                  </div>
               )}

               {/* Schedule 模式 */}
               {config.trigger_type === 'schedule' && (
                  <div className="space-y-3">
                     <div className="bg-amber-50 text-amber-700 p-3 rounded text-[11px] flex gap-2">
                      <Calendar className="w-4 h-4 shrink-0" />
                      <p>Recurring execution based on Cron syntax.</p>
                    </div>
                    <div className="space-y-2">
                       <label className="text-xs font-semibold">Cron Expression</label>
                       <input 
                          className="w-full text-xs px-2 py-1.5 border rounded bg-background font-mono"
                          placeholder="*/5 * * * *"
                          value={config.cron || ''}
                          onChange={e => updateConfig('cron', e.target.value)}
                       />
                       <p className="text-[10px] text-muted-foreground tracking-tight italic">Example: <code>0 8 * * *</code> (Every day at 8:00 AM)</p>
                    </div>
                  </div>
               )}
             </div>
          )}

          {/* ================= HTTP 请求配置 (Action) ================= */}
          {type === 'http_request' && (
            <>
              <div className="space-y-3">
                <label className="text-xs font-semibold flex items-center gap-1">
                  <Globe className="w-3 h-3 text-blue-500" /> Endpoint Settings
                </label>
                <div className="flex gap-2">
                  <select 
                    className="w-24 text-xs border rounded px-1 bg-background font-bold"
                    value={config.method || 'GET'}
                    onChange={(e) => updateConfig('method', e.target.value)}
                  >
                    <option>GET</option>
                    <option>POST</option>
                    <option>PUT</option>
                    <option>DELETE</option>
                  </select>
                  <div className="flex-1">
                    <MonacoInput 
                      value={config.url || ''} 
                      onChange={(val) => updateConfig('url', val)} 
                      contextData={executionContext} // ✅ 动态补全注入
                      language="expr"
                    />
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-semibold">Headers</label>
                <KeyValueList 
                  items={config.headers || {}} 
                  onChange={(val) => updateConfig('headers', val)} 
                />
              </div>

              <div className="space-y-2">
                <label className="text-xs font-semibold">Request Body</label>
                <MonacoInput 
                  value={config.body || ''} 
                  onChange={(val) => updateConfig('body', val)}
                  height="140px"
                  language="json"
                  contextData={executionContext}
                />
              </div>
            </>
          )}

          {/* ================= 逻辑判断 (Condition) ================= */}
          {type === 'condition' && (
            <div className="space-y-4">
               <div className="bg-purple-50 text-purple-700 p-3 rounded text-[11px] flex gap-2 border border-purple-100">
                  <AlertTriangle className="w-4 h-4 shrink-0 text-purple-500" />
                  <p>Expression must evaluate to <code>true</code> or <code>false</code>. Flow will branch based on result.</p>
               </div>
               <div className="space-y-2">
                 <label className="text-xs font-semibold">Expr Logic</label>
                 <MonacoInput 
                    value={config.expression || ''}
                    onChange={(val) => updateConfig('expression', val)}
                    height="100px"
                    language="expr" // ✅ 使用 expr 语法高亮
                    contextData={executionContext}
                 />
               </div>
            </div>
          )}

          {/* ================= 数据处理 (Expression Eval) ================= */}
          {type === 'expression' && (
            <div className="space-y-4">
               <div className="bg-indigo-50 text-indigo-700 p-3 rounded text-[11px] flex gap-2">
                  <Calculator className="w-4 h-4 shrink-0 text-indigo-500" />
                  <p>Transform data using expr functions. Result is stored in <code>steps.ID.output</code>.</p>
               </div>
               <div className="space-y-2">
                 <label className="text-xs font-semibold">Evaluation Script</label>
                 <MonacoInput 
                    value={config.expression || ''}
                    onChange={(val) => updateConfig('expression', val)}
                    height="120px"
                    language="expr" // ✅ 使用 expr 语法高亮
                    contextData={executionContext}
                 />
               </div>
            </div>
          )}

          {/* ================= 邮件发送配置 (Send Email) ================= */}
          {type === 'send_email' && (
            <>
               <div className="space-y-3 p-3 border rounded bg-slate-50/50">
                 <h4 className="text-[10px] font-bold uppercase text-muted-foreground flex items-center gap-1 mb-2">
                   <Shield className="w-3 h-3" /> SMTP Authentication
                 </h4>
                 
                 <div className="grid grid-cols-3 gap-2">
                    <div className="col-span-2 space-y-1">
                        <label className="text-[10px] font-medium text-muted-foreground">SMTP Host</label>
                        <input className="w-full text-xs px-2 py-1.5 border rounded bg-background" 
                            placeholder="smtp.office365.com"
                            value={config.host || ''}
                            onChange={e => updateConfig('host', e.target.value)}
                        />
                    </div>
                    <div className="space-y-1">
                        <label className="text-[10px] font-medium text-muted-foreground">Port</label>
                        <input className="w-full text-xs px-2 py-1.5 border rounded bg-background" 
                            placeholder="587"
                            type="number"
                            value={config.port || ''}
                            onChange={e => updateConfig('port', Number(e.target.value))}
                        />
                    </div>
                 </div>

                 <div className="grid grid-cols-2 gap-2 mt-2">
                    <div className="space-y-1">
                        <label className="text-[10px] font-medium text-muted-foreground">Username</label>
                        <input className="w-full text-xs px-2 py-1.5 border rounded bg-background" 
                            value={config.username || ''}
                            onChange={e => updateConfig('username', e.target.value)}
                        />
                    </div>
                    <div className="space-y-1">
                        <label className="text-[10px] font-medium text-muted-foreground">Password</label>
                        <input className="w-full text-xs px-2 py-1.5 border rounded bg-background font-mono" 
                            type="password"
                            placeholder="••••••••"
                            value={config.password || ''}
                            onChange={e => updateConfig('password', e.target.value)}
                        />
                    </div>
                 </div>
               </div>

               <div className="space-y-3 pt-2">
                 <div className="space-y-1.5">
                    <label className="text-xs font-semibold flex items-center gap-1">
                      <Mail className="w-3 h-3" /> Recipients (To)
                    </label>
                    <MonacoInput 
                        value={config.to || ''}
                        onChange={(val) => updateConfig('to', val)}
                        contextData={executionContext}
                        language="expr"
                    />
                 </div>
                 
                 <div className="space-y-1.5">
                    <label className="text-xs font-semibold text-muted-foreground">Subject</label>
                    <MonacoInput 
                        value={config.subject || ''}
                        onChange={(val) => updateConfig('subject', val)}
                        contextData={executionContext}
                        language="expr"
                    />
                 </div>
                 
                 <div className="space-y-1.5">
                    <label className="text-xs font-semibold text-muted-foreground">Content Body</label>
                    <MonacoInput 
                        value={config.content || ''}
                        onChange={(val) => updateConfig('content', val)}
                        height="200px"
                        language="expr"
                        contextData={executionContext}
                    />
                 </div>
               </div>
            </>
          )}

        </div>
      </ScrollArea>
    </div>
  );
}