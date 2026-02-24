import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import PlaybookEditor from "@/components/automation/PlaybookEditor";
import { Loader2 } from "lucide-react";
import { automationService } from "@/services/automation";
import type { Node, Edge } from "@xyflow/react";
import { toast } from "sonner";

export default function EditPlaybookPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  
  const [loading, setLoading] = useState(true);
  const [playbookData, setPlaybookData] = useState<{
    id: string;
    name: string;
    nodes: Node[];
    edges: Edge[];
  } | null>(null);

  useEffect(() => {
    const initData = async () => {
      setLoading(true);
      try {
        if (id === 'new') {
          // 新建模式：给一个默认的 Trigger 节点
          setPlaybookData({
            id: 'new',
            name: "New Playbook",
            nodes: [{ 
              id: 'trigger', 
              type: 'custom', 
              position: { x: 250, y: 50 }, 
              data: { label: 'Incident Created', type: 'trigger', icon: 'trigger' } 
            }],
            edges: []
          });
        } else {
          // 编辑模式：从后端拉取真实数据
          const res = await automationService.getDetail(id!);
          if (res.code === 200 && res.data) {
            const def = res.data.definition || { nodes: [], edges: [] };
            setPlaybookData({
              id: String(res.data.ID),
              name: res.data.name,
              nodes: def.nodes || [],
              edges: def.edges || []
            });
          } else {
            toast.error("Playbook not found");
            navigate("/automation");
          }
        }
      } catch (error) {
        console.error(error);
        toast.error("Failed to load playbook");
      } finally {
        setLoading(false);
      }
    };

    initData();
  }, [id, navigate]);

  if (loading) {
    return (
      <div className="h-screen w-full flex items-center justify-center bg-background">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!playbookData) return null;

  return (
    <div className="h-screen w-full bg-background">
      <PlaybookEditor 
        playbookId={playbookData.id} // 传入 ID
        playbookName={playbookData.name}
        initialNodes={playbookData.nodes}
        initialEdges={playbookData.edges}
        onBack={() => navigate('/automation')} 
      />
    </div>
  );
}