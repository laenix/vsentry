import { useEffect, useState } from "react";
import { ruleService, type RuleType } from "@/services/rules";
import type { DetectionRule } from "@/services/rules";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Plus, Trash2, Edit, History, User, Filter } from "lucide-react"; // 新增图标
import { toast } from "sonner";
import { RuleDialog } from "./RuleDialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

// 简单的日期格式化 helper
const formatDate = (dateStr?: string) => {
  if (!dateStr) return "-";
  return new Date(dateStr).toLocaleString('zh-CN', {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
  });
};

export default function RulesPage() {
  const [rules, setRules] = useState<DetectionRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<DetectionRule | null>(null);
  const [typeFilter, setTypeFilter] = useState<RuleType | "all">("all");

  // 根据类型筛选
  const filteredRules = typeFilter === "all" 
    ? rules 
    : rules.filter(r => (r.type || "alert") === typeFilter);

  const fetchRules = async () => {
    try {
      const res = await ruleService.list();
      if (res.code === 200) {
        // 兼容之前的 data.rules 结构逻辑
        let list: DetectionRule[] = [];
        if (Array.isArray(res.data)) list = res.data;
        else if (res.data && Array.isArray(res.data.rules)) list = res.data.rules;

        setRules(list);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchRules(); }, []);

  // 获取真实 ID (兼容大小写)
  const getRuleID = (rule: DetectionRule) => rule.ID || rule.id || 0;

  const toggleRule = async (rule: DetectionRule) => {
    const id = getRuleID(rule);
    if (!id) return;
    try {
      if (rule.enabled) {
        await ruleService.disable(id);
        toast.success(`Rule "${rule.name}" disabled`);
      } else {
        await ruleService.enable(id);
        toast.success(`Rule "${rule.name}" enabled`);
      }
      fetchRules();
    } catch (e) { console.error(e); }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure?")) return;
    try {
      await ruleService.delete(id);
      toast.success("Rule deleted");
      fetchRules();
    } catch (e) { console.error(e); }
  };

  return (
    <div className="p-6 h-full flex flex-col">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold">Detection Rules</h1>
          <p className="text-muted-foreground text-sm">Manage SIEM detection logic and alerting policies.</p>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Filter className="w-4 h-4 text-muted-foreground" />
            <Select value={typeFilter} onValueChange={(v) => setTypeFilter(v as RuleType | "all")}>
              <SelectTrigger className="w-[140px] h-9">
                <SelectValue placeholder="All Types" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Types</SelectItem>
                <SelectItem value="alert">报警规则</SelectItem>
                <SelectItem value="forensic">取证规则</SelectItem>
                <SelectItem value="investigation">调查规则</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button onClick={() => { setEditingRule(null); setDialogOpen(true); }}>
            <Plus className="w-4 h-4 mr-2" /> New Rule
          </Button>
        </div>
      </div>

      <div className="border rounded-md bg-card flex-1 overflow-auto">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/5">
              <TableHead className="w-[50px]">ID</TableHead>
              <TableHead className="min-w-[150px]">Rule Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Query (LogSQL)</TableHead>
              <TableHead>Severity</TableHead>
              {/* 新增列 */}
              <TableHead>Version</TableHead>
              <TableHead>Author</TableHead>
              <TableHead>Last Updated</TableHead>
              <TableHead>Enabled</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {Array.isArray(filteredRules) && filteredRules.map((rule) => {
              const id = getRuleID(rule);
              const ruleType = rule.type || "alert";
              return (
                <TableRow key={id}>
                  <TableCell className="font-mono text-xs text-muted-foreground">#{id}</TableCell>
                  <TableCell>
                    <div className="font-medium">{rule.name}</div>
                    {rule.description && <div className="text-[10px] text-muted-foreground truncate max-w-[150px]">{rule.description}</div>}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className={
                      ruleType === 'alert' ? 'border-blue-500 text-blue-500 bg-blue-500/10' :
                        ruleType === 'forensic' ? 'border-purple-500 text-purple-500 bg-purple-500/10' :
                          'border-green-500 text-green-500 bg-green-500/10'
                    }>
                      {ruleType === 'alert' ? '报警' : ruleType === 'forensic' ? '取证' : '调查'}
                    </Badge>
                    {rule.enable_backtrace && ruleType === 'alert' && (
                      <span className="ml-1 text-[10px] text-orange-500" title="Backtrace enabled">↺</span>
                    )}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground truncate max-w-[200px]" title={rule.query}>
                    {rule.query}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className={
                      rule.severity === 'critical' ? 'border-red-500 text-red-500 bg-red-500/10' :
                        rule.severity === 'high' ? 'border-orange-500 text-orange-500 bg-orange-500/10' :
                          'border-blue-500 text-blue-500 bg-blue-500/10'
                    }>
                      {rule.severity}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1 text-xs font-mono bg-muted/20 px-1.5 py-0.5 rounded w-fit">
                      <History className="w-3 h-3 text-muted-foreground" />
                      v{rule.version || 1}
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                      <User className="w-3 h-3" />
                      {rule.author_id ? `User ${rule.author_id}` : "System"}
                    </div>
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground whitespace-nowrap">
                    {formatDate(rule.UpdatedAt || rule.CreatedAt)}
                  </TableCell>
                  <TableCell>
                    <Switch checked={rule.enabled} onCheckedChange={() => toggleRule(rule)} />
                  </TableCell>
                  <TableCell className="text-right space-x-1">
                    <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => { setEditingRule(rule); setDialogOpen(true); }}>
                      <Edit className="w-3.5 h-3.5" />
                    </Button>
                    <Button variant="ghost" size="icon" className="h-8 w-8 text-red-500 hover:text-red-600 hover:bg-red-50" onClick={() => handleDelete(id)}>
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </TableCell>
                </TableRow>
              );
            })}

            {(!Array.isArray(rules) || rules.length === 0) && !loading && (
              <TableRow>
                <TableCell colSpan={9} className="h-32 text-center text-muted-foreground">
                  No detection rules found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {dialogOpen && (
        <RuleDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          initialData={editingRule}
          onSuccess={() => {
            setDialogOpen(false);
            fetchRules();
          }}
        />
      )}
    </div>
  );
}