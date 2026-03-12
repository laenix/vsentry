import React from 'react';
import { Button } from '@/components/ui/button';
import { Trash2, Plus } from 'lucide-react';
import { MonacoInput } from './MonacoInput';

interface KeyValuePair {
  key: string;
  value: string;
}

interface KeyValueListProps {
  items: Record<string, string>;
  onChange: (items: Record<string, string>) => void;
  keyPlaceholder?: string;
  valuePlaceholder?: string;
}

export function KeyValueList({ items, onChange, keyPlaceholder = "Key", valuePlaceholder = "Value" }: KeyValueListProps) {
  // 将 - 转为 Array 进行Edit
  const entries = Object.entries(items).map(([key, value]) => ({ key, value }));

  const handleChange = (index: number, field: 'key' | 'value', newValue: string) => {
    const newEntries = [...entries];
    newEntries[index][field] = newValue;
    // 转回 - const newObj = newEntries.reduce((acc, curr) => {
      if (curr.key) acc[curr.key] = curr.value;
      return acc;
    }, {} as Record<string, string>);
    onChange(newObj);
  };

  const handleDelete = (index: number) => {
    const newEntries = entries.filter((_, i) => i !== index);
    const newObj = newEntries.reduce((acc, curr) => {
      if (curr.key) acc[curr.key] = curr.value;
      return acc;
    }, {} as Record<string, string>);
    onChange(newObj);
  };

  const handleAdd = () => {
    // Add一个临时唯一的 - 防止 React Key 冲突，实际Save时会Filter空 Key
    const newEntries = [...entries, { key: '', value: '' }];
    // 这里不Need立即 - ，等UserInput Key 之后再生效，避免产生空 Key
    // 但为了 - 渲染，我们Need维护一个本地Status，或者在这里简单Handle：
    //   简单起见，这里我们不立即写回 onChange，而是依赖上面的 handleChange
    // 但为了让 - 刷出来，我们Need hack 一下：
    // 实际项目Medium建议使用 - Manage array，最后 useEffect Sync回 object
    //   这里简化Handle：Allow暂时有空 key 的对象传递（注意：Object key Cannot重复为空）
    //   为了稳健，KeyValueList 最好内部维护 Array state。
  };
  
  //   简化版：我们直接渲染 items + 一个空行
  const renderRows = [...entries, { key: '', value: '', isNew: true }];

  const handleRowChange = (index: number, key: string, value: string) => {
     const newEntries = [...entries];
     if (index === entries.length) {
        // New增行 - .push({ key, value });
     } else {
        // 修改行 - [index] = { key, value };
     }
     
     const newObj = newEntries.reduce((acc, curr) => {
       if (curr.key) acc[curr.key] = curr.value;
       return acc;
     }, {} as Record<string, string>);
     onChange(newObj);
  };

  return (
    <div className="space-y-2">
      <div className="flex text-[10px] font-bold text-muted-foreground uppercase px-1">
        <div className="flex-1">{keyPlaceholder}</div>
        <div className="flex-1">{valuePlaceholder}</div>
        <div className="w-8"></div>
      </div>
      
      {renderRows.map((item, index) => (
        <div key={index} className="flex gap-2 items-start">
          <div className="flex-1">
            <input 
               className="w-full text-xs px-2 py-1.5 border rounded bg-background"
               placeholder={keyPlaceholder}
               value={item.key}
               onChange={(e) => handleRowChange(index, e.target.value, item.value)}
            />
          </div>
          <div className="flex-1">
            <MonacoInput 
              value={item.value} 
              onChange={(val) => handleRowChange(index, item.key, val)}
              height="28px"
            />
          </div>
          <div className="w-8 flex justify-center pt-1">
            {!item.isNew && (
              <button onClick={() => handleDelete(index)} className="text-muted-foreground hover:text-destructive">
                <Trash2 className="w-4 h-4" />
              </button>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}