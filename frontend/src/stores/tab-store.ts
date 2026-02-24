import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { v4 as uuidv4 } from 'uuid'

// 1. 定义 Tab 类型，包含所有模块
export type TabType = 
  | 'overview'     // 总览
  | 'logs'         // 日志查询 (Logs Query)
  | 'incidents'    // 事件中心 (Incidents Center)
  | 'rules'        // 规则中心
  | 'automation'   // 自动化
  | 'connectors'   // 集成
  | 'ingest'       // 接入点管理
  | 'collectors'   // 采集器 (预留)
  | 'custom-logs'  // 自定义日志
  | 'settings'     // 设置

export interface Tab {
  id: string
  type: TabType
  title: string
  data?: {
    query?: string     // 专门用于存储日志查询语句
    [key: string]: any
  }
}

interface TabState {
  tabs: Tab[]
  activeTabId: string | null
  addTab: (type: TabType, title?: string, data?: any) => void
  removeTab: (id: string) => void
  setActiveTab: (id: string) => void
  updateTabData: (id: string, newData: any) => void
  closeAll: () => void
}

// 自动生成标题映射
const defaultTitles: Record<TabType, string> = {
  overview: 'Overview',
  logs: 'Logs Query',
  incidents: 'Incidents Center',
  rules: 'Rules Center',
  automation: 'Automation',
  connectors: 'Connectors',
  collectors: 'Collectors',
  ingest: 'Ingest',
  'custom-logs': 'Custom Logs',
  settings: 'Settings'
}

export const useTabStore = create<TabState>()(
  persist(
    (set, get) => ({
      tabs: [],
      activeTabId: null,

      addTab: (type, title, data) => {
        const state = get()
        
        // 模式：除了 logs 可能需要多开，其他模块通常只开一个
        // 如果你希望 logs 也能多开，把 type === 'logs' 这个判断去掉即可
        const existingTab = state.tabs.find(t => t.type === type)

        if (existingTab) {
          // 如果标签已存在，则激活它
          set({ activeTabId: existingTab.id })
          
          // 如果传入了新的 data（例如从 Incident 跳转过来的查询语句），则更新该标签
          if (data) {
            get().updateTabData(existingTab.id, data)
          }

          // ✅ 同步 URL：如果是日志查询，更新浏览器地址栏
          if (type === 'logs' && data?.query) {
            const url = new URL(window.location.href)
            url.searchParams.set('q', encodeURIComponent(data.query))
            window.history.pushState({}, '', url.pathname + url.search)
          }
          return
        }

        // 创建新标签
        const newTab: Tab = {
          id: uuidv4(),
          type,
          title: title || defaultTitles[type] || type,
          data: data || {}
        }

        set((state) => ({
          tabs: [...state.tabs, newTab],
          activeTabId: newTab.id
        }))

        // ✅ 新开标签时同步 URL
        if (type === 'logs' && data?.query) {
          const url = new URL(window.location.href)
          url.searchParams.set('q', encodeURIComponent(data.query))
          window.history.pushState({}, '', url.pathname + url.search)
        }
      },

      removeTab: (id) => set((state) => {
        const newTabs = state.tabs.filter((t) => t.id !== id)
        let newActiveId = state.activeTabId

        if (id === state.activeTabId) {
          if (newTabs.length > 0) {
            // 激活前一个标签
            const deletedIndex = state.tabs.findIndex(t => t.id === id)
            newActiveId = newTabs[deletedIndex] ? newTabs[deletedIndex].id : newTabs[newTabs.length - 1].id
          } else {
            newActiveId = null
          }
        }
        return { tabs: newTabs, activeTabId: newActiveId }
      }),

      setActiveTab: (id) => set({ activeTabId: id }),

      updateTabData: (id, newData) => set((state) => ({
        tabs: state.tabs.map((t) => 
          t.id === id ? { ...t, data: { ...t.data, ...newData } } : t
        )
      })),

      closeAll: () => set({ tabs: [], activeTabId: null }),
    }),
    {
      name: 'vsentry-tabs-storage', // 持久化存储，防止 F5 刷新后标签全丢
    }
  )
)