# asset-scanner（起步骨架）

一个 70 行左右的 Go HTTP 服务，两个接口：

- `GET /health` → `{"status":"ok"}`
- `GET /scan?url=http://...` → `{"url":"...","status_code":200,"title":"..."}`

它不是一个"项目"，是你项目的**第一个能跑起来的版本**：先证明"我能让一个服务在本地跑起来"，再往上加资产发现、POC、并发。

---

## 0. 装环境（Windows）

```powershell
# 1) 安装 Go（或者去 https://go.dev/dl/ 下载 msi 双击安装）
winget install GoLang.Go

# 2) 重开一个 PowerShell 窗口，确认安装成功
go version

# 3) 国内下载依赖会慢，配一下代理
go env -w GOPROXY=https://goproxy.cn,direct
```

再装两样东西：

- **VS Code** + 扩展 **Go**（微软官方那个）；
- **Git**，然后配置身份：

```powershell
git config --global user.name "你的名字"
git config --global user.email "你的邮箱"
```

---

## 1. 跑起来

在 `asset-scanner` 目录下：

```powershell
go run .
```

第一次会编译几秒钟。看到日志 `服务启动：http://127.0.0.1:8080` 后，浏览器打开：

```
http://127.0.0.1:8080/health
```

看到 `{"status":"ok"}` 就算成功。

> 如果提示 8080 被占用，把 `main.go` 里的 `:8080` 改成 `:8081`，重新 `go run .`。

---

## 2. 测 `/scan`

`/scan` 需要请求一个真实网页。不要扫公网，先用本地的测试站：

新开一个终端，在 `asset-scanner` 目录下：

```powershell
go run ./demo
```

它会起一个 `http://127.0.0.1:9090` 的页面，标题是 `Demo Target`。

然后访问：

```
http://127.0.0.1:8080/scan?url=http://127.0.0.1:9090
```

应该返回：

```json
{"url":"http://127.0.0.1:9090","status_code":200,"title":"Demo Target"}
```

命令行也可以：

```powershell
curl "http://127.0.0.1:8080/scan?url=http://127.0.0.1:9090"
```

---

## 3. 推到 GitHub

先在 GitHub 上建一个**空仓库**（不要勾选 README / .gitignore），名字叫 `asset-scanner`，然后：

```powershell
git init
git add .
git commit -m "init: health + scan"
git branch -M main
git remote add origin https://github.com/你的用户名/asset-scanner.git
git push -u origin main
```

如果命令行认证搞不定，装 **GitHub Desktop**：Add local repository → 选中 `asset-scanner` → Publish repository，更省事。

---

## 4. 三个改造练习（做完再往下走）

1. **加 timeout 参数**：`/scan?url=...&timeout=2`，用 `strconv.Atoi` 解析，替换掉写死的 5 秒。
2. **处理没有 title 的页面**：访问 `/scan?url=http://127.0.0.1:9090/health`（demo 没有这个路由，会 404、没有 title），要求返回 `title: ""` 而不是报错。
3. **加一个 `/version` 接口**：返回 `{"version":"0.1"}`，提交一次 commit，message 写清楚改了什么。

每做完一个，`git add . && git commit -m "..."`，再 `git push`。

---

## 5. 这份代码是怎么拆出来的

| 需求 | 子任务 | 怎么验证 |
|---|---|---|
| 起一个服务 | `http.ListenAndServe` + `/health` | 浏览器看到 `{"status":"ok"}` |
| 读 url 参数 | `r.URL.Query().Get("url")` | 访问 `/scan?url=abc` 能看到 abc |
| 请求目标 | `http.Client.Get` | 能拿到状态码 200 |
| 解析标题 | 正则 `FindSubmatch` | `title` 是 `Demo Target` |
| 返回 JSON | `encoding/json` | 浏览器看到结构化结果 |
| 推仓库 | `git init/add/commit/push` | GitHub 页面上有代码 |

**任何任务都这样拆：先写清楚"输入→处理→输出"，再拆成 5 个左右能单独跑出结果的小步。** 每一步都在 10 行以内，卡住只可能卡在一个点。

---

## 6. 常见报错

| 报错 | 原因 | 怎么办 |
|---|---|---|
| `'go' 不是内部或外部命令` | 没装 / 没重开终端 / PATH 没生效 | 重开 PowerShell，再 `go version` |
| `bind: address already in use` | 8080 被占用 | 换端口，或关掉占用进程 |
| `context deadline exceeded` | 目标 5 秒没响应 | 换成本地 demo 站 |
| `x509: certificate signed by unknown authority` | 目标用了自签证书 | v0 不支持，记下来当面试话题 |
| `go: module ... not found` | 不在 `asset-scanner` 目录下执行 | `cd asset-scanner` 再跑 |

---

## 7. 下一步（W2）

把 `/scan` 从"扫一个 URL"扩展成"输入域名 → 输出子域列表"：

1. 准备一个字典文件 `subdomains.txt`（几十行就够，比如 www、mail、api、admin、test）；
2. 逐行拼成 `xxx.example.com`，用 `net.LookupHost` 解析；
3. 用 goroutine + channel 并发解析，加超时；
4. 结果去重，输出 JSON。

先想清楚"输入是什么、输出长什么样"，再写代码。不要一上来就写并发。
