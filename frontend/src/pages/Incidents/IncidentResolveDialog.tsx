import { useState } from "react";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue, SelectGroup, SelectLabel
} from "@/components/ui/select"
import { ShieldCheck, AlertOctagon, HelpCircle, Ban, Loader2, CheckCircle2 } from "lucide-react";

// 定义符合 Sentinel 逻辑的关闭分类
export type ClosingClassification =
  | "TruePositive_Malicious"      // 真实威胁
  | "BenignPositive_Suspicious"   // 正常但可疑 (如压测)
  | "FalsePositive_IncorrectLogic" // 误报 - 逻辑错误
  | "FalsePositive_InaccurateData" // 误报 - 数据源问题
  | "Undetermined";                // 无法确定

interface IncidentResolveDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (classification: string, comment: string) => Promise<void>;
}

export function IncidentResolveDialog({ open, onOpenChange, onConfirm }: IncidentResolveDialogProps) {
  const [classification, setClassification] = useState<ClosingClassification | "">("");
  const [comment, setComment] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    if (!classification) return;
    setLoading(true);
    try {
      await onConfirm(classification, comment); // 触发 index.tsx 中的 API 调用
      onOpenChange(false);
      // 清空表单
      setClassification("");
      setComment("");
    } catch (error) {
      // 错误通常由全局拦截器或父组件处理
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[520px]">
        <DialogHeader>
          <div className="flex items-center gap-2 mb-1">
            <CheckCircle2 className="w-5 h-5 text-emerald-600" />
            <DialogTitle>Resolve Incident</DialogTitle>
          </div>
          <DialogDescription>
            Closing this incident will mark it as resolved. Please provide the classification and any findings.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-6 py-4">
          {/* 1. 分类选择 (Classification) */}
          <div className="space-y-2">
            <Label className="text-xs font-bold uppercase text-muted-foreground">
              Classification <span className="text-red-500">*</span>
            </Label>
            <Select value={classification} onValueChange={(v) => setClassification(v as ClosingClassification)}>
              <SelectTrigger className="h-11">
                <SelectValue placeholder="Select why you are closing this..." />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectLabel className="flex items-center gap-2 text-red-600 px-2 py-1.5 text-[10px] uppercase font-bold">
                    <AlertOctagon className="w-3 h-3" /> True Positive
                  </SelectLabel>
                  <SelectItem value="TruePositive_Malicious">Confirmed Malicious Activity</SelectItem>
                </SelectGroup>

                <SelectGroup>
                  <SelectLabel className="flex items-center gap-2 text-orange-600 px-2 py-1.5 text-[10px] uppercase font-bold mt-2">
                    <ShieldCheck className="w-3 h-3" /> Benign Positive
                  </SelectLabel>
                  <SelectItem value="BenignPositive_Suspicious">Authorized Test / Red Team</SelectItem>
                </SelectGroup>

                <SelectGroup>
                  <SelectLabel className="flex items-center gap-2 text-muted-foreground px-2 py-1.5 text-[10px] uppercase font-bold mt-2">
                    <Ban className="w-3 h-3" /> False Positive
                  </SelectLabel>
                  <SelectItem value="FalsePositive_IncorrectLogic">Incorrect Rule Logic</SelectItem>
                  <SelectItem value="FalsePositive_InaccurateData">Inaccurate Data Enrichment</SelectItem>
                </SelectGroup>

                <SelectGroup>
                  <SelectLabel className="flex items-center gap-2 text-muted-foreground px-2 py-1.5 text-[10px] uppercase font-bold mt-2">
                    <HelpCircle className="w-3 h-3" /> Other
                  </SelectLabel>
                  <SelectItem value="Undetermined">Undetermined / Unknown</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>

          {/* 2. 备注输入 (Comment) */}
          <div className="space-y-2">
            <Label className="text-xs font-bold uppercase text-muted-foreground">
              Investigation Notes / Comments
            </Label>
            <Textarea
              placeholder="Provide a brief summary of your investigation... (e.g., 'User confirmed password reset', 'External IP is a known CDN')"
              className="h-[120px] resize-none focus-visible:ring-emerald-500"
              value={comment}
              onChange={(e) => setComment(e.target.value)}
            />
          </div>
        </div>

        <DialogFooter className="bg-muted/5 -mx-6 px-6 py-4 border-t">
          <Button variant="ghost" onClick={() => onOpenChange(false)} disabled={loading}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!classification || loading}
            className="bg-emerald-600 hover:bg-emerald-700 text-white min-w-[120px]"
          >
            {loading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Resolving...
              </>
            ) : "Resolve Incident"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}