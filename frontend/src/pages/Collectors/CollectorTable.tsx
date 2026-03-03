import type { CollectorConfig } from "@/services/collector";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Download, Edit, Trash2, RefreshCw, Hammer, CheckCircle } from "lucide-react";
import { typeIcons } from "./constants";

interface CollectorTableProps {
  configs: CollectorConfig[];
  loading: boolean;
  emptyMessage?: string;
  onBuild: (config: CollectorConfig) => void;
  onDownload: (config: CollectorConfig) => void;
  onEdit: (config: CollectorConfig) => void;
  onDelete: (id: number) => void;
}

export function CollectorTable({ configs, loading, emptyMessage = "No collectors configured.", onBuild, onDownload, onEdit, onDelete }: CollectorTableProps) {
  const getConfigID = (c: CollectorConfig) => c.ID || c.id || 0;

  const renderActionButtons = (config: CollectorConfig) => {
    const isBuilding = config.build_status === "building";
    const isReady = config.build_status === "completed";
    const id = getConfigID(config);

    return (
      <TableCell className="text-right space-x-1">
        {isReady ? (
          <>
            <Button variant="outline" size="sm" onClick={() => onDownload(config)} className="bg-primary/5 border-primary/20 text-primary hover:bg-primary/10">
              <Download className="w-3.5 h-3.5 mr-2" /> Download
            </Button>
            <Button variant="ghost" size="icon" title="Rebuild Agent" onClick={() => onBuild(config)}>
              <RefreshCw className="w-3.5 h-3.5" />
            </Button>
          </>
        ) : (
          <Button variant="outline" size="sm" onClick={() => onBuild(config)} disabled={isBuilding}>
            {isBuilding ? <RefreshCw className="w-3.5 h-3.5 mr-2 animate-spin" /> : <Hammer className="w-3.5 h-3.5 mr-2" />}
            {isBuilding ? "Compiling..." : "Build"}
          </Button>
        )}
        <Button variant="ghost" size="icon" onClick={() => onEdit(config)}>
          <Edit className="w-3.5 h-3.5" />
        </Button>
        <Button variant="ghost" size="icon" className="text-red-500 hover:text-red-600" onClick={() => onDelete(id)}>
          <Trash2 className="w-3.5 h-3.5" />
        </Button>
      </TableCell>
    );
  };

  return (
    <div className="border rounded-md bg-card">
      <Table>
        <TableHeader>
          <TableRow className="bg-muted/5">
            <TableHead className="w-[50px]">ID</TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {configs.length > 0 ? (
            configs.map((config) => {
              const id = getConfigID(config);
              const Icon = typeIcons[config.type] || typeIcons.windows;
              return (
                <TableRow key={id}>
                  <TableCell className="font-mono text-xs text-muted-foreground">#{id}</TableCell>
                  <TableCell className="font-medium flex items-center gap-2">
                    <Icon className="w-4 h-4 text-muted-foreground" />
                    {config.name}
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary" className="capitalize">{config.type}</Badge>
                  </TableCell>
                  <TableCell>
                    {config.build_status === "completed" ? (
                      <Badge className="bg-green-500/10 text-green-500 border-green-500/20">
                        <CheckCircle className="w-3 h-3 mr-1"/> Ready
                      </Badge>
                    ) : config.build_status === "building" ? (
                      <Badge className="bg-blue-500/10 text-blue-500 border-blue-500/20">
                        <RefreshCw className="w-3 h-3 mr-1 animate-spin"/> Compiling
                      </Badge>
                    ) : (
                      <Badge variant="outline">Pending</Badge>
                    )}
                  </TableCell>
                  {renderActionButtons(config)}
                </TableRow>
              );
            })
          ) : (
            !loading && (
              <TableRow>
                <TableCell colSpan={5} className="h-32 text-center text-muted-foreground">
                  {emptyMessage}
                </TableCell>
              </TableRow>
            )
          )}
        </TableBody>
      </Table>
    </div>
  );
}