import { LayoutDashboard } from "lucide-react"

export default function OverviewPage() {
  return (
    <div className="h-full w-full flex flex-col items-center justify-center text-muted-foreground gap-4 bg-background">
       <div className="p-6 rounded-full bg-muted/30 ring-1 ring-border/50">
          <LayoutDashboard className="h-12 w-12 opacity-20" />
       </div>
       <h2 className="text-xl font-semibold text-foreground tracking-tight">Logs Overview</h2>
       <p className="text-sm opacity-60">Module is under construction.</p>
    </div>
  )
}