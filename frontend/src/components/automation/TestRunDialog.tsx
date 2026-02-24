import { useEffect, useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { incidentService } from "@/services/incidents";
import type { Incident } from "@/services/incidents";
import { MonacoInput } from "./MonacoInput"; // 复用你的 MonacoInput 做 Mock 数据
import { Loader2, Play, Info } from "lucide-react";
import { Badge } from "@/components/ui/badge";

interface TestRunDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onRun: (data: { incident_id?: number; mock_data?: any }) => void;
  triggerConfig: any; // ✅ 接收当前剧本的 Trigger 配置
}

export function TestRunDialog({ open, onOpenChange, onRun, triggerConfig }: TestRunDialogProps) {
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedId, setSelectedId] = useState<string>("");
  const [mockData, setMockData] = useState<string>(triggerConfig?.payload_template || '{\n  "test": "data"\n}');

  const triggerType = triggerConfig?.trigger_type || 'incident_created';

  useEffect(() => {
    // 只有在事件触发模式下才去拉取列表
    if (open && triggerType === 'incident_created') {
      setLoading(true);
      incidentService.list('new')
        .then(res => setIncidents(res.data || []))
        .finally(() => setLoading(false));
    }
  }, [open, triggerType]);

  const handleExecute = () => {
    if (triggerType === 'incident_created') {
      if (!selectedId) return;
      onRun({ incident_id: Number(selectedId) });
    } else {
      // 手动/定时模式：发送 JSON Mock 数据
      try {
        const parsed = JSON.parse(mockData);
        onRun({ mock_data: parsed });
      } catch (e) {
        alert("Invalid JSON in mock data");
        return;
      }
    }
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Play className="w-5 h-5 text-emerald-600" />
            Execute Test: {triggerType.replace('_', ' ').toUpperCase()}
          </DialogTitle>
          <DialogDescription>
            {triggerType === 'incident_created' 
              ? "Select an incident to provide real context for this test." 
              : "Define the input data (JSON) to simulate the trigger."}
          </DialogDescription>
        </DialogHeader>
        
        <div className="py-4 space-y-4">
          {triggerType === 'incident_created' ? (
            <div className="space-y-2">
              <label className="text-xs font-medium uppercase text-muted-foreground">Select Incident Context</label>
              <Select value={selectedId} onValueChange={setSelectedId} disabled={loading}>
                <SelectTrigger>
                  <SelectValue placeholder={loading ? "Loading..." : "Select an incident..."} />
                </SelectTrigger>
                <SelectContent>
                  {incidents.map((inc) => (
                    <SelectItem key={inc.ID} value={String(inc.ID)}>
                      #{inc.ID} - {inc.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          ) : (
            <div className="space-y-2">
              <label className="text-xs font-medium uppercase text-muted-foreground">Mock Input Context (JSON)</label>
              <MonacoInput 
                value={mockData}
                onChange={setMockData}
                height="200px"
                language="json"
              />
            </div>
          )}
          
          <div className="bg-blue-50 text-blue-700 text-[11px] p-3 rounded flex gap-2">
             <Info className="w-4 h-4 shrink-0" />
             <p>Test Run will simulate the execution based on the logic currently on the canvas.</p>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button onClick={handleExecute} disabled={(triggerType === 'incident_created' && !selectedId) || loading} className="bg-emerald-600">
            Run Test
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}