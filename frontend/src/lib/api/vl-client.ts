// src/lib/api/vl-client.ts

/**
 * 基础日志行结构
 */
export interface LogRow {
  _time: string;
  _stream: string;
  _msg: string;
  [key: string]: any;
}

/** * 默认返回类型 (兼容 LogRow 和聚合结果) 
 */
export type VLResult = LogRow | Record<string, any>;

/**
 * Hits 接口的返回结构
 */
export interface VLHitsResponse {
  hits: number;
}

/**
 * 核心：构建 Form Data (复用逻辑)
 */
function buildParams(query: string, start?: string, end?: string, limit?: number | string): URLSearchParams {
  const params = new URLSearchParams();
  params.append("query", query);
  if (limit !== undefined) params.append("limit", String(limit));
  if (start) params.append("start", start);
  if (end) params.append("end", end);
  return params;
}

/**
 * 1. 查询日志或聚合数据 (POST /select/logsql/query)
 * 支持泛型 T，用于自动推断聚合查询的返回结构
 * 注意：现在通过后端 API 代理到 VictoriaLogs
 */
export async function runVLQuery<T = VLResult>(
  query: string,
  limit: number | string = 1000,
  start?: string,
  end?: string
): Promise<T[]> {
  
  const body = buildParams(query, start, end, limit);

  // 使用 /api/select 路径，后端会代理到 VictoriaLogs
  const response = await fetch("/api/select/logsql/query", {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body: body,
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Query failed: ${response.status}`);
  }

  // 处理 JSON Lines 流式响应 (保持高性能)
  const reader = response.body?.getReader();
  const decoder = new TextDecoder();
  const result: T[] = [];
  let buffer = "";

  if (reader) {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop() || "";

      for (const line of lines) {
        if (line.trim()) {
          try {
            result.push(JSON.parse(line));
          } catch (e) {
            console.error("JSONL Parse Error:", line);
          }
        }
      }
    }
  }
  
  // 处理最后一行
  if (buffer.trim()) {
    try {
      result.push(JSON.parse(buffer));
    } catch (e) {}
  }

  return result;
}

/**
 * 2. 查询命中总数 (POST /select/logsql/hits)
 * 通过后端 API 代理
 */
export async function runVLHits(
  query: string,
  start?: string,
  end?: string
): Promise<number> {
  const body = buildParams(query, start, end);

  const response = await fetch("/api/select/logsql/hits", {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: body,
  });

  if (!response.ok) {
    throw new Error(`Hits check failed: ${response.status}`);
  }

  const data = await response.json();
  
  // ✅ 核心修复：兼容多种返回格式
  if (typeof data.hits === 'number') {
    return data.hits;
  } else if (Array.isArray(data.hits) && data.hits.length > 0) {
    // 处理 { hits: [{ total: 1927, ... }] } 格式
    return data.hits[0].total || 0;
  }
  
  return 0;
}