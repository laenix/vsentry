import { useState } from "react";
import { CaseList } from "./CaseList";
import { CaseWorkspace } from "./CaseWorkspace";

export default function ForensicsPage() {
  // 控制当前视图：null 表示在案件大厅，有数字表示进入了具体的沙箱工作台
  const [activeCaseId, setActiveCaseId] = useState<number | null>(null);

  return (
    <div className="h-full w-full bg-background overflow-hidden">
      {activeCaseId === null ? (
        <CaseList onOpenCase={setActiveCaseId} />
      ) : (
        <CaseWorkspace 
          caseId={activeCaseId} 
          onBack={() => setActiveCaseId(null)} 
        />
      )}
    </div>
  );
}