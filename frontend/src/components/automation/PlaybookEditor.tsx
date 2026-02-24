import React, { useState, useCallback, useMemo, useRef } from 'react';
import {
    ReactFlow,
    ReactFlowProvider,
    addEdge,
    useNodesState,
    useEdgesState,
    Controls,
    Background,
    MiniMap,
    Handle,
    Position,
    BackgroundVariant,
    useReactFlow
} from '@xyflow/react';

import type { Connection, Edge, Node, NodeMouseHandler } from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { Button } from "@/components/ui/button";
import {
    Save, Play, ChevronLeft, Zap, Terminal,
    AlertTriangle, Globe, Mail, Shield, Loader2, Calculator
} from "lucide-react";
import { TestRunDialog } from "./TestRunDialog";
import ConfigPanel from './ConfigPanel';
import { automationService } from "@/services/automation";
import { toast } from 'sonner';
import { useNavigate } from 'react-router-dom';

// --- 自定义节点 UI 定义 ---
const icons: any = {
    trigger: Zap,
    action: Terminal,
    condition: AlertTriangle,
    expression: Calculator,
    http_request: Globe,
    send_email: Mail,
    block_ip: Shield
};

const CustomNode = ({ data, selected }: any) => {
    const Icon = icons[data.type] || icons[data.icon] || Terminal;
    return (
        <div className={`shadow-sm rounded-md bg-card border min-w-[180px] transition-all ${selected ? 'border-primary ring-1 ring-primary shadow-md' : 'border-border'}`}>
            {data.type !== 'trigger' && <Handle type="target" position={Position.Top} className="w-2 h-2 bg-muted-foreground" />}
            <div className="flex items-center gap-3 p-3">
                <div className={`p-2 rounded-md ${data.type === 'trigger' ? 'bg-amber-100 text-amber-700' :
                        data.type === 'expression' ? 'bg-purple-100 text-purple-700' :
                            'bg-slate-100 text-slate-700'
                    }`}>
                    <Icon className="w-4 h-4" />
                </div>
                <div className="flex flex-col">
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">{data.type?.replace('_', ' ')}</span>
                    <span className="text-sm font-medium leading-none truncate max-w-[140px]">{data.label}</span>
                </div>
            </div>
            {data.type === 'condition' ? (
                <div className="flex justify-between px-4 pb-2 -mb-3 relative z-10">
                    <div className="relative">
                        <Handle type="source" position={Position.Bottom} id="true" className="w-2 h-2 bg-emerald-500 !left-2" />
                        <span className="text-[9px] text-emerald-600 font-bold absolute top-2 left-0">YES</span>
                    </div>
                    <div className="relative">
                        <Handle type="source" position={Position.Bottom} id="false" className="w-2 h-2 bg-red-500 !left-auto !right-2" />
                        <span className="text-[9px] text-red-600 font-bold absolute top-2 right-0">NO</span>
                    </div>
                </div>
            ) : (
                <Handle type="source" position={Position.Bottom} className="w-2 h-2 bg-primary" />
            )}
        </div>
    );
};
const nodeTypes = { custom: CustomNode };

// --- 侧边栏拖拽项 ---
function DraggableItem({ icon: Icon, type, label, color }: { icon: any, type: string, label: string, color: string }) {
    const onDragStart = (event: React.DragEvent, nodeType: string, nodeLabel: string) => {
        event.dataTransfer.setData('application/reactflow/type', nodeType);
        event.dataTransfer.setData('application/reactflow/label', nodeLabel);
        event.dataTransfer.effectAllowed = 'move';
    };
    return (
        <div
            className="p-2 border bg-card rounded text-sm flex items-center gap-2 cursor-grab hover:border-primary/50 transition-colors select-none"
            draggable
            onDragStart={(e) => onDragStart(e, type, label)}
        >
            <Icon className={`w-4 h-4 ${color}`} /> {label}
        </div>
    );
}

// --- 编辑器主内容 ---
interface PlaybookEditorProps {
    playbookId: string;
    playbookName?: string;
    initialNodes?: Node[];
    initialEdges?: Edge[];
    onBack: () => void;
}

function PlaybookEditorContent({ playbookId, initialNodes, initialEdges, onBack, playbookName }: PlaybookEditorProps) {
    const navigate = useNavigate();
    const reactFlowWrapper = useRef<HTMLDivElement>(null);
    const { screenToFlowPosition } = useReactFlow();

    const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes || []);
    const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges || []);
    const [testDialogOpen, setTestDialogOpen] = useState(false);
    const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
    const [isSaving, setIsSaving] = useState(false);
    const [isRunning, setIsRunning] = useState(false);
    const [executionContext, setExecutionContext] = useState<Record<string, any>>({});

    // ✅ 核心：提取当前剧本的 Trigger 配置
    const triggerNode = nodes.find(n => n.data.type === 'trigger');
    const triggerConfig = triggerNode?.data.config || {};

    const onConnect = useCallback((params: Connection) => setEdges((eds) => addEdge(params, eds)), [setEdges]);

    const selectedNode = useMemo(() => nodes.find((n) => n.id === selectedNodeId) || null, [nodes, selectedNodeId]);
    const onNodeClick: NodeMouseHandler = useCallback((_, node) => setSelectedNodeId(node.id), []);
    const onPaneClick = useCallback(() => setSelectedNodeId(null), []);

    const handleUpdateNode = useCallback((id: string, newData: any) => {
        setNodes((nds) =>
            nds.map((node) => {
                if (node.id === id) {
                    return { ...node, data: { ...node.data, ...newData } };
                }
                return node;
            })
        );
    }, [setNodes]);

    const onDragOver = useCallback((event: React.DragEvent) => {
        event.preventDefault();
        event.dataTransfer.dropEffect = 'move';
    }, []);

    const onDrop = useCallback((event: React.DragEvent) => {
        event.preventDefault();
        const type = event.dataTransfer.getData('application/reactflow/type');
        const label = event.dataTransfer.getData('application/reactflow/label');
        if (!type) return;

        const position = screenToFlowPosition({ x: event.clientX, y: event.clientY });
        const newNode: Node = {
            id: `${type}_${Date.now()}`,
            type: 'custom',
            position,
            data: { label, type, config: {} },
        };
        setNodes((nds) => nds.concat(newNode));
        setSelectedNodeId(newNode.id);
    }, [screenToFlowPosition, setNodes]
    );

    const handleSave = async () => {
        setIsSaving(true);
        const payload = {
            name: playbookName || "Untitled Playbook",
            definition: { nodes, edges, viewport: { x: 0, y: 0, zoom: 1 } }
        };
        try {
            if (playbookId === 'new') {
                const res = await automationService.create(payload);
                if (res.code === 200) {
                    toast.success("Playbook Created!");
                    navigate(`/automation/edit/${res.data.ID}`, { replace: true });
                }
            } else {
                await automationService.update(playbookId, payload);
                toast.success("Playbook Saved Successfully");
            }
        } catch (error) {
            toast.error("Failed to save playbook");
        } finally {
            setIsSaving(false);
        }
    };

    // ✅ 修正后的 Test Run 逻辑：支持手动触发 Mock 数据
    const handleTestRun = async (testData: { incident_id?: number; mock_data?: any }) => {
        if (playbookId === 'new') {
            toast.warning("Save the playbook before running tests.");
            return;
        }

        try {
            setIsRunning(true);
            const payload: any = { dry_run: false };

            if (testData.incident_id) payload.incident_id = testData.incident_id;
            if (testData.mock_data) payload.mock_context = testData.mock_data;

            const res = await automationService.runTest(playbookId, payload); //

            if (res.code === 200) {
                const execId = res.data.execution_id;
                toast.success(`Run #${execId} started. Polling for results...`);

                let attempts = 0;
                const poll = setInterval(async () => {
                    attempts++;
                    const detailRes = await automationService.getExecutionDetail(execId); //
                    if (detailRes.data && detailRes.data.status !== 'running') {
                        clearInterval(poll);
                        setIsRunning(false);
                        setExecutionContext(detailRes.data.logs || {}); //
                        toast.success("Run finished. Context injected for autocomplete.");
                    }
                    if (attempts > 20) {
                        clearInterval(poll);
                        setIsRunning(false);
                        toast.error("Execution timeout.");
                    }
                }, 1000);
            }
        } catch (error) {
            setIsRunning(false);
            toast.error("Failed to start test run.");
        }
    };

    return (
        <div className="h-full flex flex-col bg-background">
            <div className="h-14 border-b px-4 flex items-center justify-between bg-card z-10 shrink-0">
                <div className="flex items-center gap-4">
                    <Button variant="ghost" size="sm" onClick={onBack} className="gap-1 pl-0 text-muted-foreground">
                        <ChevronLeft className="w-4 h-4" /> Back
                    </Button>
                    <div className="flex flex-col">
                        <span className="text-sm font-bold">{playbookName || "Untitled Playbook"}</span>
                        <span className="text-[10px] text-muted-foreground">Draft • {nodes.length} nodes</span>
                    </div>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" size="sm" onClick={handleSave} disabled={isSaving} className="gap-2">
                        {isSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
                        Save
                    </Button>
                    <Button
                        size="sm"
                        onClick={() => setTestDialogOpen(true)}
                        disabled={isRunning}
                        className="bg-emerald-600 hover:bg-emerald-700 gap-2"
                    >
                        {isRunning ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4" />}
                        Test Run
                    </Button>
                </div>
            </div>

            <div className="flex-1 flex overflow-hidden">
                <div className="w-56 border-r bg-muted/10 p-3 flex flex-col gap-4 overflow-y-auto shrink-0">
                    <div className="space-y-2">
                        <h4 className="text-xs font-bold text-muted-foreground uppercase tracking-tight">Triggers</h4>
                        <DraggableItem icon={Zap} type="trigger" label="Incident Created" color="text-amber-500" />
                        <DraggableItem icon={Zap} type="trigger" label="Manual/Schedule" color="text-amber-500" />
                    </div>
                    <div className="space-y-2">
                        <h4 className="text-xs font-bold text-muted-foreground uppercase tracking-tight">Logic</h4>
                        <DraggableItem icon={AlertTriangle} type="condition" label="Condition (If/Else)" color="text-purple-500" />
                        <DraggableItem icon={Calculator} type="expression" label="Expression (Eval)" color="text-blue-600" />
                    </div>
                    <div className="space-y-2">
                        <h4 className="text-xs font-bold text-muted-foreground uppercase tracking-tight">Actions</h4>
                        <DraggableItem icon={Globe} type="http_request" label="HTTP Request" color="text-blue-500" />
                        <DraggableItem icon={Shield} type="block_ip" label="Block IP" color="text-red-500" />
                        <DraggableItem icon={Mail} type="send_email" label="Send Email" color="text-indigo-500" />
                    </div>
                </div>

                <div className="flex-1 h-full bg-slate-50 relative" ref={reactFlowWrapper}>
                    <ReactFlow
                        nodes={nodes}
                        edges={edges}
                        onNodesChange={onNodesChange}
                        onEdgesChange={onEdgesChange}
                        onConnect={onConnect}
                        onNodeClick={onNodeClick}
                        onPaneClick={onPaneClick}
                        onDrop={onDrop}
                        onDragOver={onDragOver}
                        nodeTypes={nodeTypes}
                        fitView
                        attributionPosition="bottom-right"
                    >
                        <Background gap={16} size={1} color="#cbd5e1" variant={BackgroundVariant.Dots} />
                        <Controls className="bg-white border-none shadow-md text-black" />
                        <MiniMap className='bg-white border rounded shadow-sm' nodeColor="#94a3b8" />
                    </ReactFlow>
                </div>

                <div className="w-80 border-l bg-card shrink-0 hidden lg:block transition-all z-20 shadow-xl">
                    <ConfigPanel
                        selectedNode={selectedNode}
                        onUpdateNode={handleUpdateNode}
                        executionContext={executionContext} //
                    />
                </div>
            </div>

            <TestRunDialog
                open={testDialogOpen}
                onOpenChange={setTestDialogOpen}
                onRun={handleTestRun} //
                triggerConfig={triggerConfig} //
            />
        </div>
    );
}

export default function PlaybookEditor(props: PlaybookEditorProps) {
    return (
        <ReactFlowProvider>
            <PlaybookEditorContent {...props} />
        </ReactFlowProvider>
    );
}