import { useEffect, useState } from "react";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue
} from "@/components/ui/select"
import { User, Loader2, UserPlus, Shield } from "lucide-react";
import { apiClient } from "@/lib/api/vsentry-client"; // 确保路径正确
import { toast } from "sonner";

interface Analyst {
  id: number;
  username: string;
  role: string;
}

interface IncidentAssignDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (userId: number) => Promise<void>;
  currentAssignee?: number;
}

export function IncidentAssignDialog({ open, onOpenChange, onConfirm, currentAssignee }: IncidentAssignDialogProps) {
  const [analysts, setAnalysts] = useState<Analyst[]>([]);
  const [selectedUserId, setSelectedUserId] = useState<string>("");
  const [fetching, setFetching] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  // 1. 当弹窗打开时，拉取最新的分析师列表
  useEffect(() => {
    if (open) {
      const fetchAnalysts = async () => {
        setFetching(true);
        try {
          // 调用你刚才在后端补全的 /users/list 接口
          const res = await apiClient.get("/users/list");
          if (res.code === 200) {
            setAnalysts(res.data || []);
          }
        } catch (error) {
          console.error("Failed to load analysts:", error);
          toast.error("Failed to load analyst list");
        } finally {
          setFetching(false);
        }
      };
      fetchAnalysts();
    }
  }, [open]);

  const handleSubmit = async () => {
    if (!selectedUserId) return;
    setSubmitting(true);
    try {
      await onConfirm(parseInt(selectedUserId));
      onOpenChange(false);
      setSelectedUserId("");
    } catch (error) {
      // 错误已由 index.tsx 处理
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[420px]">
        <DialogHeader>
          <div className="flex items-center gap-2 mb-1">
            <UserPlus className="w-5 h-5 text-primary" />
            <DialogTitle>Assign Incident</DialogTitle>
          </div>
          <DialogDescription>
            Select a security analyst to investigate this incident.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="space-y-2">
            <Label className="text-xs font-bold text-muted-foreground uppercase">
              Select Security Analyst
            </Label>
            <Select
              value={selectedUserId}
              onValueChange={setSelectedUserId}
              disabled={fetching || submitting}
            >
              <SelectTrigger className="h-12">
                <SelectValue placeholder={fetching ? "Loading analysts..." : "Select a user..."} />
              </SelectTrigger>
              <SelectContent>
                {analysts.length > 0 ? (
                  analysts.map(user => (
                    <SelectItem
                      key={user.id}
                      value={user.id.toString()}
                      disabled={user.id === currentAssignee}
                    >
                      <div className="flex items-center gap-3 py-1">
                        <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary/10 text-primary border border-primary/20">
                          <User className="w-4 h-4" />
                        </div>
                        <div className="flex flex-col items-start">
                          <span className="text-sm font-medium">
                            {user.username} {user.id === currentAssignee && "(Current)"}
                          </span>
                          <span className="text-[10px] text-muted-foreground flex items-center gap-1">
                            <Shield className="w-3 h-3" /> {user.role || "Analyst"}
                          </span>
                        </div>
                      </div>
                    </SelectItem>
                  ))
                ) : (
                  <div className="p-4 text-center text-xs text-muted-foreground">
                    No analysts available
                  </div>
                )}
              </SelectContent>
            </Select>
          </div>
        </div>

        <DialogFooter className="border-t pt-4 bg-muted/5 -mx-6 px-6">
          <Button
            variant="ghost"
            onClick={() => onOpenChange(false)}
            disabled={submitting}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!selectedUserId || submitting || fetching}
            className="min-w-[100px]"
          >
            {submitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Assigning...
              </>
            ) : "Confirm Assignment"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}