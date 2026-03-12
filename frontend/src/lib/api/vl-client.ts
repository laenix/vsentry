//   src/lib/api/vl-client.ts

/**
 * 基础Log行结构
 */
export interface LogRow {
  _time: string;
  _stream: string;
  _msg: string;
  [key: string]: any;
}

/** * 默认ReturnType (兼容 LogRow Sum聚合结果) 
 */
export type VLResult = LogRow | Record<string, any>;

/**
 * Hits Interface的Return结构
 */
export interface VLHitsResponse {
  hits: number;
}

/**
 * Get Token（与 vsentry-client.ts 保持一致）
 */
function getAuthHeader(): Record<string, string> {
  const token = localStorage.getItem("vsentry_token");
  if (token) {
    return { Authorization: `Bearer ${token}` };
  }
  return {};
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
 * 构建 Hits QueryParameter (Need step Parameter)
 */
function buildHitsParams(query: string, start?: string, end?: string, step?: string): URLSearchParams {
  const params = new URLSearchParams();
  params.append("query", query);
  if (start) params.append("start", start);
  if (end) params.append("end", end);
  // step - hits Time直方图，不传则自动计算
  if (step && step !== "auto") {
    params.append("step", step);
  }
  return params;
}

/**
 * 1. QueryLog或聚合Data (POST /select/logsql/query)
 * 支持泛型 T，用于自动推断聚合Query的Return结构
 * 注意：现在通过后端 API 代理到 VictoriaLogs
 */
export async function runVLQuery<T = VLResult>(
  query: string,
  limit: number | string = 1000,
  start?: string,
  end?: string
): Promise<T[]> {
  
  const body = buildParams(query, start, end, limit);

  //   使用 /api/select Path，后端会代理到 VictoriaLogs
  const response = await fetch("/api/select/logsql/query", {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
      ...getAuthHeader(), // Add - Auth
    },
    body: body,
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Query failed: ${response.status}`);
  }

  // Handle - Lines 流式Response (保持HighPerformance)
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
  
  // Handle最后一行 - (buffer.trim()) {
    try {
      result.push(JSON.parse(buffer));
    } catch (e) {}
  }

  return result;
}

/**
 * 2. Query命MediumTotal (POST /select/logsql/hits)
 * 通过后端 API 代理
 * 注意：VictoriaLogs /select/logsql/hits 在 v1.46.0 有兼容性问题，暂不抛出Error
 */
export async function runVLHits(
  query: string,
  start?: string,
  end?: string,
  step?: string
): Promise<number> {
  try {
    const body = buildHitsParams(query, start, end, step);
    const response = await fetch("/api/select/logsql/hits", {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
        ...getAuthHeader(),
      },
      body: body,
    });

    if (!response.ok) {
      // hits - ，不Block主Query
      return 0;
    }

    const data = await response.json();
    
    if (typeof data.hits === 'number') {
      return data.hits;
    } else if (Array.isArray(data.hits) && data.hits.length > 0) {
      return data.hits[0].total || 0;
    }
  } catch (e) {
    console.warn("Hits query failed:", e);
  }
  
  return 0;
}