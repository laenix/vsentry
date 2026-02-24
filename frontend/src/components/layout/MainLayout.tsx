import React, { useEffect } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import {
  LayoutDashboard,
  FileText,
  ShieldAlert,
  Shield,
  Zap,
  Unplug,
  Database,
  Settings,
  Activity,
  Server
} from "lucide-react";
import { useTabStore } from "@/stores/tab-store";
import type { TabType } from "@/stores/tab-store";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils"; // 假设你有这个工具函数，如果没有可以直接写 className字符串

// 定义菜单项配置
const MENU_ITEMS: { type: TabType; icon: any; label: string }[] = [
  { type: 'overview', icon: LayoutDashboard, label: 'Overview' },
  { type: 'logs', icon: FileText, label: 'Logs Query' },
  { type: 'rules', icon: Shield, label: 'Rules Center' },
  { type: 'incidents', icon: ShieldAlert, label: 'Incidents Center' },
  { type: 'automation', icon: Zap, label: 'Automation' },
  { type: 'ingest', icon: Database, label: 'Ingest' },
  { type: 'collectors', icon: Server, label: 'Collectors' },
  { type: 'connectors', icon: Unplug, label: 'Connectors' },
  { type: 'custom-logs', icon: FileText, label: 'Custom Tables' },
  { type: 'settings', icon: Settings, label: 'Settings' },
];

interface MainLayoutProps {
  children: React.ReactNode;
}

export function MainLayout({ children }: MainLayoutProps) {
  const { activeTabId, tabs, addTab, setActiveTab } = useTabStore();
  const location = useLocation();
  const navigate = useNavigate();

  // 1. ✅ Deep Linking: 初始化时根据 URL 自动打开 Tab
  useEffect(() => {
    // 移除开头的 '/'，例如 "/incidents" -> "incidents"
    // 如果是根路径 "/"，默认给 "overview"
    const path = location.pathname === '/' ? 'overview' : location.pathname.substring(1);

    // 简单的映射检查，防止无效路径报错
    const isValidTab = MENU_ITEMS.some(item => item.type === path);

    if (isValidTab) {
      addTab(path as TabType);
    } else if (path === 'login') {
      // login 由 App.tsx 路由处理，这里忽略
    } else {
      // 如果路径无法识别，默认回退到 Overview
      if (path !== '') addTab('overview');
    }
  }, []); // 仅挂载时执行一次

  // 获取当前激活的 Tab 类型，用于高亮菜单
  const currentTab = tabs.find(t => t.id === activeTabId);
  const currentTabType = currentTab?.type || 'overview';

  return (
    <div className="flex h-screen w-full bg-muted/40 overflow-hidden">
      {/* 侧边栏 Sidebar */}
      <aside className="fixed inset-y-0 left-0 z-10 w-64 border-r bg-background flex flex-col transition-all duration-300">
        {/* Logo 区域 */}
        <div className="flex h-16 shrink-0 items-center border-b px-6 gap-2 font-bold text-lg tracking-tight text-primary">
          <div className="p-1.5 bg-primary/10 rounded-lg">
            <Activity className="w-5 h-5" />
          </div>
          SOC Platform
        </div>

        {/* 菜单区域 */}
        <ScrollArea className="flex-1 py-4">
          <nav className="grid gap-1 px-3 text-sm font-medium">
            {MENU_ITEMS.map((item) => {
              const isActive = currentTabType === item.type;

              return (
                <Link
                  key={item.type}
                  to={`/${item.type}`} // ✅ 关键：生成真实的 href，支持右键打开
                  onClick={(e) => {
                    // ✅ 关键：左键点击阻止跳转，使用 SPA 逻辑
                    e.preventDefault();

                    // 1. 切换 Tab
                    addTab(item.type);

                    // 2. 手动更新 URL (不刷新页面)
                    window.history.pushState({}, "", `/${item.type}`);
                  }}
                  className={cn(
                    "flex items-center gap-3 rounded-lg px-3 py-2.5 transition-all hover:text-primary",
                    isActive
                      ? "bg-primary/10 text-primary font-semibold"
                      : "text-muted-foreground hover:bg-muted"
                  )}
                >
                  <item.icon className={cn("h-4 w-4", isActive ? "text-primary" : "text-muted-foreground")} />
                  {item.label}

                  {/* 可选：如果是 Incidents 且有未读数，可以在这里加个 Badge */}
                  {item.type === 'incidents' && (
                    <span className="ml-auto flex h-2 w-2 rounded-full bg-red-600" />
                  )}
                </Link>
              );
            })}
          </nav>
        </ScrollArea>

        {/* 底部用户信息区域 (可选) */}
        <div className="mt-auto border-t p-4">
          <Button variant="outline" className="w-full justify-start gap-2" onClick={() => navigate('/login')}>
            <span className="h-5 w-5 rounded-full bg-muted flex items-center justify-center text-[10px]">AD</span>
            <div className="flex flex-col items-start text-xs">
              <span className="font-semibold">Admin User</span>
              <span className="text-muted-foreground">SOC Analyst</span>
            </div>
          </Button>
        </div>
      </aside>

      {/* 主内容区域 */}
      <main className="flex flex-col flex-1 pl-64 h-full min-w-0 overflow-hidden bg-background">
        {/* 这里的 children 通常就是 <TabContent /> */}
        {children}
      </main>
    </div>
  );
}