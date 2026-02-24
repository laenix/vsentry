import { X, Activity } from "lucide-react"
import { useTabStore } from "@/stores/tab-store"
import { cn } from "@/lib/utils"

// 1. 导入所有页面组件 (确保你已经在对应的 pages 目录下创建了 index.tsx)
import LogsPage from "@/pages/Logs"
import DashboardPage from "@/pages/Dashboard"
import ConnectorsPage from "@/pages/Connectors"
import RulesPage from "@/pages/Rules"
import IncidentsPage from "@/pages/Incidents"
import AutomationPage from "@/pages/Automation"
import IngestPage from "@/pages/Ingest"
import CollectorsPage from "@/pages/Collectors"
import CustomLogsPage from "@/pages/CustomLogs"
import SettingsPage from "@/pages/Settings"

export function TabContent() {
  const { tabs, activeTabId, setActiveTab, removeTab } = useTabStore()

  // 2. 根据 Tab 类型映射对应的页面组件
  const renderContent = (type: string) => {
    switch (type) {
      case 'dashboard':    return <DashboardPage />
      case 'logs':        return <LogsPage />
      case 'connectors':  return <ConnectorsPage />
      case 'rules':       return <RulesPage />
      case 'incidents':   return <IncidentsPage />
      case 'automation':  return <AutomationPage />
      case 'ingest':      return <IngestPage />
      case 'collectors':  return <CollectorsPage />
      case 'custom-logs': return <CustomLogsPage />
      case 'settings':    return <SettingsPage />
      default:            return <DashboardPage />
    }
  }

  return (
    <div className="flex flex-col h-full w-full bg-background">
      {/* 顶部浏览器风格 Tab 栏 */}
      <div className="flex items-center border-b border-border bg-muted/10 overflow-x-auto no-scrollbar h-[36px] flex-none">
        {tabs.map((tab) => (
          <div
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              "group flex items-center gap-2 px-3 h-full border-r border-border/40 text-xs cursor-pointer select-none min-w-[100px] max-w-[180px] transition-all relative",
              activeTabId === tab.id 
                ? "bg-background text-foreground font-medium before:absolute before:top-0 before:left-0 before:right-0 before:h-[2px] before:bg-primary" 
                : "text-muted-foreground hover:bg-muted/20 hover:text-foreground"
            )}
          >
            <span className="truncate flex-1">{tab.title}</span>
            <div
              role="button"
              onClick={(e) => {
                e.stopPropagation() 
                removeTab(tab.id)
              }}
              className={cn(
                "opacity-0 group-hover:opacity-100 p-0.5 rounded-sm hover:bg-muted-foreground/20 transition-opacity",
                activeTabId === tab.id && "opacity-100" 
              )}
            >
              <X className="h-3 w-3" />
            </div>
          </div>
        ))}
      </div>

      {/* 核心 Keep-Alive 容器 */}
      <div className="flex-1 overflow-hidden relative">
        {tabs.map((tab) => (
          <div
            key={tab.id}
            className={cn(
              "absolute inset-0 h-full w-full overflow-hidden bg-background",
              // 关键：通过 display: none 隐藏非活跃标签，从而保持页面内的状态（如编辑器光标、滚动位置）
              activeTabId === tab.id ? "z-10 block" : "z-0 hidden"
            )}
          >
            {renderContent(tab.type)}
          </div>
        ))}
        
        {/* 空状态 */}
        {tabs.length === 0 && (
          <div className="h-full flex flex-col items-center justify-center text-muted-foreground gap-4 bg-muted/5">
             <div className="p-4 rounded-full bg-muted/20">
                <Activity className="h-8 w-8 opacity-20" />
             </div>
             <p className="text-sm">Select a module from the sidebar to begin.</p>
          </div>
        )}
      </div>
    </div>
  )
}