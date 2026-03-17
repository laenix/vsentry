import { useState, useEffect, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { toast } from 'sonner';
import { 
  Save, Play, ArrowLeft, Plus, Trash2, Copy, Download, 
  Zap, Globe, Mail, Shield, Server, AlertTriangle, FileCode
} from 'lucide-react';
import { automationService } from '@/services/automation';

// ===== 类型定义 =====

interface TriggerConfig {
  source: string;
  severity: string;
  conditions: string[];
}

interface ActionConfig {
  name: string;
  type: string;
  config: Record<string, any>;
}

interface PlaybookFormData {
  name: string;
  description: string;
  enabled: boolean;
  trigger: TriggerConfig;
  actions: ActionConfig[];
}

const TRIGGER_SOURCES = [
  { value: 'falco', label: 'Falco', icon: '🛡️' },
  { value: 'tetragon', label: 'Tetragon', icon: '🔵' },
  { value: 'manual', label: 'Manual', icon: '👤' },
  { value: 'webhook', label: 'Webhook', icon: '🌐' },
];

const ACTION_TYPES = [
  { value: 'webhook', label: 'HTTP Request', icon: Globe },
  { value: 'email', label: 'Send Email', icon: Mail },
  { value: 'kubernetes', label: 'Kubernetes Action', icon: Server },
  { value: 'forensics', label: 'Forensics Capture', icon: Shield },
  { value: 'expression', label: 'Expression', icon: FileCode },
  { value: 'condition', label: 'Condition Branch', icon: AlertTriangle },
];

const SEVERITY_LEVELS = [
  { value: 'critical', label: 'Critical', color: 'text-red-500' },
  { value: 'high', label: 'High', color: 'text-orange-500' },
  { value: 'medium', label: 'Medium', color: 'text-yellow-500' },
  { value: 'low', label: 'Low', color: 'text-blue-500' },
];

// ===== 默认数据 =====
const defaultFormData: PlaybookFormData = {
  name: 'new-playbook',
  description: '',
  enabled: true,
  trigger: {
    source: 'falco',
    severity: 'critical',
    conditions: [''],
  },
  actions: [
    {
      name: 'notify-team',
      type: 'webhook',
      config: {
        url: '',
        method: 'POST',
        headers: {},
        body: '',
      },
    },
  ],
};

// ===== YAML 生成器 =====
function generateYAML(data: PlaybookFormData): string {
  const lines: string[] = [];
  
  lines.push('apiVersion: vsentry.io/v1');
  lines.push('kind: Playbook');
  lines.push('metadata:');
  lines.push(`  name: ${data.name}`);
  lines.push('  labels:');
  lines.push('    category: incident-response');
  lines.push('');
  lines.push('spec:');
  lines.push(`  enabled: ${data.enabled}`);
  
  if (data.description) {
    lines.push('  description: |');
    data.description.split('\n').forEach(line => {
      lines.push(`    ${line}`);
    });
  }
  
  lines.push('  trigger:');
  lines.push(`    source: ${data.trigger.source}`);
  if (data.trigger.severity) {
    lines.push(`    severity: ${data.trigger.severity}`);
  }
  if (data.trigger.conditions.filter(c => c.trim()).length > 0) {
    lines.push('    conditions:');
    data.trigger.conditions.filter(c => c.trim()).forEach(cond => {
      lines.push(`      - "${cond}"`);
    });
  }
  
  lines.push('  actions:');
  data.actions.forEach((action, idx) => {
    lines.push(`    - name: ${action.name}`);
    lines.push(`      type: ${action.type}`);
    lines.push('      config:');
    
    // 根据类型输出不同配置
    if (action.type === 'webhook') {
      if (action.config.url) lines.push(`        url: ${action.config.url}`);
      if (action.config.method) lines.push(`        method: ${action.config.method}`);
      if (action.config.body) {
        lines.push('        body: |');
        action.config.body.split('\n').forEach(line => {
          lines.push(`          ${line}`);
        });
      }
    } else if (action.type === 'email') {
      if (action.config.to) lines.push(`        to: ${action.config.to}`);
      if (action.config.subject) lines.push(`        subject: ${action.config.subject}`);
      if (action.config.content) {
        lines.push('        content: |');
        action.config.content.split('\n').forEach(line => {
          lines.push(`          ${line}`);
        });
      }
    } else if (action.type === 'kubernetes') {
      if (action.config.action) lines.push(`        action: ${action.config.action}`);
      if (action.config.kind) lines.push(`        kind: ${action.config.kind}`);
      if (action.config.selector) lines.push(`        selector: ${action.config.selector}`);
    } else if (action.type === 'forensics') {
      if (action.config.capture) lines.push(`        capture: ${action.config.capture}`);
      if (action.config.timeout) lines.push(`        timeout: ${action.config.timeout}`);
      if (action.config.storage) lines.push(`        storage: ${action.config.storage}`);
    } else if (action.type === 'expression') {
      if (action.config.expression) {
        lines.push('        expression: |');
        action.config.expression.split('\n').forEach(line => {
          lines.push(`          ${line}`);
        });
      }
    }
  });
  
  return lines.join('\n');
}

// ===== 主组件 =====
export default function CRDEditorPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<PlaybookFormData>(defaultFormData);
  const [yamlPreview, setYamlPreview] = useState('');
  const [activeTab, setActiveTab] = useState('form');
  const [importedYAML, setImportedYAML] = useState('');

  // 生成 YAML 预览
  useEffect(() => {
    setYamlPreview(generateYAML(formData));
  }, [formData]);

  // 加载现有 Playbook
  useEffect(() => {
    if (id && id !== 'new') {
      loadPlaybook(id);
    }
  }, [id]);

  const loadPlaybook = async (playbookId: string) => {
    setLoading(true);
    try {
      const res = await automationService.getDetail(playbookId);
      if (res.code === 200 && res.data) {
        // 从现有数据转换
        const def = typeof res.data.definition === 'string' 
          ? JSON.parse(res.data.definition) 
          : res.data.definition;
        
        setFormData({
          name: res.data.name,
          description: res.data.description || '',
          enabled: res.data.is_active,
          trigger: {
            source: res.data.trigger_type || 'manual',
            severity: def?.trigger?.severity || 'critical',
            conditions: def?.trigger?.conditions || [''],
          },
          actions: def?.actions || [],
        });
      }
    } catch (error) {
      toast.error('Failed to load playbook');
    } finally {
      setLoading(false);
    }
  };

  // 保存 Playbook
  const handleSave = async () => {
    setLoading(true);
    try {
      const payload = {
        name: formData.name,
        description: formData.description,
        is_active: formData.enabled,
        trigger_type: formData.trigger.source,
        definition: {
          trigger: formData.trigger,
          actions: formData.actions,
        },
      };

      let res;
      if (id === 'new') {
        res = await automationService.create(payload);
      } else {
        res = await automationService.update(id, payload);
      }

      if (res.code === 200) {
        toast.success('Playbook saved successfully');
        navigate('/automation');
      } else {
        toast.error(res.msg || 'Failed to save');
      }
    } catch (error) {
      toast.error('Failed to save playbook');
    } finally {
      setLoading(false);
    }
  };

  // 导入 YAML
  const handleImportYAML = () => {
    if (!importedYAML.trim()) {
      toast.error('Please paste YAML content');
      return;
    }
    try {
      // 简单解析 (实际项目中应使用 yaml 库)
      const lines = importedYAML.split('\n');
      let inActions = false;
      let currentAction: any = null;
      const newData = { ...formData };

      lines.forEach(line => {
        const trimmed = line.trim();
        
        // metadata.name
        if (trimmed.startsWith('name:')) {
          newData.name = trimmed.replace('name:', '').trim();
        }
        // description
        else if (trimmed.startsWith('description:') || (trimmed === 'description: |')) {
          // 简化处理
        }
        // enabled
        else if (trimmed.startsWith('enabled:')) {
          newData.enabled = trimmed.includes('true');
        }
        // trigger.source
        else if (trimmed.startsWith('source:') && !trimmed.startsWith('      - "')) {
          if (!inActions) {
            newData.trigger.source = trimmed.replace('source:', '').trim();
          }
        }
        // trigger.severity
        else if (trimmed.startsWith('severity:')) {
          newData.trigger.severity = trimmed.replace('severity:', '').trim();
        }
        // conditions
        else if (trimmed.startsWith('- "')) {
          const cond = trimmed.replace('- "', '').replace('"', '');
          if (!inActions) {
            if (!newData.trigger.conditions.includes(cond)) {
              newData.trigger.conditions.push(cond);
            }
          }
        }
        // actions
        else if (trimmed === 'actions:') {
          inActions = true;
          newData.actions = [];
        }
        else if (trimmed.startsWith('- name:')) {
          if (currentAction) {
            newData.actions.push(currentAction);
          }
          currentAction = {
            name: trimmed.replace('- name:', '').trim(),
            type: 'webhook',
            config: {},
          };
        }
        else if (currentAction && trimmed.startsWith('type:')) {
          currentAction.type = trimmed.replace('type:', '').trim();
        }
        else if (currentAction && trimmed.startsWith('url:')) {
          currentAction.config.url = trimmed.replace('url:', '').trim();
        }
      });

      if (currentAction) {
        newData.actions.push(currentAction);
      }

      // 清理空条件
      newData.trigger.conditions = newData.trigger.conditions.filter(c => c.trim());
      if (newData.trigger.conditions.length === 0) {
        newData.trigger.conditions = [''];
      }

      setFormData(newData);
      setActiveTab('form');
      toast.success('YAML imported successfully');
    } catch (error) {
      toast.error('Failed to parse YAML');
    }
  };

  // 复制 YAML
  const handleCopyYAML = () => {
    navigator.clipboard.writeText(yamlPreview);
    toast.success('YAML copied to clipboard');
  };

  // 下载 YAML 文件
  const handleDownloadYAML = () => {
    const blob = new Blob([yamlPreview], { type: 'text/yaml' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${formData.name}.yaml`;
    a.click();
    URL.revokeObjectURL(url);
  };

  // 更新 Trigger
  const updateTrigger = (field: string, value: any) => {
    setFormData(prev => ({
      ...prev,
      trigger: { ...prev.trigger, [field]: value },
    }));
  };

  // 更新 Action
  const updateAction = (index: number, field: string, value: any) => {
    setFormData(prev => {
      const actions = [...prev.actions];
      if (field === 'config') {
        actions[index] = { ...actions[index], config: { ...actions[index].config, ...value } };
      } else {
        actions[index] = { ...actions[index], [field]: value };
      }
      return { ...prev, actions };
    });
  };

  // 添加 Action
  const addAction = () => {
    setFormData(prev => ({
      ...prev,
      actions: [
        ...prev.actions,
        {
          name: `action-${prev.actions.length + 1}`,
          type: 'webhook',
          config: { url: '', method: 'POST' },
        },
      ],
    }));
  };

  // 删除 Action
  const removeAction = (index: number) => {
    setFormData(prev => ({
      ...prev,
      actions: prev.actions.filter((_, i) => i !== index),
    }));
  };

  // 添加 Condition
  const addCondition = () => {
    setFormData(prev => ({
      ...prev,
      trigger: {
        ...prev.trigger,
        conditions: [...prev.trigger.conditions, ''],
      },
    }));
  };

  // 删除 Condition
  const removeCondition = (index: number) => {
    setFormData(prev => ({
      ...prev,
      trigger: {
        ...prev.trigger,
        conditions: prev.trigger.conditions.filter((_, i) => i !== index),
      },
    }));
  };

  return (
    <div className="h-screen flex flex-col bg-background">
      {/* Header */}
      <div className="flex items-center justify-between px-6 py-4 border-b bg-card">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={() => navigate('/automation')}>
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <div>
            <h1 className="text-xl font-bold">Playbook CRD Editor</h1>
            <p className="text-sm text-muted-foreground">Declarative YAML configuration</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleCopyYAML}>
            <Copy className="w-4 h-4 mr-2" />
            Copy YAML
          </Button>
          <Button variant="outline" size="sm" onClick={handleDownloadYAML}>
            <Download className="w-4 h-4 mr-2" />
            Download
          </Button>
          <Button size="sm" onClick={handleSave} disabled={loading}>
            <Save className="w-4 h-4 mr-2" />
            {loading ? 'Saving...' : 'Save'}
          </Button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left Panel - Form */}
        <div className="w-1/2 border-r flex flex-col">
          <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
            <TabsList className="mx-6 mt-4">
              <TabsTrigger value="form">Form Editor</TabsTrigger>
              <TabsTrigger value="yaml">Import YAML</TabsTrigger>
            </TabsList>

            <TabsContent value="form" className="flex-1 m-0">
              <ScrollArea className="h-full p-6">
                <div className="space-y-6">
                  {/* Basic Info */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">Basic Information</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div className="grid gap-2">
                        <Label>Playbook Name</Label>
                        <Input 
                          value={formData.name}
                          onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                          placeholder="e.g., detect-and-isolate-threat"
                        />
                      </div>
                      <div className="grid gap-2">
                        <Label>Description</Label>
                        <Textarea 
                          value={formData.description}
                          onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                          placeholder="Describe what this playbook does..."
                          rows={3}
                        />
                      </div>
                      <div className="flex items-center gap-2">
                        <input 
                          type="checkbox" 
                          id="enabled"
                          checked={formData.enabled}
                          onChange={(e) => setFormData(prev => ({ ...prev, enabled: e.target.checked }))}
                          className="rounded"
                        />
                        <Label htmlFor="enabled" className="font-normal">Enabled</Label>
                      </div>
                    </CardContent>
                  </Card>

                  {/* Trigger */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base flex items-center gap-2">
                        <Zap className="w-4 h-4 text-amber-500" />
                        Trigger Configuration
                      </CardTitle>
                      <CardDescription>Define when this playbook should be executed</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div className="grid grid-cols-2 gap-4">
                        <div className="grid gap-2">
                          <Label>Source</Label>
                          <Select 
                            value={formData.trigger.source}
                            onValueChange={(v) => updateTrigger('source', v)}
                          >
                            <SelectTrigger>
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              {TRIGGER_SOURCES.map(s => (
                                <SelectItem key={s.value} value={s.value}>
                                  <span className="mr-2">{s.icon}</span>
                                  {s.label}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>
                        <div className="grid gap-2">
                          <Label>Severity</Label>
                          <Select 
                            value={formData.trigger.severity}
                            onValueChange={(v) => updateTrigger('severity', v)}
                          >
                            <SelectTrigger>
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              {SEVERITY_LEVELS.map(s => (
                                <SelectItem key={s.value} value={s.value}>
                                  <span className={s.color}>{s.label}</span>
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>
                      </div>
                      
                      <div className="grid gap-2">
                        <Label>Conditions (LogSQL)</Label>
                        {formData.trigger.conditions.map((cond, idx) => (
                          <div key={idx} className="flex gap-2">
                            <Input 
                              value={cond}
                              onChange={(e) => {
                                const newConds = [...formData.trigger.conditions];
                                newConds[idx] = e.target.value;
                                updateTrigger('conditions', newConds);
                              }}
                              placeholder='e.g., severity = critical'
                            />
                            {formData.trigger.conditions.length > 1 && (
                              <Button variant="ghost" size="icon" onClick={() => removeCondition(idx)}>
                                <Trash2 className="w-4 h-4" />
                              </Button>
                            )}
                          </div>
                        ))}
                        <Button variant="outline" size="sm" onClick={addCondition}>
                          <Plus className="w-4 h-4 mr-2" />
                          Add Condition
                        </Button>
                      </div>
                    </CardContent>
                  </Card>

                  {/* Actions */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base flex items-center gap-2">
                        <Play className="w-4 h-4 text-green-500" />
                        Response Actions
                      </CardTitle>
                      <CardDescription>Define actions to execute when trigger fires</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      {formData.actions.map((action, idx) => (
                        <div key={idx} className="border rounded-lg p-4 space-y-4">
                          <div className="flex items-center justify-between">
                            <span className="font-medium">Action #{idx + 1}</span>
                            <Button variant="ghost" size="icon" onClick={() => removeAction(idx)}>
                              <Trash2 className="w-4 h-4 text-red-500" />
                            </Button>
                          </div>
                          
                          <div className="grid grid-cols-2 gap-4">
                            <div className="grid gap-2">
                              <Label>Name</Label>
                              <Input 
                                value={action.name}
                                onChange={(e) => updateAction(idx, 'name', e.target.value)}
                              />
                            </div>
                            <div className="grid gap-2">
                              <Label>Type</Label>
                              <Select 
                                value={action.type}
                                onValueChange={(v) => updateAction(idx, 'type', v)}
                              >
                                <SelectTrigger>
                                  <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                  {ACTION_TYPES.map(t => (
                                    <SelectItem key={t.value} value={t.value}>
                                      {t.label}
                                    </SelectItem>
                                  ))}
                                </SelectContent>
                              </Select>
                            </div>
                          </div>

                          {/* Type-specific config */}
                          {action.type === 'webhook' && (
                            <div className="grid gap-2">
                              <Label>URL</Label>
                              <Input 
                                value={action.config.url || ''}
                                onChange={(e) => updateAction(idx, 'config', { url: e.target.value })}
                                placeholder="https://hooks.slack.com/services/xxx"
                              />
                              <Label>Method</Label>
                              <Select 
                                value={action.config.method || 'POST'}
                                onValueChange={(v) => updateAction(idx, 'config', { method: v })}
                              >
                                <SelectTrigger>
                                  <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                  <SelectItem value="GET">GET</SelectItem>
                                  <SelectItem value="POST">POST</SelectItem>
                                  <SelectItem value="PUT">PUT</SelectItem>
                                  <SelectItem value="DELETE">DELETE</SelectItem>
                                </SelectContent>
                              </Select>
                              <Label>Body (Template)</Label>
                              <Textarea 
                                value={action.config.body || ''}
                                onChange={(e) => updateAction(idx, 'config', { body: e.target.value })}
                                placeholder='{"text": "Alert: {{ .incident.name }}"}'
                                rows={3}
                              />
                            </div>
                          )}

                          {action.type === 'email' && (
                            <div className="grid gap-2">
                              <Label>To</Label>
                              <Input 
                                value={action.config.to || ''}
                                onChange={(e) => updateAction(idx, 'config', { to: e.target.value })}
                                placeholder="security@company.com"
                              />
                              <Label>Subject</Label>
                              <Input 
                                value={action.config.subject || ''}
                                onChange={(e) => updateAction(idx, 'config', { subject: e.target.value })}
                                placeholder="Security Alert"
                              />
                            </div>
                          )}

                          {action.type === 'kubernetes' && (
                            <div className="grid grid-cols-2 gap-2">
                              <div className="grid gap-2">
                                <Label>Action</Label>
                                <Select 
                                  value={action.config.action || 'patch'}
                                  onValueChange={(v) => updateAction(idx, 'config', { action: v })}
                                >
                                  <SelectTrigger>
                                    <SelectValue />
                                  </SelectTrigger>
                                  <SelectContent>
                                    <SelectItem value="patch">Patch</SelectItem>
                                    <SelectItem value="delete">Delete</SelectItem>
                                    <SelectItem value="evict">Evict</SelectItem>
                                  </SelectContent>
                                </Select>
                              </div>
                              <div className="grid gap-2">
                                <Label>Kind</Label>
                                <Select 
                                  value={action.config.kind || 'Pod'}
                                  onValueChange={(v) => updateAction(idx, 'config', { kind: v })}
                                >
                                  <SelectTrigger>
                                    <SelectValue />
                                  </SelectTrigger>
                                  <SelectContent>
                                    <SelectItem value="Pod">Pod</SelectItem>
                                    <SelectItem value="Service">Service</SelectItem>
                                    <SelectItem value="NetworkPolicy">NetworkPolicy</SelectItem>
                                  </SelectContent>
                                </Select>
                              </div>
                              <div className="grid gap-2 col-span-2">
                                <Label>Selector</Label>
                                <Input 
                                  value={action.config.selector || ''}
                                  onChange={(e) => updateAction(idx, 'config', { selector: e.target.value })}
                                  placeholder='{{ .incident.pod }}'
                                />
                              </div>
                            </div>
                          )}

                          {action.type === 'forensics' && (
                            <div className="grid grid-cols-2 gap-2">
                              <div className="grid gap-2">
                                <Label>Capture</Label>
                                <Select 
                                  value={action.config.capture || 'memory'}
                                  onValueChange={(v) => updateAction(idx, 'config', { capture: v })}
                                >
                                  <SelectTrigger>
                                    <SelectValue />
                                  </SelectTrigger>
                                  <SelectContent>
                                    <SelectItem value="memory">Memory</SelectItem>
                                    <SelectItem value="filesystem">Filesystem</SelectItem>
                                    <SelectItem value="all">All</SelectItem>
                                  </SelectContent>
                                </Select>
                              </div>
                              <div className="grid gap-2">
                                <Label>Timeout</Label>
                                <Input 
                                  value={action.config.timeout || '30s'}
                                  onChange={(e) => updateAction(idx, 'config', { timeout: e.target.value })}
                                  placeholder="30s"
                                />
                              </div>
                              <div className="grid gap-2 col-span-2">
                                <Label>Storage</Label>
                                <Input 
                                  value={action.config.storage || ''}
                                  onChange={(e) => updateAction(idx, 'config', { storage: e.target.value })}
                                  placeholder="s3://bucket/path"
                                />
                              </div>
                            </div>
                          )}

                          {action.type === 'expression' && (
                            <div className="grid gap-2">
                              <Label>Expression</Label>
                              <Textarea 
                                value={action.config.expression || ''}
                                onChange={(e) => updateAction(idx, 'config', { expression: e.target.value })}
                                placeholder="query_logs('severity >= critical', last_1h)"
                                rows={3}
                              />
                            </div>
                          )}
                        </div>
                      ))}
                      
                      <Button variant="outline" onClick={addAction}>
                        <Plus className="w-4 h-4 mr-2" />
                        Add Action
                      </Button>
                    </CardContent>
                  </Card>
                </div>
              </ScrollArea>
            </TabsContent>

            {/* Import YAML Tab */}
            <TabsContent value="yaml" className="flex-1 m-0">
              <div className="p-6 h-full">
                <Textarea 
                  value={importedYAML}
                  onChange={(e) => setImportedYAML(e.target.value)}
                  placeholder="Paste your Playbook CRD YAML here..."
                  className="h-[calc(100%-60px)] font-mono text-sm"
                />
                <Button className="mt-4" onClick={handleImportYAML}>
                  Import YAML
                </Button>
              </div>
            </TabsContent>
          </Tabs>
        </div>

        {/* Right Panel - YAML Preview */}
        <div className="w-1/2 flex flex-col bg-muted/30">
          <div className="flex items-center justify-between px-6 py-4 border-b">
            <h2 className="font-semibold flex items-center gap-2">
              <FileCode className="w-4 h-4" />
              YAML Preview
            </h2>
            <span className="text-xs text-muted-foreground">
              Generated from form
            </span>
          </div>
          <ScrollArea className="flex-1">
            <pre className="p-6 text-sm font-mono text-muted-foreground whitespace-pre-wrap">
              {yamlPreview}
            </pre>
          </ScrollArea>
        </div>
      </div>
    </div>
  );
}
