import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { v4 as uuidv4 } from 'uuid'

// 1. 定义 Tab 类型，包含所有模块
export type TabType = 
  | 'overview'     // 总览
  | 'logs'         // 日志查询 (Logs Query)
  | 'incidents'    // 事件中心 (Incidents Center)
  | 'investigation' // ✅ 新增：调查中心
  | 'forensics'     // ✅ 新增：取证中心
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
    query?: string     
    incident_id?: number // ✅ 新增：允许传递事件ID作为上下文
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
  investigation: 'Investigation', // ✅ 新增
  forensics: 'Forensics Sandbox', // ✅ 新增
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
        
        // 允许调查(investigation)和日志(logs)多开
        // 因为分析师可能同时打开针对不同 Incident 的调查画布
        if (type !== 'logs' && type !== 'investigation') {
          const existingTab = state.tabs.find(t => t.type === type)
          if (existingTab) {
            set({ activeTabId: existingTab.id })
            if (data) get().updateTabData(existingTab.id, data)
            return
          }
        }

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