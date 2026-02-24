import { useEffect, useState } from "react";
import { ingestService, configService, type IngestConfig, type SystemConfig } from "@/services/ingest";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Plus, Trash2, RefreshCw, Server, Database, Network, Copy, Key } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const formatDate = (dateStr?: string) => {
  if (!dateStr) return "-";
  return new Date(dateStr).toLocaleString("zh-CN", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

export default function IngestPage() {
  const [ingests, setIngests] = useState<IngestConfig[]>([]);
  const [tokens, setTokens] = useState<Record<number, string>>({});
  const [externalUrl, setExternalUrl] = useState<string>("http://localhost:8088");
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingIngest, setEditingIngest] = useState<IngestConfig | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const [formData, setFormData] = useState({
    name: "",
    endpoint: "",
    type: "victorialogs",
    source: "build-in",
    _stream_fields: "",
  });

  const fetchIngests = async () => {
    try {
      // Fetch config first
      try {
        const configRes = await configService.get();
        if (configRes.code === 200 && configRes.data?.external_url) {
          setExternalUrl(configRes.data.external_url);
        }
      } catch (e) {
        console.error("Failed to fetch config");
      }

      const res = await ingestService.list();
      if (res.code === 200) {
        let list: IngestConfig[] = [];
        if (Array.isArray(res.data)) {
          list = res.data;
        } else if (res.data?.ingests) {
          list = res.data.ingests;
        }
        setIngests(list);
        
        // Fetch tokens for each ingest
        const tokenMap: Record<number, string> = {};
        for (const ingest of list) {
          const id = ingest.ID || ingest.id;
          if (id) {
            try {
              const tokenRes = await ingestService.getAuth(id);
              if (tokenRes.code === 200 && tokenRes.data?.token) {
                tokenMap[id] = tokenRes.data.token;
              }
            } catch (e) {
              console.error(`Failed to fetch token for ingest ${id}`);
            }
          }
        }
        setTokens(tokenMap);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchIngests();
  }, []);

  const getIngestID = (ingest: IngestConfig) => ingest.ID || ingest.id || 0;

  const handleOpenDialog = (ingest?: IngestConfig) => {
    if (ingest) {
      setEditingIngest(ingest);
      setFormData({
        name: ingest.name,
        endpoint: ingest.endpoint,
        type: ingest.type,
        source: ingest.source,
        _stream_fields: ingest._stream_fields || "",
      });
    } else {
      setEditingIngest(null);
      setFormData({
        name: "",
        endpoint: "",
        type: "victorialogs",
        source: "build-in",
        _stream_fields: "",
      });
    }
    setDialogOpen(true);
  };

  const handleSubmit = async () => {
    if (!formData.name || !formData.endpoint) {
      toast.error("Name and Endpoint are required");
      return;
    }

    setSubmitting(true);
    try {
      if (editingIngest) {
        await ingestService.update({ id: getIngestID(editingIngest), ...formData });
        toast.success("Ingest updated successfully");
      } else {
        await ingestService.add(formData);
        toast.success("Ingest created successfully");
      }
      setDialogOpen(false);
      fetchIngests();
    } catch (err) {
      console.error(err);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure you want to delete this ingest?")) return;
    try {
      await ingestService.delete(id);
      toast.success("Ingest deleted");
      fetchIngests();
    } catch (err) {
      console.error(err);
    }
  };

  const copyToken = (token: string) => {
    navigator.clipboard.writeText(token);
    toast.success("Token copied to clipboard");
  };

  const getExternalUrl = () => {
    return `${externalUrl}/api/ingest/collect`;
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case "victorialogs":
        return <Database className="w-4 h-4" />;
      case "elasticsearch":
        return <Server className="w-4 h-4" />;
      default:
        return <Network className="w-4 h-4" />;
    }
  };

  return (
    <div className="p-6 h-full flex flex-col">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold">Ingest Management</h1>
          <p className="text-muted-foreground text-sm">
            Configure log collection endpoints and data streams.
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={fetchIngests}>
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
          <Button onClick={() => handleOpenDialog()}>
            <Plus className="w-4 h-4 mr-2" />
            New Ingest
          </Button>
        </div>
      </div>

      <div className="border rounded-md bg-card flex-1 overflow-auto">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/5">
              <TableHead className="w-[50px]">ID</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Endpoint</TableHead>
              <TableHead>Token</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Source</TableHead>
              <TableHead>Stream Fields</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {Array.isArray(ingests) && ingests.length > 0 ? (
              ingests.map((ingest) => {
                const id = getIngestID(ingest);
                return (
                  <TableRow key={id}>
                    <TableCell className="font-mono text-xs text-muted-foreground">
                      #{id}
                    </TableCell>
                    <TableCell className="font-medium">{ingest.name}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground truncate max-w-[200px]">
                      {ingest.endpoint}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <code className="text-xs bg-muted px-1.5 py-0.5 rounded max-w-[100px] truncate">
                          {tokens[id]?.substring(0, 12) || "-"}...
                        </code>
                        {tokens[id] && (
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6"
                            onClick={() => copyToken(tokens[id])}
                            title="Copy full token"
                          >
                            <Copy className="w-3 h-3" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="flex items-center gap-1 w-fit">
                        {getTypeIcon(ingest.type)}
                        {ingest.type}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary">{ingest.source}</Badge>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground max-w-[150px] truncate">
                      {ingest._stream_fields || "-"}
                    </TableCell>
                    <TableCell className="text-right space-x-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => handleOpenDialog(ingest)}
                      >
                        <RefreshCw className="w-3.5 h-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-red-500 hover:text-red-600 hover:bg-red-50"
                        onClick={() => handleDelete(id)}
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </Button>
                    </TableCell>
                  </TableRow>
                );
              })
            ) : (
              !loading && (
                <TableRow>
                  <TableCell colSpan={8} className="h-32 text-center text-muted-foreground">
                    No ingest configurations found.
                  </TableCell>
                </TableRow>
              )
            )}
          </TableBody>
        </Table>
      </div>

      {/* Ingest Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>
              {editingIngest ? "Edit Ingest" : "Create New Ingest"}
            </DialogTitle>
            <DialogDescription>
              Configure the log collection endpoint and stream settings.
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                placeholder="e.g., Production Logs"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="endpoint">Endpoint URL</Label>
              <Input
                id="endpoint"
                placeholder="http://localhost:9428/insert/logsql"
                value={formData.endpoint}
                onChange={(e) => setFormData({ ...formData, endpoint: e.target.value })}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="type">Type</Label>
                <Select
                  value={formData.type}
                  onValueChange={(value) => setFormData({ ...formData, type: value })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="victorialogs">VictoriaLogs</SelectItem>
                    <SelectItem value="elasticsearch">Elasticsearch</SelectItem>
                    <SelectItem value="custom">Custom</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="source">Source</Label>
                <Select
                  value={formData.source}
                  onValueChange={(value) => setFormData({ ...formData, source: value })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="build-in">Build-in</SelectItem>
                    <SelectItem value="custom">Custom</SelectItem>
                    <SelectItem value="syslog">Syslog</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="stream_fields">Stream Fields</Label>
              <Input
                id="stream_fields"
                placeholder="_stream_fields=channel,source,host"
                value={formData._stream_fields}
                onChange={(e) => setFormData({ ...formData, _stream_fields: e.target.value })}
              />
              <p className="text-xs text-muted-foreground">
                Comma-separated fields for log stream identification
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSubmit} disabled={submitting}>
              {submitting ? "Saving..." : editingIngest ? "Update" : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}