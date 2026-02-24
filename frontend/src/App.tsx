import React from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { MainLayout } from "@/components/layout/MainLayout";
import { TabContent } from "@/components/tabs/TabContent";
import { Toaster } from "@/components/ui/toaster"; // 或 sonner
import LoginPage from "@/pages/Login";
// ✅ 引入新页面
import EditPlaybookPage from "@/pages/Automation/EditPlaybook";

// 1. 简单的权限守卫组件 (保持不变)
const RequireAuth = ({ children }: { children: React.JSX.Element }) => {
  const token = localStorage.getItem("vsentry_token");
  if (!token) {
    return <Navigate to="/login" replace />;
  }
  return children;
};

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* 2. 登录页路由 (公开) */}
        <Route path="/login" element={<LoginPage />} />

        {/* ✅ 3. Automation 编辑器路由 (独立全屏，不走 Tab 系统) */}
        {/* 注意：必须放在 "/*" 之前，否则会被 MainLayout 拦截 */}
        <Route 
          path="/automation/edit/:id" 
          element={
            <RequireAuth>
              {/* 这里不包 MainLayout，直接渲染全屏编辑器 */}
              <EditPlaybookPage />
            </RequireAuth>
          } 
        />

        {/* 4. 主应用路由 (受保护，带侧边栏和 Tab 系统) */}
        <Route
          path="/*"
          element={
            <RequireAuth>
              <MainLayout>
                <TabContent />
                <Toaster />
              </MainLayout>
            </RequireAuth>
          }
        />
      </Routes>
    </BrowserRouter>
  );
}

export default App;