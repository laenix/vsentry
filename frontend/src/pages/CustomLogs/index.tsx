import { useEffect, useState } from "react";
import { customTableService, type CustomTable } from "@/services/customtable";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Plus, Trash2, Edit, Database, RefreshCw } from "lucide-react";
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

const formatDate = (dateStr?: string) => {
  if (!dateStr) return "-";
  return new Date(dateStr).toLocaleString("zh-CN", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

export default function CustomTablesPage() {
  const [tables, setTables] = useState<CustomTable[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingTable, setEditingTable] = useState<CustomTable | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const [formData, setFormData] = useState({
    name: "",
    stream_fields: "",
    description: "",
    query: "*",
  });

  const fetchTables = async () => {
    try {
      const res = await customTableService.list();
      if (res.code === 200) {
        let list: CustomTable[] = [];
        if (Array.isArray(res.data)) {
          list = res.data;
        } else if (res.data?.tables) {
          list = res.data.tables;
        }
        setTables(list);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTables();
  }, []);

  const getTableID = (table: CustomTable) => table.ID || table.id || 0;

  const handleOpenDialog = (table?: CustomTable) => {
    if (table) {
      setEditingTable(table);
      setFormData({
        name: table.name,
        stream_fields: table.stream_fields,
        description: table.description || "",
        query: table.query || "*",
      });
    } else {
      setEditingTable(null);
      setFormData({
        name: "",
        stream_fields: "",
        description: "",
        query: "*",
      });
    }
    setDialogOpen(true);
  };

  const handleSubmit = async () => {
    if (!formData.name || !formData.stream_fields) {
      toast.error("Name and Stream Fields are required");
      return;
    }

    setSubmitting(true);
    try {
      if (editingTable) {
        await customTableService.update({ id: getTableID(editingTable), ...formData });
        toast.success("Table updated successfully");
      } else {
        await customTableService.add(formData);
        toast.success("Table created successfully");
      }
      setDialogOpen(false);
      fetchTables();
    } catch (err) {
      console.error(err);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure you want to delete this table?")) return;
    try {
      await customTableService.delete(id);
      toast.success("Table deleted");
      fetchTables();
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <div className="p-6 h-full flex flex-col">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold">Custom Tables</h1>
          <p className="text-muted-foreground text-sm">
            Define custom log tables using stream fields for targeted queries.
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={fetchTables}>
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
          <Button onClick={() => handleOpenDialog()}>
            <Plus className="w-4 h-4 mr-2" />
            New Table
          </Button>
        </div>
      </div>

      <div className="border rounded-md bg-card flex-1 overflow-auto">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/5">
              <TableHead className="w-[50px]">ID</TableHead>
              <TableHead>Table Name</TableHead>
              <TableHead>Stream Fields</TableHead>
              <TableHead>Default Query</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {tables.length > 0 ? (
              tables.map((table) => {
                const id = getTableID(table);
                return (
                  <TableRow key={id}>
                    <TableCell className="font-mono text-xs text-muted-foreground">
                      #{id}
                    </TableCell>
                    <TableCell className="font-medium flex items-center gap-2">
                      <Database className="w-4 h-4 text-muted-foreground" />
                      {table.name}
                    </TableCell>
                    <TableCell>
                      <code className="text-xs bg-muted px-1.5 py-0.5 rounded">
                        {table.stream_fields}
                      </code>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground max-w-[150px] truncate">
                      {table.query || "*"}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground max-w-[200px] truncate">
                      {table.description || "-"}
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground whitespace-nowrap">
                      {formatDate(table.created_at)}
                    </TableCell>
                    <TableCell className="text-right space-x-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => handleOpenDialog(table)}
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
                    No custom tables defined. Create one to get started.
                  </TableCell>
                </TableRow>
              )
            )}
          </TableBody>
        </Table>
      </div>

      {/* Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>
              {editingTable ? "Edit Custom Table" : "Create Custom Table"}
            </DialogTitle>
            <DialogDescription>
              Define a custom table with stream fields for efficient log queries.
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="name">Table Name</Label>
              <Input
                id="name"
                placeholder="e.g., Windows Events, nginx Logs"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="stream_fields">Stream Fields</Label>
              <Input
                id="stream_fields"
                placeholder="e.g., host,service or channel,source"
                value={formData.stream_fields}
                onChange={(e) => setFormData({ ...formData, stream_fields: e.target.value })}
              />
              <p className="text-xs text-muted-foreground">
                Comma-separated fields used to group logs into tables
              </p>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="query">Default Query</Label>
              <Input
                id="query"
                placeholder="*"
                value={formData.query}
                onChange={(e) => setFormData({ ...formData, query: e.target.value })}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                placeholder="Optional description"
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
              {submitting ? "Saving..." : editingTable ? "Update" : "Create"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}