// src/pages/Automation/index.tsx
import { useState } from "react";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";
import { useNavigate } from "react-router-dom";
import PlaybookList from "./PlaybookList";
import RunHistoryList from "./RunHistoryList";

export default function AutomationPage() {
  const [activeTab, setActiveTab] = useState("playbooks");
  const navigate = useNavigate();

  return (
    <div className="flex flex-col h-full bg-background p-6 gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Automation</h1>
          <p className="text-muted-foreground text-sm">Orchestrate security workflows and automate response.</p>
        </div>
        <Button onClick={() => navigate("/automation/edit/new")} className="bg-primary shadow-sm">
          <Plus className="w-4 h-4 mr-2" /> Create Playbook
        </Button>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
        <div className="flex justify-between items-center border-b">
          <TabsList className="bg-transparent p-0 h-9">
            <TabsTrigger 
              value="playbooks" 
              className="data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none px-4 bg-transparent"
            >
              Playbooks Library
            </TabsTrigger>
            {/* ✅ 新增：绑定关系管理标签 */}
            <TabsTrigger 
              value="bindings" 
              className="data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none px-4 bg-transparent"
            >
              Rule Bindings
            </TabsTrigger>
            <TabsTrigger 
              value="history" 
              className="data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none rounded-none px-4 bg-transparent"
            >
              Run History
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="playbooks" className="flex-1 mt-6 outline-none">
          <PlaybookList viewMode="list" />
        </TabsContent>

        {/* ✅ 新增：将 PlaybookList 以绑定管理模式运行 */}
        <TabsContent value="bindings" className="flex-1 mt-6 outline-none">
          <PlaybookList viewMode="binding" />
        </TabsContent>

        <TabsContent value="history" className="flex-1 mt-6 outline-none">
          <RunHistoryList />
        </TabsContent>
      </Tabs>
    </div>
  );
}