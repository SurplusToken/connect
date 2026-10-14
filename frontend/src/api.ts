export interface User { email: string; nickname?: string }
export interface Group { id: number; name: string; platform: string; allow_messages_dispatch: boolean; status: string }
export interface Key { id: number; name: string; group_id: number; status: string; mask: string; expires_at?: string | null }
export interface Installation { executable: string; installed: boolean; version: string; version_status: string; source: string; detail: string; family?: string; name?: string }
export interface Detection extends Installation { desktop?: Installation }
export interface DetectedClient { detection?: Detection; id: string; path: string; installed: boolean; supported: boolean; version: string; reason: string }
export interface Session { session: string; user: User; groups: Group[]; keys: Key[]; clients: DetectedClient[]; requires_2fa?: boolean; fixture?: boolean; warning?: string }
export interface Catalog { models: string[]; group: Group; messages: boolean }
export interface Change { field: string; before: string; after: string }
export interface Plan { id: string; client: string; path: string; changes: Change[]; noop: boolean; template?: string; model_count?: number }
export interface ConfigRecord { id: string; client: string; path: string; at: string; state: string; operation: string; changes: Change[]; backup: string }
export const native = location.protocol === "wails:" || location.hostname === "wails.localhost";
function remembered(): string { try { return native ? "" : localStorage.getItem("connect-session") || ""; } catch { return ""; } }
let session = remembered();
export function hasRememberedSession() { return !!session; }
export function clearRememberedSession() { session = ""; try { localStorage.removeItem("connect-session"); } catch {} }
export class APIError extends Error { constructor(message: string, public status: number) { super(message); } }

export async function request<T>(action: string, body: object = {}): Promise<T> {
  let status: number;
  let data: T & {error?: string; session?: string};
  if (native) {
    const { Call } = await import("@wailsio/runtime");
    const pending = Call.ByName("surplustoken-connect/internal/httpapi.Bridge.Request", action, session, JSON.stringify(body));
    const timeout = setTimeout(() => pending.cancel(), 45000);
    try {
      const response = await pending as {status:number; data:typeof data};
      status = response.status; data = response.data;
    } catch { throw new Error("原生连接服务不可用或请求超时，请重新打开 Connect。"); }
    finally { clearTimeout(timeout); }
  } else {
    let response: Response;
    try { response = await fetch(`/connect/${action}`, {method:"POST",headers:{"Content-Type":"application/json","X-Connect-Request":"1","X-Connect-Session":session},body:JSON.stringify(body),signal:AbortSignal.timeout(45000)}); }
    catch { throw new Error("本地连接服务不可用或请求超时，请检查网络和 Connect 服务。"); }
    try { data = await response.json(); } catch { throw new Error("连接服务响应异常，请确认本地 API 已启动。"); }
    status = response.status;
  }
  if (status < 200 || status >= 300) throw new APIError(data.error || "操作失败",status);
  if (data.session) { session = data.session; if (!native) { try { localStorage.setItem("connect-session",session); } catch {} } }
  if (action === "logout") clearRememberedSession();
  return data;
}
