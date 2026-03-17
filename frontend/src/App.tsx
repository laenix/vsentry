import React from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { MainLayout } from "@/components/layout/MainLayout";
import { TabContent } from "@/components/tabs/TabContent";
import { Toaster } from "@/components/ui/toaster"; // 或 sonner
import LoginPage from "@/pages/Login";
// ✅ 引入NewPage
import EditPlaybookPage from "@/pages/Automation/EditPlaybook";
import CRDEditorPage from "@/pages/Automation/CRDEditor";

// 1. 简单的Permission守卫Group件 (保持不变)
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
        {/* 2. Login页路由 (公开) */}
        <Route path="/login" element={<LoginPage />} />

        {/* ✅ 3. Automation Edit器路由 (独立全屏，不走 Tab System) */}
        {/* 注意：必须放在 "/*" 之前，否则会被 MainLayout 拦截 */}
        <Route 
          path="/automation/edit/:id" 
          element={
            <RequireAuth>
              {/* 这里不包 MainLayout，直接Render全屏Edit器 */}
              <EditPlaybookPage />
            </RequireAuth>
          } 
        />

        {/* ✅ CRD Editor 路由 (表单 + YAML 预览) */}
        <Route 
          path="/automation/crd/:id" 
          element={
            <RequireAuth>
              <CRDEditorPage />
            </RequireAuth>
          } 
        />

        {/* 4. 主Application路由 (受保护，带侧边栏和 Tab System) */}
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