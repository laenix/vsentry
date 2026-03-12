import { Package, Server, Activity } from "lucide-react";

// 平台图标映射 - const typeIcons: Record<string, any> = {
  windows: Package,
  linux: Server,
  macos: Activity,
};

// 增强的Data源Interface - interface DataSource {
  type: string;
  path: string;
  label: string;
  enabled: boolean;
  event_ids_str?: string; 
  query?: string;         
  presets?: { name: string, ids: string }[]; 
}