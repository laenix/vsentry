import { create } from 'zustand'
import { v4 as uuidv4 } from 'uuid'

// 1. 定义 Tab Type，包含所有模块
export type TabType = 
  | 'overview'     // 总览
  | 'logs'         // LogQuery
  | 'incidents'    // EventMedium心
  | 'investigation' // InvestigationMedium心
  | 'forensics'     // ForensicsMedium心
  | 'rules'        // RuleMedium心
  | 'automation'   // Automation
  | 'connectors'   // 集成
  | 'ingest'       // 接入点Manage
  | 'collectors'   // Collect器
  | 'custom-logs'  // 自定义Log
  | 'settings'     // Settings

export interface Tab {
  id: string
  type: TabType
  title: string
  data?: {
    query?: string     
    incident_id?: number // ✅ New增：Allow传递EventID作为上下文
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
  investigation: 'Investigation',
  forensics: 'Forensics Sandbox',
  rules: 'Rules Center',
  automation: 'Automation',
  connectors: 'Connectors',
  collectors: 'Collectors',
  ingest: 'Ingest',
  'custom-logs': 'Custom Logs',
  settings: 'Settings'
}

export const useTabStore = create<TabState>()(
  (set, get) => ({
    tabs: [],
    activeTabId: null,

    addTab: (type, title, data) => {
      const state = get()
      
      const existingTab = state.tabs.find(t => t.type === type)
      if (existingTab) {
        set({ activeTabId: existingTab.id })
        if (data) get().updateTabData(existingTab.id, data)
        return
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
          // 激活ago一个标签
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
  })
)
