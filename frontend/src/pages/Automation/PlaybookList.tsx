// src/pages/Automation/PlaybookList.tsx
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Edit2, MoreHorizontal, Zap, Link as LinkIcon, Loader2, ShieldCheck } from "lucide-react";
import { automationService } from "@/services/automation"; //
import type { Playbook } from "@/services/automation";       //
import { ruleService } from "@/services/rules";       //
import type { DetectionRule } from "@/services/rules";        //
import { toast } from "sonner";
import { Checkbox } from "@/components/ui/checkbox";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

interface PlaybookListProps {
  viewMode: 'list' | 'binding';
}

export default function PlaybookList({ viewMode }: PlaybookListProps) {
  const navigate = useNavigate();
  const [data, setData] = useState<Playbook[]>([]);
  const [allRules, setAllRules] = useState<DetectionRule[]>([]); //
  const [loading, setLoading] = useState(true);

  // 绑定模式专用状态
  const [selectedPb, setSelectedPb] = useState<Playbook | null>(null);
  const [boundRuleIds, setBoundRuleIds] = useState<number[]>([]);
  const [syncing, setSyncing] = useState(false);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [pbRes, ruleRes] = await Promise.all([
        automationService.getList(), //
        ruleService.list()           //
      ]);
      setData(pbRes.data || []);
      setAllRules(ruleRes.data || []); //
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  // 当在绑定模式选中剧本时，拉取其绑定的规则
  useEffect(() => {
    if (viewMode === 'binding' && selectedPb) {
      setSyncing(true);
      automationService.getBoundRules(selectedPb.ID) //
        .then(res => setBoundRuleIds((res.data || []).map((r: any) => r.ID || r.id)))
        .finally(() => setSyncing(false));
    }
  }, [selectedPb, viewMode]);

  const handleToggleBinding = async (ruleId: number, checked: boolean) => {
    if (!selectedPb) return;
    const newIds = checked ? [...boundRuleIds, ruleId] : boundRuleIds.filter(id => id !== ruleId);
    try {
      setSyncing(true);
      await automationService.bindRules(selectedPb.ID, newIds); //
      setBoundRuleIds(newIds);
      toast.success("Binding updated");
    } catch (e) {
      toast.error("Failed to update binding");
    } finally {
      setSyncing(false);
    }
  };

  if (loading && data.length === 0) return <div className="flex justify-center p-12"><Loader2 className="animate-spin" /></div>;

  // --- 模式 A: 纯列表展示 ---
  if (viewMode === 'list') {
    return (
      <div className="border rounded-md bg-card shadow-sm">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Playbook Name</TableHead>
              <TableHead>Trigger</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.map((pb) => (
              <TableRow key={pb.ID}>
                <TableCell className="font-medium cursor-pointer" onClick={() => navigate(`/automation/edit/${pb.ID}`)}>
                  <div className="flex flex-col">
                    <span className="text-sm font-semibold">{pb.name}</span>
                    <span className="text-[10px] text-muted-foreground">ID: {pb.ID}</span>
                  </div>
                </TableCell>
                <TableCell className="text-xs capitalize">{pb.trigger_type?.replace('_', ' ') || 'Manual'}</TableCell>
                <TableCell><Badge variant={pb.isActive ? "outline" : "secondary"}>{pb.isActive ? "Active" : "Disabled"}</Badge></TableCell>
                <TableCell className="text-right">
                  <Button variant="ghost" size="icon" onClick={() => navigate(`/automation/edit/${pb.ID}`)}><Edit2 className="w-4 h-4" /></Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    );
  }

  // --- 模式 B: 左右分栏绑定管理 ---
  return (
    <div className="grid grid-cols-12 gap-6 h-[calc(100vh-280px)]">
      {/* 左侧：Playbook 列表 */}
      <Card className="col-span-4 flex flex-col shadow-sm">
        <CardHeader className="py-3 px-4 border-b bg-muted/20">
          <CardTitle className="text-xs font-bold uppercase tracking-widest text-muted-foreground">Orchestrators</CardTitle>
        </CardHeader>
        <ScrollArea className="flex-1 p-2">
          {data.map(pb => (
            <div 
              key={pb.ID}
              onClick={() => setSelectedPb(pb)}
              className={`p-3 mb-1 rounded-md cursor-pointer transition-all ${selectedPb?.ID === pb.ID ? 'bg-primary text-primary-foreground shadow-md' : 'hover:bg-muted'}`}
            >
              <div className="text-xs font-bold leading-none mb-1">{pb.name}</div>
              <div className={`text-[9px] ${selectedPb?.ID === pb.ID ? 'opacity-80' : 'opacity-40'}`}>ID: {pb.ID}</div>
            </div>
          ))}
        </ScrollArea>
      </Card>

      {/* 右侧：Rule 勾选列表 */}
      <Card className="col-span-8 flex flex-col shadow-sm">
        <CardHeader className="py-3 px-4 border-b bg-muted/20 flex flex-row items-center justify-between">
          <CardTitle className="text-xs font-bold uppercase tracking-widest text-muted-foreground">
            {selectedPb ? `Bound Rules for: ${selectedPb.name}` : "Associated Detection Rules"}
          </CardTitle>
          {syncing && <Loader2 className="w-4 h-4 animate-spin text-primary" />}
        </CardHeader>
        <ScrollArea className="flex-1 p-6">
          {!selectedPb ? (
            <div className="h-full flex flex-col items-center justify-center opacity-20 text-center space-y-2">
              <ShieldCheck className="w-12 h-12" />
              <p className="text-xs italic">Select a playbook from the left to manage rule associations.</p>
            </div>
          ) : (
            <div className="space-y-3">
              {allRules.map(rule => (
                <div key={rule.ID} className="flex items-center justify-between p-4 border rounded-xl hover:bg-muted/30 transition-all border-muted/60">
                  <div className="flex items-center gap-4">
                    <Checkbox 
                      id={`rule-${rule.ID}`}
                      checked={boundRuleIds.includes(rule.ID)}
                      onCheckedChange={(checked) => handleToggleBinding(rule.ID, !!checked)}
                      disabled={syncing}
                    />
                    <div className="flex flex-col">
                      <label htmlFor={`rule-${rule.ID}`} className="text-sm font-bold cursor-pointer">{rule.name}</label>
                      <span className="text-[10px] text-muted-foreground">{rule.severity} • {rule.query?.slice(0, 50)}...</span>
                    </div>
                  </div>
                  {boundRuleIds.includes(rule.ID) && <Badge className="bg-emerald-50 text-emerald-700 border-none text-[9px]">Linked</Badge>}
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
      </Card>
    </div>
  );
}