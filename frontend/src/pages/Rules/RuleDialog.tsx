import React, { useEffect, useState } from "react"
import { 
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription 
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { 
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue 
} from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { LogSQLEditor } from "@/components/editor/LogSQLEditor"
import { ruleService } from "@/services/rules"
import type { DetectionRule } from "@/services/rules"
import { runVLQuery } from "@/lib/api/vl-client" // ✅ 引入日志查询 API
import { toast } from "sonner"
import { Loader2, Play, Info, Clock, CheckCircle, AlertTriangle, Terminal } from "lucide-react"
import { cn } from "@/lib/utils"
import { ScrollArea } from "@/components/ui/scroll-area"

interface RuleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialData: DetectionRule | null
  onSuccess: () => void
}

// 预设 Cron
const CRON_PRESETS = [
  { label: "Every 1 Minute", value: "0 */1 * * * *" },
  { label: "Every 5 Minutes", value: "0 */5 * * * *" },
  { label: "Every 15 Minutes", value: "0 */15 * * * *" },
  { label: "Every 30 Minutes", value: "0 */30 * * * *" },
  { label: "Every 1 Hour", value: "0 0 */1 * * * *" },
  { label: "Daily (Midnight)", value: "0 0 0 * * *" },
]

const SEVERITY_OPTIONS = [
  { value: "critical", label: "Critical", color: "text-red-600" },
  { value: "high", label: "High", color: "text-orange-500" },
  { value: "medium", label: "Medium", color: "text-yellow-500" },
  { value: "low", label: "Low", color: "text-blue-500" },
]

export function RuleDialog({ open, onOpenChange, initialData, onSuccess }: RuleDialogProps) {
  const isEditMode = !!initialData
  const [loading, setLoading] = useState(false)
  const [testing, setTesting] = useState(false) // ✅ 试运行 Loading 状态
  const [testResult, setTestResult] = useState<{ success: boolean; msg: string; sample?: string } | null>(null) // ✅ 试运行结果

  // 表单状态
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [query, setQuery] = useState("")
  const [severity, setSeverity] = useState("medium")
  const [enabled, setEnabled] = useState(true)
  const [interval, setInterval] = useState("0 */5 * * * *") 

  useEffect(() => {
    if (open) {
      setTestResult(null) // 重置测试结果
      if (initialData) {
        setName(initialData.name)
        setDescription(initialData.description || "")
        setQuery(initialData.query)
        setSeverity(initialData.severity)
        setInterval(initialData.interval)
        setEnabled(initialData.enabled)
      } else {
        setName("")
        setDescription("")
        setQuery("_time:5m AND _msg:error")
        setSeverity("medium")
        setInterval("0 */5 * * * *")
        setEnabled(true)
      }
    }
  }, [open, initialData])

  // ✅ 真正的试运行逻辑
  const handleTestRun = async () => {
    if (!query.trim()) {
      toast.error("Please enter a query first");
      return;
    }

    setTesting(true);
    setTestResult(null);

    try {
      // 这里的 "10" 会被传递给 VictoriaLogs 的 /select/logsql?limit=10 接口
      const res = await runVLQuery(query, "10");
      
      if (res.length > 0) {
        setTestResult({
          success: true,
          msg: `Query valid! Matched ${res.length} logs (Limit applied: 10).`,
          sample: JSON.stringify(res[0], null, 2) // 取第一条做样本
        });
        toast.success("Test run successful");
      } else {
        setTestResult({
          success: true,
          msg: "Query is valid (no syntax errors), but 0 logs matched.",
        });
        toast.info("Test run successful: No hits");
      }
    } catch (error: any) {
      console.error(error);
      setTestResult({
        success: false,
        msg: `Syntax Error: ${error.message || "Unknown error"}`,
      });
      toast.error("Test run failed");
    } finally {
      setTesting(false);
    }
  };

  const handleSubmit = async () => {
    if (!name.trim()) return toast.error("Rule name is required")
    if (!query.trim()) return toast.error("LogSQL query is required")
    if (!interval.trim()) return toast.error("Schedule is required")

    setLoading(true)
    try {
      const payload = {
        name,
        description,
        query,
        severity,
        interval,
        enabled,
      }

      if (isEditMode && initialData) {
        const id = initialData.ID || initialData.id;
        await ruleService.update({ ...payload, id })
        toast.success("Rule updated")
      } else {
        await ruleService.add(payload)
        toast.success("Rule created")
      }

      onSuccess()
      onOpenChange(false)
    } catch (error: any) {
      console.error(error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] flex flex-col gap-0 p-0 overflow-hidden">
        <DialogHeader className="px-6 py-4 border-b bg-muted/10">
          <DialogTitle>{isEditMode ? "Edit Detection Rule" : "Create New Detection Rule"}</DialogTitle>
          <DialogDescription>
            Configure the logic to automatically detect threats in your logs.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Basic Info */}
          <div className="grid grid-cols-2 gap-6">
            <div className="space-y-3">
              <div className="space-y-2">
                <Label htmlFor="rule-name">Rule Name</Label>
                <Input 
                  id="rule-name" 
                  placeholder="e.g. Brute Force Attempt (SSH)" 
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  autoFocus
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="desc">Description (Optional)</Label>
                <Input 
                  id="desc" 
                  placeholder="Brief description..." 
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="text-xs"
                />
              </div>
            </div>

            <div className="space-y-3">
              <div className="space-y-2">
                <Label>Severity</Label>
                <Select value={severity} onValueChange={setSeverity}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {SEVERITY_OPTIONS.map(opt => (
                      <SelectItem key={opt.value} value={opt.value} className="cursor-pointer">
                        <span className={cn("font-medium", opt.color)}>{opt.label}</span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label className="flex justify-between text-xs">
                  <span>Schedule (Cron)</span>
                </Label>
                <div className="flex gap-2">
                  <Select onValueChange={setInterval}>
                    <SelectTrigger className="w-[140px] text-xs"><SelectValue placeholder="Presets" /></SelectTrigger>
                    <SelectContent>
                      {CRON_PRESETS.map(pre => (
                        <SelectItem key={pre.value} value={pre.value} className="text-xs">{pre.label}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <div className="relative flex-1">
                    <Clock className="absolute left-2 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
                    <Input 
                      value={interval}
                      onChange={(e) => setInterval(e.target.value)}
                      className="pl-8 font-mono text-xs"
                      placeholder="0 */5 * * * *"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Query Editor & Test Run */}
          <div className="space-y-2 flex flex-col flex-1 min-h-[300px]">
            <div className="flex items-center justify-between">
              <Label className="flex items-center gap-2">Detection Logic (LogSQL)</Label>
              {/* ✅ Test Run Button */}
              <Button 
                variant="secondary" 
                size="xs" 
                className="h-6 text-xs gap-1 border hover:bg-accent" 
                onClick={handleTestRun}
                disabled={testing}
              >
                {testing ? <Loader2 className="w-3 h-3 animate-spin" /> : <Play className="w-3 h-3 text-green-600" />}
                {testing ? "Running..." : "Test Run (Limit 10)"}
              </Button>
            </div>
            
            <div className="border rounded-md overflow-hidden flex-1 relative shadow-sm min-h-[250px]">
              <div className="absolute inset-0">
                <LogSQLEditor value={query} onChange={(val) => setQuery(val || "")} />
              </div>
            </div>

            {/* ✅ 试运行结果展示区 */}
            {testResult && (
              <div className={cn(
                "rounded-md border p-3 text-xs animate-in fade-in slide-in-from-top-2",
                testResult.success ? "bg-green-50/50 border-green-200" : "bg-red-50/50 border-red-200"
              )}>
                <div className="flex items-center gap-2 font-semibold mb-1">
                  {testResult.success ? <CheckCircle className="w-4 h-4 text-green-600" /> : <AlertTriangle className="w-4 h-4 text-red-600" />}
                  <span className={testResult.success ? "text-green-700" : "text-red-700"}>
                    {testResult.success ? "Test Passed" : "Query Failed"}
                  </span>
                </div>
                <p className="text-muted-foreground ml-6 mb-2">{testResult.msg}</p>
                {testResult.sample && (
                  <div className="ml-6 mt-2">
                    <div className="text-[10px] uppercase font-bold text-muted-foreground mb-1 flex items-center gap-1">
                      <Terminal className="w-3 h-3" /> Sample Log Hit
                    </div>
                    <ScrollArea className="h-32 w-full rounded border bg-background">
                      <pre className="p-2 font-mono text-[10px] leading-relaxed">
                        {testResult.sample}
                      </pre>
                    </ScrollArea>
                  </div>
                )}
              </div>
            )}
            
            {!testResult && (
              <p className="text-[11px] text-muted-foreground flex items-center gap-1.5 bg-blue-50/50 p-2 rounded text-blue-600">
                <Info className="w-3.5 h-3.5" />
                Tip: The scheduler executes this query based on the Cron interval.
              </p>
            )}
          </div>

          <div className="flex items-center justify-between border rounded-lg p-3 bg-muted/10">
            <div className="space-y-0.5">
              <Label className="text-base">Enable Rule</Label>
              <p className="text-xs text-muted-foreground">If disabled, the scheduler will skip this rule.</p>
            </div>
            <Switch checked={enabled} onCheckedChange={setEnabled} />
          </div>
        </div>

        <DialogFooter className="px-6 py-4 border-t bg-muted/10">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>Cancel</Button>
          <Button onClick={handleSubmit} disabled={loading} className="min-w-[100px]">
            {loading ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : null}
            {isEditMode ? "Save Changes" : "Create Rule"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}