import { useEffect, useState } from "react";
import { connectorService, type Connector, type ConnectorTemplate } from "@/services/connector";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Plus, Trash2, Edit, Plug, Search, Shield, Cloud, Database, Server, Router, Key, Activity } from "lucide-react";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const typeIcons: Record<string, any> = {
  app: Database,
  security: Shield,
  network: Router,
  cloud: Cloud,
  database: Database,
  middleware: Server,
};

const protocolLabels: Record<string, string> = {
  api: "API",
  syslog: "Syslog",
  ssh: "SSH",
  jdbc: "JDBC",
  kafka: "Kafka",
  s3: "S3",
};

export default function ConnectorsPage() {
  const [connectors, setConnectors] = useState<Connector[]>([]);
  const [templates, setTemplates] = useState<ConnectorTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingConnector, setEditingConnector] = useState<Connector | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [activeTab, setActiveTab] = useState< string>("all");

  const [formData, setFormData] = useState({
    name: "",
    type: "app",
    protocol: "api",
    host: "",
    port: 443,
    username: "",
    password: "",
    api_key: "",
    endpoint: "",
    description: "",
  });

  const fetchData = async () => {
    try {
      const [connRes, tmplRes] = await Promise.all([
        connectorService.list(),
        connectorService.templates(),
      ]);
      
      if (connRes.code === 200) {
        setConnectors(Array.isArray(connRes.data) ? connRes.data : []);
      }
      if (tmplRes.code === 200) {
        setTemplates(tmplRes.data || []);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const getConnectorID = (c: Connector) => c.ID || c.id || 0;

  const handleOpenDialog = (connector?: Connector, template?: ConnectorTemplate) => {
    if (connector) {
      setEditingConnector(connector);
      setFormData({
        name: connector.name,
        type: connector.type,
        protocol: connector.protocol,
        host: connector.host || "",
        port: connector.port || 443,
        username: connector.username || "",
        password: connector.password || "",
        api_key: connector.api_key || "",
        endpoint: connector.endpoint || "",
        description: connector.description || "",
      });
    } else if (template) {
      setEditingConnector(null);
      setFormData({
        name: template.name,
        type: template.type,
        protocol: template.protocol,
        host: "",
        port: template.default_port,
        username: "",
        password: "",
        api_key: "",
        endpoint: "",
        description: template.description,
      });
    } else {
      setEditingConnector(null);
      setFormData({
        name: "",
        type: "app",
        protocol: "api",
        host: "",
        port: 443,
        username: "",
        password: "",
        api_key: "",
        endpoint: "",
        description: "",
      });
    }
    setDialogOpen(true);
  };

  const handleSubmit = async () => {
    if (!formData.name) {
      toast.error("Name is required");
      return;
    }

    setSubmitting(true);
    try {
      if (editingConnector) {
        await connectorService.update({ id: getConnectorID(editingConnector), ...formData });
        toast.success("Connector updated successfully");
      } else {
        await connectorService.add(formData);
        toast.success("Connector created successfully");
      }
      setDialogOpen(false);
      fetchData();
    } catch (err) {
      console.error(err);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure you want to delete this connector?")) return;
    try {
      await connectorService.delete(id);
      toast.success("Connector deleted");
      fetchData();
    } catch (err) {
      console.error(err);
    }
  };

  const filteredConnectors = activeTab === "all" 
    ? connectors 
    : connectors.filter(c => c.type === activeTab);

  const filteredTemplates = activeTab === "all"
    ? templates
    : templates.filter(t => t.type === activeTab);

  return (
    <div className="p-6 h-full flex flex-col">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h1 className="text-2xl font-bold flex items-center gap-2">
              <Plug className="w-6 h-6" />
              Connectors
            </h1>
            <p className="text-muted-foreground text-sm">
              Integrate with third-party applications and services.
            </p>
          </div>
          <Button onClick={() => handleOpenDialog()}>
            <Plus className="w-4 h-4 mr-2" />
            Add Connector
          </Button>
        </div>

        <TabsList className="mb-4">
          <TabsTrigger value="all">All</TabsTrigger>
          <TabsTrigger value="app">Applications</TabsTrigger>
          <TabsTrigger value="security">Security</TabsTrigger>
          <TabsTrigger value="network">Network</TabsTrigger>
          <TabsTrigger value="cloud">Cloud</TabsTrigger>
          <TabsTrigger value="database">Database</TabsTrigger>
        </TabsList>

        <TabsContent value={activeTab} className="flex-1 mt-0">
          {/* Connector Templates */}
          <div className="mb-6">
            <h3 className="text-sm font-medium text-muted-foreground mb-3">Available Integrations</h3>
            <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-3">
              {filteredTemplates.map((template) => {
                const Icon = typeIcons[template.type] || Plug;
                return (
                  <Card 
                    key={template.id} 
                    className="cursor-pointer hover:border-primary transition-colors"
                    onClick={() => handleOpenDialog(undefined, template)}
                  >
                    <CardContent className="p-3 flex items-center gap-2">
                      <Icon className="w-5 h-5 text-muted-foreground" />
                      <div className="min-w-0">
                        <p className="text-sm font-medium truncate">{template.name}</p>
                        <p className="text-xs text-muted-foreground">{protocolLabels[template.protocol]}</p>
                      </div>
                    </CardContent>
                  </Card>
                );
              })}
            </div>
          </div>

          {/* Configured Connectors */}
          <div>
            <h3 className="text-sm font-medium text-muted-foreground mb-3">Configured Connectors</h3>
            <div className="border rounded-md bg-card">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/5">
                    <TableHead className="w-[50px]">ID</TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Protocol</TableHead>
                    <TableHead>Host:Port</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredConnectors.length > 0 ? (
                    filteredConnectors.map((connector) => {
                      const id = getConnectorID(connector);
                      const Icon = typeIcons[connector.type] || Plug;
                      return (
                        <TableRow key={id}>
                          <TableCell className="font-mono text-xs text-muted-foreground">#{id}</TableCell>
                          <TableCell className="font-medium flex items-center gap-2">
                            <Icon className="w-4 h-4 text-muted-foreground" />
                            {connector.name}
                          </TableCell>
                          <TableCell>
                            <Badge variant="secondary" className="capitalize">{connector.type}</Badge>
                          </TableCell>
                          <TableCell>
                            <Badge variant="outline">{protocolLabels[connector.protocol] || connector.protocol}</Badge>
                          </TableCell>
                          <TableCell className="font-mono text-xs">
                            {connector.host ? `${connector.host}:${connector.port}` : "-"}
                          </TableCell>
                          <TableCell>
                            {connector.is_enabled ? (
                              <Badge className="bg-green-500">Active</Badge>
                            ) : (
                              <Badge variant="secondary">Disabled</Badge>
                            )}
                          </TableCell>
                          <TableCell className="text-right space-x-1">
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8"
                              onClick={() => handleOpenDialog(connector)}
                            >
                              <Edit className="w-3.5 h-3.5" />
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
                        <TableCell colSpan={7} className="h-32 text-center text-muted-foreground">
                          No connectors configured. Click a template above to add one.
                        </TableCell>
                      </TableRow>
                    )
                  )}
                </TableBody>
              </Table>
            </div>
          </div>
        </TabsContent>
      </Tabs>

      {/* Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>
              {editingConnector ? "Edit Connector" : "Configure Connector"}
            </DialogTitle>
            <DialogDescription>
              Enter connection details for the integration.
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4 max-h-[60vh] overflow-y-auto">
            <div className="grid gap-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label>Type</Label>
                <Select value={formData.type} onValueChange={(v) => setFormData({ ...formData, type: v })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="app">Application</SelectItem>
                    <SelectItem value="security">Security</SelectItem>
                    <SelectItem value="network">Network</SelectItem>
                    <SelectItem value="cloud">Cloud</SelectItem>
                    <SelectItem value="database">Database</SelectItem>
                    <SelectItem value="middleware">Middleware</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label>Protocol</Label>
                <Select value={formData.protocol} onValueChange={(v) => setFormData({ ...formData, protocol: v })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="api">API</SelectItem>
                    <SelectItem value="syslog">Syslog</SelectItem>
                    <SelectItem value="ssh">SSH</SelectItem>
                    <SelectItem value="jdbc">JDBC</SelectItem>
                    <SelectItem value="kafka">Kafka</SelectItem>
                    <SelectItem value="s3">S3</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="host">Host</Label>
                <Input
                  id="host"
                  placeholder="192.168.1.1 or example.com"
                  value={formData.host}
                  onChange={(e) => setFormData({ ...formData, host: e.target.value })}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="port">Port</Label>
                <Input
                  id="port"
                  type="number"
                  value={formData.port}
                  onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value) || 443 })}
                />
              </div>
            </div>

            {(formData.protocol === "api" || formData.protocol === "ssh") && (
              <>
                <div className="grid gap-2">
                  <Label htmlFor="username">Username</Label>
                  <Input
                    id="username"
                    value={formData.username}
                    onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="password">Password</Label>
                  <Input
                    id="password"
                    type="password"
                    value={formData.password}
                    onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  />
                </div>
              </>
            )}

            {formData.protocol === "api" && (
              <div className="grid gap-2">
                <Label htmlFor="api_key">API Key</Label>
                <Input
                  id="api_key"
                  value={formData.api_key}
                  onChange={(e) => setFormData({ ...formData, api_key: e.target.value })}
                />
              </div>
            )}

            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSubmit} disabled={submitting}>
              {submitting ? "Saving..." : editingConnector ? "Update" : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}