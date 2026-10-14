import { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import { Button } from "@cloudflare/kumo/components/button";
import { Select } from "@cloudflare/kumo/components/select";
import { Table } from "@cloudflare/kumo/components/table";
import { Tabs } from "@cloudflare/kumo/components/tabs";
import { Popover } from "@cloudflare/kumo/components/popover";
import { Collapsible } from "@cloudflare/kumo/components/collapsible";
import { Dialog } from "@cloudflare/kumo/components/dialog";
import { Input } from "@cloudflare/kumo/components/input";
import { InputGroup } from "@cloudflare/kumo/components/input-group";
import {
  SquaresFourIcon,
  SignOutIcon,
  ClockCounterClockwiseIcon,
  GearIcon,
  ArrowRightIcon,
  KeyIcon,
  PlusIcon,
  ArrowClockwiseIcon,
  TerminalIcon,
  XIcon,
  EyeIcon,
  EyeSlashIcon,
} from "@phosphor-icons/react";
import "@cloudflare/kumo/styles/standalone";
import "./style.css";
import { clients } from "./clients";
import { native, APIError, hasRememberedSession, clearRememberedSession, request, type User, type Session, type Group, type Key, type DetectedClient, type Catalog, type Plan, type ConfigRecord } from "./api";

const navigation = [
  { id: "clients", label: "客户端", icon: SquaresFourIcon },
  { id: "keys", label: "API Key", icon: KeyIcon },
  { id: "history", label: "配置记录", icon: ClockCounterClockwiseIcon },
  { id: "settings", label: "设置", icon: GearIcon },
];
const logos: Record<string, string> = {
  codex: "openai",
  claude: "claude-color",
  opencode: "opencode",
  kimi: "kimi-color",
};
function App() {
  const [user, setUser] = useState<User | null>(null);
  const [account, setAccount] = useState(() => { try { return localStorage.getItem("connect-account") || ""; } catch { return ""; } });
  const [page, setPage] = useState("clients");
  const [mode, setMode] = useState("permanent");
  const [configuring, setConfiguring] = useState<string | null>(null);
  const targets = configuring ? [configuring] : [];
  const [models, setModels] = useState<Record<string, string>>({});
  const [preview, setPreview] = useState(false);
  const [message, setMessage] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [fixtureMode, setFixtureMode] = useState(false);
  const [busy, setBusy] = useState(true);
  const [recovering, setRecovering] = useState(true);
  const [restoreFailed, setRestoreFailed] = useState(false);
  const [otpRequired, setOtpRequired] = useState(false);
  const [otp, setOtp] = useState("");
  const [loginError, setLoginError] = useState("");
  const [groups, setGroups] = useState<Group[]>([]);
  const [keys, setKeys] = useState<Key[]>([]);
  const [keyId, setKeyId] = useState<string | null>(null);
  const [createGroup, setCreateGroup] = useState<string | null>(null);
  const [keySearch, setKeySearch] = useState("");
  const [keyGroupFilter, setKeyGroupFilter] = useState<string | null>("all");
  const [keyStatusFilter, setKeyStatusFilter] = useState<string | null>("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [detected, setDetected] = useState<DetectedClient[]>([]);
  const [clientKeys, setClientKeys] = useState<Record<string,string>>({});
  const [catalogs, setCatalogs] = useState<Record<string,Catalog>>({});
  const [plans, setPlans] = useState<Plan[]>([]);
  const [records, setRecords] = useState<ConfigRecord[]>([]);
  const [restoring, setRestoring] = useState(false);
  const [configurationError, setConfigurationError] = useState("");
  const errorText = (error: unknown) => error instanceof Error ? error.message : "操作失败";
  const canConfigure = (id: string) => !!detected.find(c => c.id === id)?.supported;
  const usableKey = (key: Key) => key.status === "active" && (!key.expires_at || new Date(key.expires_at).getTime() > Date.now());
  const compatibleKeys = (id: string) => keys.filter(key => usableKey(key) && (id !== "claude" || groups.some(group => group.id === key.group_id && (group.platform === "anthropic" || group.allow_messages_dispatch))) && (id !== "opencode" || groups.some(group => group.id === key.group_id && ["openai","anthropic","deepseek","kimi"].includes(group.platform))));
  const selectKey = async (id: string) => {
    setKeyId(id);
    setClientKeys(previous => {
      const next = {...previous};
      for (const client of clients) if (!client.locked && !next[client.id] && compatibleKeys(client.id).some(key => String(key.id) === id)) next[client.id] = id;
      return next;
    });
  };
  const acceptSession = async (data: Session) => {
    setOtpRequired(false); setOtp(""); setFixtureMode(!!data.fixture); setUser(data.user); setGroups(data.groups); setKeys(data.keys); setDetected(data.clients); setPage("clients");
    try { localStorage.setItem("connect-account",data.user.email); } catch {}
    setAccount(data.user.email);
    const available = data.keys.filter(usableKey);
    setKeyId(available.length === 1 ? String(available[0].id) : null);
    const initial: Record<string,string> = {};
    for (const client of clients) {
      const candidates = available.filter(key => (client.id !== "claude" || data.groups.some(group => group.id === key.group_id && (group.platform === "anthropic" || group.allow_messages_dispatch))) && (client.id !== "opencode" || data.groups.some(group => group.id === key.group_id && ["openai","anthropic","deepseek","kimi"].includes(group.platform))));
      if (!client.locked && candidates.length === 1) initial[client.id] = String(candidates[0].id);
    }
    setClientKeys(initial); setCatalogs({}); setModels({}); setConfiguring(null);
    if (data.warning) setMessage(data.warning);
  };
  const restoreSession = async () => {
    setBusy(true); setRecovering(true); setLoginError(""); setRestoreFailed(false);
    try {
      if (native || hasRememberedSession()) {
        const data = await request<Session & {saved?:boolean}>(native ? "resume" : "session");
        if (data.user) await acceptSession(data);
        else if (data.warning) setLoginError(data.warning);
      }
    } catch(error) {
      if (error instanceof APIError && error.status === 401) clearRememberedSession();
      else { setRestoreFailed(true); setLoginError(errorText(error)); }
    } finally { setRecovering(false); setBusy(false); }
  };
  useEffect(() => { void restoreSession(); }, []);
  const login = async () => {
    setBusy(true); setLoginError("");
    try {
      const data = await request<Session>(otpRequired ? "login-otp" : "login", otpRequired ? {code:otp} : {email:account.trim(),password});
      setPassword(""); setShowPassword(false);
      if (data.requires_2fa) { setOtpRequired(true); return; }
      await acceptSession(data);
    } catch (error) { setPassword(""); setLoginError(errorText(error)); } finally { setBusy(false); }
  };
  const refreshHistory = async () => setRecords(await request<ConfigRecord[]>("history"));
  const loadClientModels = async (client: string, key: string) => {
    setBusy(true); setConfigurationError(""); setPlans([]);
    try {
      const catalog = await request<Catalog>("models",{key_id:Number(key)});
      setCatalogs(previous => ({...previous,[key]:catalog}));
      setModels(previous => ({...previous,[client]:catalog.models.includes(previous[client]) ? previous[client] : catalog.models.length === 1 ? catalog.models[0] : ""}));
    } catch(error) { setConfigurationError(errorText(error)); } finally { setBusy(false); }
  };
  const openConfiguration = async (client: string) => {
    setConfiguring(client); setRestoring(false); setConfigurationError(""); setPlans([]); setPreview(true);
    if (clientKeys[client] && client !== "opencode") await loadClientModels(client,clientKeys[client]);
  };
  const prepare = async () => {
    if (!configuring || !clientKeys[configuring] || (configuring !== "opencode" && !models[configuring]) || !canConfigure(configuring)) return;
    setBusy(true); setConfigurationError("");
    try { setPlans([await request<Plan>("plan", {client:configuring,model:models[configuring],key_id:Number(clientKeys[configuring]),mode})]); }
    catch(error) { setConfigurationError(errorText(error)); } finally { setBusy(false); }
  };
  const apply = async () => {
    setBusy(true); const results: string[] = [];
    for (const plan of plans) {
      try { const result = await request<ConfigRecord>("apply", {plan_id:plan.id}); results.push(`${clients.find(c => c.id === plan.client)?.name}：${result.state === "unchanged" ? "配置未变化" : restoring ? "已恢复" : "已写入，请重新打开客户端"}`); }
      catch(error) { results.push(`${clients.find(c => c.id === plan.client)?.name}：${errorText(error)}`); }
    }
    setMessage(results.join("；")); setPreview(false); setPlans([]); setBusy(false);
    if (page === "history") { try { await refreshHistory(); } catch(error) { setMessage(errorText(error)); } }
  };
  const temporary = mode === "temporary";
  const go = (next: string) => {
    setPage(next);
    setPassword("");
    setShowPassword(false);
    setMessage("");
    if (next === "history") void refreshHistory().catch(error => setMessage(errorText(error)));
  };
  if (user === null) {
    return (
      <main className="auth-screen">
        <section className="auth-content" aria-labelledby="welcome-title">
          <div className="auth-heading">
            <img src="/brand/susu.png" width={72} height={72} alt="SuSu" />
            <h1 id="welcome-title">欢迎回来</h1>
            {recovering && <p role="status">正在恢复登录…</p>}
          </div>
          <form
            className="auth-form"
            onSubmit={(event) => {
              event.preventDefault();
              void login();
            }}
          >
            <Input
              label="账号"
              name="username"
              type="email"
              placeholder="输入账号邮箱"
              value={account}
              onChange={(event) => setAccount(event.target.value)}
              autoComplete="username"
              required
              disabled={busy || otpRequired}
            />
            {!otpRequired && <InputGroup label="密码">
              <InputGroup.Input
                id="gateway-password"
                name="password"
                type={showPassword ? "text" : "password"}
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="输入密码"
                autoComplete="current-password"
                required
              />
              <InputGroup.Addon align="end">
                <InputGroup.Button
                  type="button"
                  shape="square"
                  icon={showPassword ? EyeSlashIcon : EyeIcon}
                  aria-label={showPassword ? "隐藏密码" : "显示密码"}
                  aria-controls="gateway-password"
                  onClick={() => setShowPassword((visible) => !visible)}
                />
              </InputGroup.Addon>
            </InputGroup>}
            {otpRequired && <Input label="验证码" value={otp} onChange={e => setOtp(e.target.value)} inputMode="numeric" pattern="[0-9]{6}" maxLength={6} autoComplete="one-time-code" required />}
            {loginError && <p className="login-error" role="alert">{loginError}</p>}
            {restoreFailed && <Button type="button" disabled={busy} onClick={() => void restoreSession()}>重试恢复登录</Button>}
            <Button disabled={busy} type="submit" variant="primary" className="login-button">
              {busy ? "登录中…" : otpRequired ? "验证并登录" : "登录"}
            </Button>
          </form>
        </section>
      </main>
    );
  }
  return (
    <Dialog.Root open={preview} onOpenChange={open => { if (!busy) setPreview(open); }}>
      <div className="shell">
        <aside className="sidebar">
          <div className="brand">
            <img
              className="brand-avatar"
              src="/brand/susu.png"
              width={32}
              height={32}
              alt="SuSu"
            />
            <div>
              SurplusToken<small>Connect</small>
            </div>
          </div>
          <nav aria-label="主导航">
            {navigation.map((n) => (
              <Button
                key={n.id}
                variant="ghost"
                className={`nav ${page === n.id ? "active" : ""}`}
                aria-label={n.label}
                aria-current={page === n.id ? "page" : undefined}
                onClick={() => go(n.id)}
              >
                <n.icon size={20} />
                <span className="nav-label">{n.label}</span>
              </Button>
            ))}
          </nav>
          <div className="sidebar-user">
            <Popover>
              <Popover.Trigger
                render={<Button variant="ghost" disabled={busy} className="account-trigger" aria-label="账号菜单" />}
              >
                <span className="user-avatar" aria-hidden="true">
                  {(user.nickname || user.email).slice(0, 2).toUpperCase()}
                </span>
                <span className="account-nickname" title={user.nickname?.trim() || "无昵称用户"}>{user.nickname?.trim() || "无昵称用户"}</span>
              </Popover.Trigger>
              <Popover.Content side="top" align="start" sideOffset={10} className="account-popover">
                <div className="account-summary">
                  <span className="user-avatar" aria-hidden="true">
                    {(user.nickname || user.email).slice(0, 2).toUpperCase()}
                  </span>
                  <div className="account-details">
                    <Popover.Title>{user.nickname?.trim() || "无昵称用户"}</Popover.Title>
                    <Popover.Description className="account-email">{user.email}</Popover.Description>
                  </div>
                </div>
                <div className="account-actions">
                  <Button
                    variant="ghost"
                    icon={SignOutIcon}
                    onClick={async () => {
                      setBusy(true);
                      try { await request("logout"); } catch(error) { setMessage(errorText(error)); setBusy(false); return; }
                      setUser(null); setRestoreFailed(false); setAccount(user.email); setPassword(""); setShowPassword(false); setPreview(false); setPlans([]); setKeys([]); setGroups([]); setCatalogs({}); setClientKeys({}); setKeyId(null); setConfiguring(null); setModels({}); setMessage(""); setBusy(false);
                    }}
                  >
                    退出登录
                  </Button>
                </div>
              </Popover.Content>
            </Popover>
          </div>
        </aside>
        <div className="workspace">
          <header className="topbar">
            <h1>{navigation.find((n) => n.id === page)?.label}</h1>
            {page === "clients" && (
              <Button
                icon={ClockCounterClockwiseIcon}
                disabled={busy}
                onClick={async () => { setBusy(true); try { setDetected(await request<DetectedClient[]>("detect")); } catch(error) { setMessage(errorText(error)); } finally { setBusy(false); } }}
              >
                检测客户端
              </Button>
            )}
          </header>
          <main>
            {fixtureMode && <p className="notice">测试目录模式：只修改测试文件，不影响本机客户端。</p>}
            {page === "clients" && (
              <>
                <section className="section">
                  <div className="mode-row">
                    <span>接入方式</span>
                    <Tabs
                      className="connection-mode"
                      labels={{
                        scrollStart: "向左滚动选项",
                        scrollEnd: "向右滚动选项",
                      }}
                      variant="segmented"
                      value={mode}
                      onValueChange={setMode}
                      tabs={[
                        { value: "temporary", label: "临时启动" },
                        { value: "permanent", label: "永久配置" },
                      ]}
                    />
                  </div>
                  <p className="mode-help">
                    {temporary
                      ? "临时终端启动将在后续接入。"
                      : "写入原生配置，之后直接打开客户端。"}
                  </p>
                </section>
                <section className="clients-section">
                  <div className="section-heading"><h2>客户端</h2></div>
                  <div className="client-list">
                    {clients.map((c) => (
                      <article
                        key={c.id}
                        className="client-row"
                        title={detected.find(d => d.id === c.id)?.reason || (c.id === "claude" && !compatibleKeys(c.id).length ? "没有兼容 Messages API 的 Key" : "")}
                      >
                        <div className="client-choice">
                          <span
                            className={`app-logo ${c.id}`}
                            aria-hidden="true"
                          >
                            {logos[c.id] ? (
                              <img
                                src={`/icons/${logos[c.id]}.svg`}
                                width={22}
                                height={22}
                                alt=""
                              />
                            ) : (
                              <TerminalIcon size={22} />
                            )}
                          </span>
                          <div className="client-name">
                            <span>{c.name}</span>
                            <DetectionBadge client={detected.find(d => d.id === c.id)} name={c.name} protectedClient={!!c.locked} />
                          </div>
                        </div>
                        <Button className="client-configure" aria-label={`${c.name} 配置`} disabled={c.locked || busy || temporary || !canConfigure(c.id)} onClick={() => void openConfiguration(c.id)}>配置</Button>
                      </article>
                    ))}
                  </div>
                </section>
              </>
            )}
            {page === "keys" && (
              <section className="key-management">
                <div className="key-toolbar">
                  <Button aria-label="刷新密钥" icon={ArrowClockwiseIcon} disabled={busy} onClick={async () => {
                    setBusy(true); setMessage("");
                    try { const data = await request<Session>("session"); setKeys(data.keys); setGroups(data.groups); setClientKeys(previous => Object.fromEntries(Object.entries(previous).filter(([,id]) => data.keys.some(key => String(key.id) === id && usableKey(key))))); if (!data.keys.some(k => String(k.id) === keyId)) setKeyId(null); }
                    catch(error) { setMessage(errorText(error)); } finally { setBusy(false); }
                  }} />
                  <Button variant="primary" icon={PlusIcon} disabled={busy} onClick={() => setCreateOpen(!createOpen)}>创建密钥</Button>
                </div>
                <div className="key-filters">
                  <Input aria-label="搜索密钥" placeholder="搜索名称或密钥末尾…" value={keySearch} onChange={event => setKeySearch(event.target.value)} />
                  <Select<string> aria-label="筛选分组" value={keyGroupFilter} items={{all:"全部分组",...Object.fromEntries(groups.map(g => [String(g.id),g.name]))}} onValueChange={setKeyGroupFilter} alignItemWithTrigger={false} />
                  <Select<string> aria-label="筛选状态" value={keyStatusFilter} items={{all:"全部状态",active:"启用",disabled:"禁用",expired:"已过期"}} onValueChange={setKeyStatusFilter} alignItemWithTrigger={false} />
                </div>
                {createOpen && <section className="create-key-row">
                  <Select<string> label="分组" placeholder="选择分组" value={createGroup} items={Object.fromEntries(groups.filter(g => g.status === "active").map(g => [String(g.id),g.name]))} onValueChange={setCreateGroup} alignItemWithTrigger={false} />
                  <Button disabled={busy || !createGroup} onClick={async () => {
                    setBusy(true); setMessage("");
                    try { const key = await request<Key>("create-key", {group_id:Number(createGroup)}); setKeys(previous => [...previous,key]); setCreateOpen(false); await selectKey(String(key.id)); }
                    catch(error) { setMessage(errorText(error)); } finally { setBusy(false); }
                  }}>确认创建</Button>
                </section>}
                <div className="key-table-wrap">
                  <Table aria-label="API 密钥列表">
                    <Table.Header><Table.Row>
                      <Table.Head>名称</Table.Head><Table.Head>API 密钥</Table.Head><Table.Head>分组</Table.Head><Table.Head>过期时间</Table.Head><Table.Head>状态</Table.Head><Table.Head>操作</Table.Head>
                    </Table.Row></Table.Header>
                    <Table.Body>{keys.filter(key => {
                      const expired = !!key.expires_at && new Date(key.expires_at).getTime() <= Date.now();
                      const status = expired ? "expired" : key.status;
                      return (keyGroupFilter === "all" || String(key.group_id) === keyGroupFilter) && (keyStatusFilter === "all" || status === keyStatusFilter) && `${key.name} ${key.mask}`.toLowerCase().includes(keySearch.toLowerCase());
                    }).map(key => {
                      const expired = !!key.expires_at && new Date(key.expires_at).getTime() <= Date.now();
                      const active = key.status === "active" && !expired;
                      const current = String(key.id) === keyId;
                      return <Table.Row key={key.id} variant={current ? "selected" : "default"}>
                        <Table.Cell><span className="key-name" title={key.name}>{key.name}</span></Table.Cell>
                        <Table.Cell><code className="key-mask">{key.mask}</code></Table.Cell>
                        <Table.Cell><span className="tag">{groups.find(g => g.id === key.group_id)?.name ?? "—"}</span></Table.Cell>
                        <Table.Cell><span className="key-expiry" title={key.expires_at ? new Date(key.expires_at).toLocaleString("zh-CN") : undefined}>{key.expires_at ? new Date(key.expires_at).toLocaleDateString("zh-CN") : "永不过期"}</span></Table.Cell>
                        <Table.Cell><span className={`key-status ${active ? "enabled" : ""}`}>{expired ? "已过期" : active ? "启用" : key.status === "disabled" ? "禁用" : key.status}</span></Table.Cell>
                        <Table.Cell><Button disabled={busy || !active} onClick={() => void selectKey(String(key.id))}>{current ? "当前使用" : "使用密钥"}</Button></Table.Cell>
                      </Table.Row>;
                    })}</Table.Body>
                  </Table>
                  {!keys.length && <p className="key-empty">还没有 API 密钥</p>}
                </div>
                <p className="preview-note">共 {keys.length} 个密钥</p>
              </section>
            )}
            {page === "history" && (
              records.length ? <div className="history-list">{records.map(record => <section className="panel history-record" key={record.id}>
                <div><strong>{clients.find(c => c.id === record.client)?.name} · {record.operation === "restore" && record.state === "applied" ? "已恢复" : ({applied:"已写入",pending:"待恢复",failed:"失败",conflict:"有冲突",reverted:"已撤回"} as Record<string,string>)[record.state] ?? record.state}</strong><p>{new Date(record.at).toLocaleString("zh-CN")}</p><p className="config-path">{record.path}</p></div>
                <Button disabled={busy || record.state !== "applied" || record.operation === "restore"} onClick={async () => { setBusy(true); try { const plan = await request<Plan>("restore",{record_id:record.id}); setPlans([plan]); setRestoring(true); setPreview(true); } catch(error) { setMessage(errorText(error)); } finally { setBusy(false); } }}>恢复修改</Button>
              </section>)}</div> : <section className="panel empty-state"><ClockCounterClockwiseIcon size={28} /><h2>还没有配置记录</h2><Button onClick={() => go("clients")}>去配置客户端 <ArrowRightIcon /></Button></section>
            )}
            {page === "settings" && (
              <section className="panel">
                <h2>应用偏好</h2>
                {[
                  ["界面语言", "应用显示语言", "简体中文"],
                  [
                    "配置写入前备份",
                    "永久配置时保留原文件，便于恢复",
                    "默认启用",
                  ],
                  [
                    "Codex 环境保护",
                    "当前阶段禁止启动及修改 Codex CLI / Desktop",
                    "已保护",
                  ],
                  [
                    "本地凭据存储",
                    native ? "登录状态存入系统凭据库；密码不保存" : "开发网页记住本地会话；服务重启后需要登录",
                    "本机存储",
                  ],
                ].map(([title, description, status]) => (
                  <div className="setting-row" key={title}>
                    <div>
                      <strong>{title}</strong>
                      <p>{description}</p>
                    </div>
                    <span className="tag">{status}</span>
                  </div>
                ))}
              </section>
            )}
            {message && (
              <div className="feedback" role="status">
                {message}
                <Button
                  variant="ghost"
                  shape="square"
                  aria-label="关闭提示"
                  onClick={() => setMessage("")}
                >
                  <XIcon />
                </Button>
              </div>
            )}
          </main>
        </div>
      </div>
      <Dialog size="lg" className="plan-dialog" aria-describedby={undefined}>
        <div className="dialog-heading">
          <Dialog.Title>
            {restoring ? "恢复配置修改" : `${clients.find(client => client.id === configuring)?.name || "客户端"} 配置`}
          </Dialog.Title>
          <Dialog.Close
            render={(props) => (
              <Button
                {...props}
                variant="ghost"
                shape="square"
                aria-label="关闭预览"
              >
                <XIcon />
              </Button>
            )}
          />
        </div>
        {!restoring && <div className="preview-models">{targets.map(client => <div key={client}>
          <Select<string> label={`${clients.find(item => item.id === client)?.name} API Key`} value={clientKeys[client] || null} placeholder={compatibleKeys(client).length ? "选择 API Key" : "没有兼容 Key"} disabled={busy} items={Object.fromEntries(compatibleKeys(client).map(key => [String(key.id),`${key.name} · ${key.mask}`]))} onValueChange={key => {
            setClientKeys(previous => ({...previous,[client]:key || ""})); setModels(previous => ({...previous,[client]:""})); setPlans([]);setConfigurationError("");
            if (key && client !== "opencode") void loadClientModels(client,key);
          }} alignItemWithTrigger={false} />
          {!compatibleKeys(client).length && <Button onClick={() => {setPreview(false);go("keys");}}>管理 API Key <ArrowRightIcon /></Button>}
          {client !== "opencode" && <Select<string> label={`${clients.find(item => item.id === client)?.name} 默认模型`} value={models[client] || null} placeholder="选择默认模型" disabled={busy || !clientKeys[client]} items={Object.fromEntries((catalogs[clientKeys[client]]?.models ?? []).map(model => [model,model]))} onValueChange={model => {setModels(previous => ({...previous,[client]:model || ""}));setPlans([]);}} alignItemWithTrigger={false} />}
        </div>)}</div>}
        {configurationError && <p className="login-error" role="alert">{configurationError}</p>}
        {plans.length > 0 && <div className="plan-list">
          {plans.map(plan => <div className="plan-item" key={plan.id}>
            {plan.noop ? <p>配置已一致，无需写入。</p> : <div className="plan-summary">
              <span>{restoring ? "恢复该客户端的配置修改" : plan.client === "opencode" ? plan.template || "OpenCode 供应商模板" : "Claude Code 连接配置"}</span>
              {!restoring && plan.client === "opencode" && <span>{plan.model_count} 个可用模型</span>}
            </div>}
            <Collapsible.Root>
              <Collapsible.DefaultTrigger>查看变更详情</Collapsible.DefaultTrigger>
              <Collapsible.DefaultPanel>
                <p className="config-path">{plan.path}</p>
                <div className="change-list">{plan.changes.map(change => <div key={change.field}><code>{change.field}</code><span>{change.before} → {change.after}</span></div>)}</div>
              </Collapsible.DefaultPanel>
            </Collapsible.Root>
          </div>)}
        </div>}
        <div className="dialog-actions">
          <Dialog.Close
            render={(props) => <Button {...props}>返回调整</Button>}
          />
          <Button variant="primary" disabled={busy || (!restoring && targets.some(client => !clientKeys[client] || (client !== "opencode" && !models[client])))} onClick={() => void (plans.length ? apply() : prepare())}>
            {busy ? "处理中…" : restoring ? "确认恢复" : plans.length ? "应用配置" : "预览配置变更"}
          </Button>
        </div>
      </Dialog>
    </Dialog.Root>
  );
}
function DetectionBadge({client,name,protectedClient}:{client?:DetectedClient;name:string;protectedClient:boolean}) {
 const info = client?.detection;
 return <Popover>
  <Popover.Trigger render={<Button variant="ghost" className={`client-version version-trigger ${protectedClient ? "protected" : ""}`} aria-label={`${name} 安装详情`} />}>
    {client?.version ? `${info?.desktop?.installed ? "CLI " : ""}v${client.version}` : client?.installed ? "版本未知" : "未发现安装"}
    {info?.desktop?.installed && <span> · Desktop {info.desktop.version || "版本未知"}</span>}
    {protectedClient && <span> · 已保护</span>}
    {!protectedClient && !client?.supported && <span> · 配置待适配</span>}
  </Popover.Trigger>
  <Popover.Content side="bottom" align="start" sideOffset={8} className="detection-popover">
    <Popover.Title>{name} 安装详情</Popover.Title>
    <dl>
      <dt>命令行</dt><dd>{client?.installed ? "已发现" : "未在已检查路径中发现"}</dd>
      {info?.executable && <><dt>位置</dt><dd className="config-path">{info.executable}</dd></>}
      <dt>版本</dt><dd>{client?.version || "未知"}{info?.version_status === "metadata" ? "（包信息）" : ""}</dd>
      {info?.detail && <><dt>检测说明</dt><dd>{info.detail}</dd></>}
      {info?.desktop && <><dt>{info.desktop.name || "Desktop"}</dt><dd>{info.desktop.installed ? `${info.desktop.version || "版本未知"} · ${info.desktop.executable}` : info.desktop.detail}</dd></>}
      {!client?.supported && <><dt>配置适配</dt><dd>{client?.reason}</dd></>}
    </dl>
  </Popover.Content>
 </Popover>;
}

async function bootstrap() {
  // The standalone browser preview must not call the native bridge.
  if (location.protocol === "wails:" || location.hostname === "wails.localhost") {
    await import("@wailsio/runtime");
  }
  createRoot(document.getElementById("app")!).render(<App />);
}
void bootstrap().catch((error) => {
  console.error("Connect initialization failed", error);
  document.getElementById("app")!.textContent = "应用初始化失败，请重新打开 Connect。";
});
