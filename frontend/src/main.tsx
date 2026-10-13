import { useState } from "react";
import { createRoot } from "react-dom/client";
import { Button } from "@cloudflare/kumo/components/button";
import { Select } from "@cloudflare/kumo/components/select";
import { Tabs } from "@cloudflare/kumo/components/tabs";
import { Checkbox } from "@cloudflare/kumo/components/checkbox";
import { Popover } from "@cloudflare/kumo/components/popover";
import { Dialog } from "@cloudflare/kumo/components/dialog";
import { Input } from "@cloudflare/kumo/components/input";
import { InputGroup } from "@cloudflare/kumo/components/input-group";
import {
  SquaresFourIcon,
  SignOutIcon,
  ClockCounterClockwiseIcon,
  GearIcon,
  ArrowRightIcon,
  LockKeyIcon,
  TerminalIcon,
  XIcon,
  EyeIcon,
  EyeSlashIcon,
} from "@phosphor-icons/react";
import "@cloudflare/kumo/styles/standalone";
import "./style.css";
import { clients } from "./clients";

const navigation = [
  { id: "clients", label: "客户端", icon: SquaresFourIcon },
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
  const [user, setUser] = useState<{ email: string; nickname?: string } | null>(null);
  const [account, setAccount] = useState("");
  const [page, setPage] = useState("clients");
  const [mode, setMode] = useState("temporary");
  const [selected, setSelected] = useState(["claude"]);
  const [models, setModels] = useState<Record<string, string>>({});
  const [preview, setPreview] = useState(false);
  const [message, setMessage] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const temporary = mode === "temporary";
  const go = (next: string) => {
    setPage(next);
    setPassword("");
    setShowPassword(false);
    setMessage("");
  };
  if (user === null) {
    return (
      <main className="auth-screen">
        <section className="auth-content" aria-labelledby="welcome-title">
          <div className="auth-heading">
            <img src="/brand/susu.png" width={72} height={72} alt="SuSu" />
            <h1 id="welcome-title">欢迎回来</h1>
          </div>
          <form
            className="auth-form"
            onSubmit={(event) => {
              event.preventDefault();
              // UI flow preview only; gateway authentication is not connected yet.
              setUser({ email: account.trim() });
              setPassword("");
              setShowPassword(false);
              go("clients");
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
            />
            <InputGroup label="密码">
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
            </InputGroup>
            <Button type="submit" variant="primary" className="login-button">
              登录
            </Button>
          </form>
        </section>
      </main>
    );
  }
  return (
    <Dialog.Root open={preview} onOpenChange={setPreview}>
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
                render={<Button variant="ghost" className="account-trigger" aria-label="账号菜单" />}
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
                    onClick={() => {
                      setUser(null);
                      setAccount("");
                      setPassword("");
                      setShowPassword(false);
                      setPreview(false);
                      setMessage("");
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
                onClick={() =>
                  setMessage("客户端检测尚未接入，未执行任何程序。")
                }
              >
                检测客户端
              </Button>
            )}
          </header>
          <main>
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
                      ? "通过 Connect 启动，使用独立配置目录。"
                      : "写入原生配置，之后直接打开客户端。"}
                  </p>
                </section>
                <section className="clients-section">
                  <div className="section-heading">
                    <h2>
                      客户端 <span>已选 {selected.length}</span>
                    </h2>
                    <Button
                      variant="ghost"
                      onClick={() =>
                        setSelected(
                          selected.length === 4
                            ? []
                            : clients.filter((c) => !c.locked).map((c) => c.id),
                        )
                      }
                    >
                      {selected.length === 4 ? "取消全选" : "全选"}
                    </Button>
                  </div>
                  <div className="client-list">
                    {clients.map((c) => (
                      <article
                        key={c.id}
                        className={`client-row ${selected.includes(c.id) ? "selected" : ""}`}
                      >
                        <div className="client-choice">
                          <Checkbox
                            aria-label={`选择 ${c.name}`}
                            checked={selected.includes(c.id)}
                            disabled={c.locked}
                            onCheckedChange={(checked) =>
                              setSelected((previous) =>
                                checked
                                  ? [...previous, c.id]
                                  : previous.filter((id) => id !== c.id),
                              )
                            }
                          />
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
                          <span>{c.name}</span>
                        </div>
                        <span className={`tag ${c.locked ? "protected" : ""}`}>
                          {c.locked ? "已保护" : "未检测"}
                        </span>
                        {c.locked ? (
                          <span className="protected-copy">
                            <span className="inline-icon">
                              <LockKeyIcon />
                            </span>
                            <span>不启动或修改 Codex</span>
                          </span>
                        ) : (
                          <Select<string>
                            aria-label={`${c.name} 模型`}
                            className="model-select"
                            placeholder="选择模型"
                            value={models[c.id] ?? null}
                            onValueChange={(v) =>
                              setModels((prev) => ({
                                ...prev,
                                [c.id]: v ?? "",
                              }))
                            }
                            items={Object.fromEntries(
                              c.models.slice(1).map((m) => [m, m]),
                            )}
                            alignItemWithTrigger={false}
                            sideOffset={6}
                          />
                        )}
                      </article>
                    ))}
                  </div>
                </section>
                <div className="action-bar">
                  <span>
                    {temporary ? "保留原有全局配置" : "写入前自动备份"}
                  </span>
                  <Dialog.Trigger
                    render={(props) => (
                      <Button
                        {...props}
                        id="preview"
                        variant="primary"
                        disabled={!selected.length}
                      >
                        预览{temporary ? "启动方案" : "配置变更"}{" "}
                        <ArrowRightIcon />
                      </Button>
                    )}
                  />
                </div>
                <p className="preview-note">
                  界面预览：模型为示例，不执行启动或写入。
                </p>
              </>
            )}
            {page === "history" && (
              <section className="panel empty-state">
                <ClockCounterClockwiseIcon size={28} />
                <div className="empty-copy">
                  <h2>还没有配置记录</h2>
                  <p>首次应用永久配置后，原文件备份与变更摘要会显示在这里。</p>
                </div>
                <Button onClick={() => go("clients")}>
                  去配置客户端 <ArrowRightIcon />
                </Button>
              </section>
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
                    "后续接入系统钥匙串，当前未存储任何凭据",
                    "待接入",
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
      <Dialog size="lg" className="plan-dialog">
        <div className="dialog-heading">
          <Dialog.Title>
            {temporary ? "临时启动方案" : "永久配置变更"}
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
        <Dialog.Description className="dialog-description">
          {temporary
            ? "使用专属配置目录；通过 Connect 启动所选工具。"
            : "写入客户端原生配置；写入前保存备份与原文件校验值。"}
        </Dialog.Description>
        <div className="plan-list">
          {clients
            .filter((c) => selected.includes(c.id))
            .map((c) => (
              <div className="plan-item" key={c.id}>
                <strong>{c.name}</strong>
                <p>{models[c.id] || "尚未选择模型"}</p>
                <code>
                  {temporary ? `${c.env} → <Connect 数据目录>/${c.id}` : c.path}
                </code>
                <p>
                  {temporary
                    ? c.id === "opencode"
                      ? "配置层与凭据层需分别隔离"
                      : c.id === "omp"
                        ? "需同时处理 profile 与跨工具配置发现"
                        : "仅为新启动的子进程设置环境变量"
                    : "备份 → 最小字段合并 → 校验 → 原子替换"}
                </p>
              </div>
            ))}
        </div>
        <p className="notice">
          这是方案预览。模型获取及执行功能尚未接入。
        </p>
        <div className="dialog-actions">
          <Dialog.Close
            render={(props) => <Button {...props}>返回调整</Button>}
          />
          <Button
            variant="primary"
            onClick={() => {
              setPreview(false);

            }}
          >
            完成预览
          </Button>
        </div>
      </Dialog>
    </Dialog.Root>
  );
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
