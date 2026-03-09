import { useEffect, useState, useRef } from "react";
import { forensicsService, type ForensicTask, type ForensicFile } from "@/services/forensics";
import { useTabStore } from "@/stores/tab-store"; // ✅ 引入 Tab 仓库，用于跳转日志分析
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ArrowLeft, UploadCloud, Search, Trash2, FileText, Network, AlertCircle, Loader2, CheckCircle2, Clock } from "lucide-react";
import { toast } from "sonner";
import { ScrollArea } from "@/components/ui/scroll-area";

interface CaseWorkspaceProps {
  caseId: number;
  onBack: () => void;
}

export function CaseWorkspace({ caseId, onBack }: CaseWorkspaceProps) {
  const [task, setTask] = useState<ForensicTask | null>(null);
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  
  const { addTab } = useTabStore();

  const fetchTaskDetails = async () => {
    try {
      const res = await forensicsService.getTask(caseId);
      if (res.code === 200 && res.data) {
        setTask(res.data);
      }
    } catch (e) {
      console.error(e);
    }
  };

  // 1. 初始化加载
  useEffect(() => {
    fetchTaskDetails();
  }, [caseId]);

  // 2. 智能轮询引擎：如果发现有文件处于 pending 或 parsing 状态，每 3 秒刷新一次
  useEffect(() => {
    if (!task || !task.files) return;
    
    const isProcessing = task.files.some(f => f.parse_status === 'pending' || f.parse_status === 'parsing');
    if (isProcessing) {
      const timer = setInterval(fetchTaskDetails, 3000);
      return () => clearInterval(timer);
    }
  }, [task]);

  // 触发隐藏的文件选择器
  const triggerUpload = () => {
    fileInputRef.current?.click();
  };

  // 处理文件上传
  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files || files.length === 0) return;
    
    const file = files[0];
    setUploading(true);
    try {
      await forensicsService.uploadFile(caseId, file);
      toast.success("File uploaded, parsing started.");
      fetchTaskDetails(); // 立即刷新列表看状态
    } catch (err: any) {
      toast.error("Upload failed", { description: err.message });
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = ""; // 重置 input
    }
  };

  const handleDeleteFile = async (fileId: number) => {
    if (!confirm("Delete this evidence file?")) return;
    try {
      await forensicsService.deleteFile(fileId);
      toast.success("File deleted");
      fetchTaskDetails();
    } catch (err) {
      toast.error("Delete failed");
    }
  };

  // ✅ 核心联动：一键跳转 LogSQL 查询
  const handleHuntInSandbox = () => {
    // 利用 VictoriaLogs 的强大能力，限定只查当前沙箱隔离环境的数据
    const sandboxQuery = `env="forensics" AND task_id="${caseId}"`;
    addTab('logs', `Hunt: ${task?.name}`, { query: sandboxQuery });
  };

  if (!task) return <div className="p-6 flex items-center justify-center h-full"><Loader2 className="animate-spin text-primary" /></div>;

  // 辅助渲染文件图标
  const getFileIcon = (type: string) => {
    if (type.includes('pcap')) return <Network className="w-4 h-4 text-blue-500" />;
    return <FileText className="w-4 h-4 text-slate-500" />;
  };

  // 辅助渲染状态 Badge
  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'completed': return <Badge variant="outline" className="border-green-200 text-green-600 bg-green-50"><CheckCircle2 className="w-3 h-3 mr-1"/> Completed</Badge>;
      case 'failed': return <Badge variant="destructive" className="bg-red-100 text-red-700 hover:bg-red-100 border-none"><AlertCircle className="w-3 h-3 mr-1"/> Failed</Badge>;
      case 'parsing': return <Badge variant="secondary" className="bg-blue-50 text-blue-600 border-none"><Loader2 className="w-3 h-3 mr-1 animate-spin"/> Parsing</Badge>;
      default: return <Badge variant="secondary" className="text-muted-foreground"><Clock className="w-3 h-3 mr-1"/> Pending</Badge>;
    }
  };

  const totalEvents = task.files?.reduce((acc, f) => acc + (f.event_count || 0), 0) || 0;

  return (
    <div className="flex flex-col h-full bg-background">
      {/* 顶部导航栏 */}
      <div className="h-14 border-b bg-muted/10 px-4 flex items-center justify-between flex-none">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={onBack} className="h-8 w-8">
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <h2 className="text-sm font-bold flex items-center gap-2">
              <span className="text-muted-foreground font-normal">Case Workspace /</span> {task.name}
            </h2>
          </div>
        </div>
        
        {/* 沙箱分析入口 */}
        <Button className="bg-purple-600 hover:bg-purple-700 shadow-sm" onClick={handleHuntInSandbox} disabled={totalEvents === 0}>
          <Search className="w-4 h-4 mr-2" /> 
          Deep Hunt ({totalEvents} Events)
        </Button>
      </div>

      <div className="flex-1 p-6 flex flex-col min-h-0">
        <Card className="flex-1 shadow-sm flex flex-col min-h-0 border-t-4 border-t-primary/20">
          <CardHeader className="flex flex-row items-center justify-between py-4 border-b flex-none">
            <div>
              <CardTitle className="text-base">Evidence Vault</CardTitle>
              <CardDescription className="text-xs">Upload raw PCAP, EVTX or text logs. They will be parsed and isolated.</CardDescription>
            </div>
            
            {/* 隐藏的 File Input */}
            <input type="file" ref={fileInputRef} className="hidden" onChange={handleFileChange} />
            <Button variant="outline" onClick={triggerUpload} disabled={uploading}>
              {uploading ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : <UploadCloud className="w-4 h-4 mr-2" />}
              {uploading ? "Uploading..." : "Upload Evidence"}
            </Button>
          </CardHeader>
          
          <CardContent className="flex-1 p-0 overflow-hidden relative">
            <ScrollArea className="h-full w-full">
              {!task.files || task.files.length === 0 ? (
                <div className="absolute inset-0 flex flex-col items-center justify-center text-muted-foreground/50 space-y-3">
                  <UploadCloud className="w-10 h-10 opacity-20" />
                  <p className="text-sm">Vault is empty. Upload evidence to start parsing.</p>
                </div>
              ) : (
                <Table>
                  <TableHeader className="sticky top-0 bg-background/95 backdrop-blur z-10 border-b">
                    <TableRow>
                      <TableHead className="w-[300px]">File Name</TableHead>
                      <TableHead className="w-[100px]">Type</TableHead>
                      <TableHead className="w-[120px]">Size</TableHead>
                      <TableHead className="w-[150px]">Status</TableHead>
                      <TableHead className="w-[100px] text-right">Extracted</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {task.files.map(file => (
                      <TableRow key={file.id} className="group h-12">
                        <TableCell className="font-medium text-xs">
                          <div className="flex items-center gap-2 truncate max-w-[280px]" title={file.original_name}>
                            {getFileIcon(file.file_type)}
                            {file.original_name}
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary" className="uppercase text-[9px] font-mono">{file.file_type}</Badge>
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground font-mono">
                          {(file.file_size / 1024 / 1024).toFixed(2)} MB
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col gap-1">
                            {getStatusBadge(file.parse_status)}
                            {file.parse_status === 'failed' && (
                              <span className="text-[9px] text-red-500 truncate max-w-[150px]" title={file.parse_message}>{file.parse_message}</span>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="text-right font-mono text-xs">
                          {file.event_count > 0 ? file.event_count : "-"}
                        </TableCell>
                        <TableCell className="text-right">
                          <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity" onClick={() => handleDeleteFile(file.id)}>
                            <Trash2 className="w-4 h-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </ScrollArea>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}