# 从 75 行 main.go 入门 Go

> 写给：大三、没学过 Go、但会一点 Java / Python / C 的同学
> 对应代码：`main.go`（75 行）+ `demo/main.go`（19 行）
> 目标：不是“看完一本 Go 书”，而是**把这 75 行每一行为什么这么写讲清楚，并且能删掉重写**。

---

## 0. 先看清：这 75 行到底用了哪些 Go 知识

很多人学 Go 的方式是：打开《Go 语言圣经》从第 1 章看到第 10 章，看到一半放弃。
更有效的方式是：**以这个文件为“考纲”，缺什么补什么**。它虽然只有 75 行，但已经构成一个后端服务的最小闭环：

```
HTTP 请求
  → net/http 起服务、注册路由
  → handler 函数（函数值 / 接口）
  → 读 query 参数（string）
  → http.Client 发请求（指针、超时）
  → io.Reader → []byte（接口、切片）
  → regexp 提取 title（正则、子匹配）
  → struct + tag → encoding/json → 响应
```

按依赖顺序，把它拆成 7 个学习单元：

| 单元 | 本文件中的证据 | 学到什么程度 |
|---|---|---|
| 1. 程序结构 | `package main`、`import (...)`、`go.mod` | 能解释目录、包、模块、`go run` 的关系 |
| 2. 变量与类型 | `var titleRe`、`:=`、`string/int`、`map`、`[]byte`、`struct` | 能默写；理解零值 |
| 3. 函数与错误 | 多返回值、`if err != nil`、`defer`、`_` | 能默写；理解“错误是值” |
| 4. 指针与方法 | `&http.Client{...}`、`client.Get` | 能解释 `&` 和 `.` 为什么能连用 |
| 5. 接口 | `http.ResponseWriter`、`io.Reader`、`any` | **能解释“隐式实现”**，这是 Go 的分水岭 |
| 6. 标准库 | `net/http`、`io`、`regexp`、`encoding/json`、`strings`、`time`、`log` | 会查文档，知道每个函数为什么存在 |
| 7. 并发 | 本文件**没有**（但 net/http 内部已经在并发） | 知道边界，W2 再学 |

> 一句话：**读懂这 75 行 = Go 入门；能不看原文重写 = 真的会了。**

---

## 1. 学习方法：三遍读代码法

### 第一遍：当黑盒跑起来（约 30 分钟，不查语法）

1. 按 `README.md` 装好 Go，`go run .` 跑起来；
2. 浏览器访问 `/health`，再开一个终端 `go run ./demo`，访问 `/scan?url=http://127.0.0.1:9090`；
3. 动手改三个地方，看输出变化：
   - 把 `:8080` 改成 `:8081`；
   - 把 `5 * time.Second` 改成 `1 * time.Second`；
   - 把正则里的 `(?is)` 删掉，再用大写 `<TITLE>` 的页面试一次。

这一遍只建立“输入 → 输出”的直觉。**遇到不懂的符号先跳过**，不要一开始就陷进语法。

### 第二遍：逐行翻译（约 2–3 小时）

把 `main.go` 翻译成你熟悉的语言（Java/Python 伪代码），每个不认识的符号记到一张“符号表”里：

| 符号 | 含义 | 首次出现 |
|---|---|---|
| `package main` | 声明这是可执行包 | 第 1 行 |
| `import (...)` | 分组导入，未使用会编译失败 | 第 3 行 |
| `var` / `:=` | 声明变量 / 短变量声明 | 第 13 / 48 行 |
| `&` | 取地址，得到指针 | 第 48 行 |
| `*` | 指针类型（`*http.Request`）或乘法 | 第 33 / 48 行 |
| `_` | 空白标识符，表示“故意丢弃” | 第 57、74 行 |
| `` `...` `` | 原始字符串，不转义 | 第 14 行 |
| `<<` | 左移，`1<<20` = 1 MiB | 第 57 行 |
| `nil` | 空值（指针/接口/map/slice 等） | 第 29 行 |
| `defer` | 函数返回前才执行 | 第 54 行 |
| `any` | `interface{}` 的别名，任意类型 | 第 71 行 |

翻译完，你应该能回答：**这段代码的输入是什么、输出是什么、哪一步可能失败、失败了返回什么。**

### 第三遍：删掉重写（约半天）

新建一个目录，只给自己需求：

> 写一个 HTTP 服务：`/health` 返回 `{"status":"ok"}`；`/scan?url=...` 请求目标网页，返回状态码和 `<title>`。

不看原文，卡住再瞄一眼。**能写出来，知识才是你的。**

> 读每一行时问自己 5 个问题：
> 1. 这行的输入从哪来？
> 2. 输出到哪去？
> 3. 失败会怎样？
> 4. 为什么不用另一种写法？
> 5. 把这行删掉会怎样？

---

## 2. 先拆掉 7 个旧习惯

如果你学过 Java / Python / C，Go 会让你“别扭”的地方主要在这：

| 你熟悉的 | Go 的做法 | 本文件对应 |
|---|---|---|
| `class` + 继承 | 没有类、没有继承；用 `struct` + 组合 | `type ScanResult struct` |
| `try/catch` | 没有异常；错误是返回值，逐层 `if err != nil` | `resp, err := client.Get` |
| `this` | 方法有“接收者”，写在函数名前面 | `client.Get` 的接收者是 `*http.Client`（本文件没有自定义方法） |
| 对象引用 | 默认按值传递；需要共享/避免拷贝才用指针 | `&http.Client{...}` |
| Maven/pip | `go.mod` 管模块，标准库开箱即用 | `module asset-scanner` |
| 函数重载/默认参数 | 都不支持；用不同函数名或选项结构体 | 本文件没有 |
| 数组 | 更常用 `slice`（`[]byte`、`[]string`） | `body` 是 `[]byte` |

**最重要的一条**：Go 的接口是**隐式实现**的。你没有写 `implements`，只要方法签名对上，就自动满足接口。`http.ResponseWriter`、`io.Reader` 都是这样来的。

---

## 3. 七个学习单元（按顺序学）

### 单元 1：程序结构（第 1–11 行 + go.mod）

```go
package main

import (
    "encoding/json"
    "io"
    ...
)
```

必须理解：

- 每个 `.go` 文件都必须有 `package xxx` 声明（通常在第一行，前面可以有注释）；**同一个目录下的普通 `.go` 文件属于同一个包**。
- `package main` + `func main()` 才是可执行程序；库包不能直接 `go run`。
- `import` 是**包路径**，不是文件名；标准库路径就是 `net/http` 这种。
- **导入未使用 = 编译错误**；局部变量未使用也是编译错误。这是 Go 逼你保持干净。
- `go.mod` 里的 `module asset-scanner` 是模块名；`go 1.21` 是语言版本。
- `demo/` 是另一个目录 = 另一个 `main` 包，所以能 `go run ./demo` 单独跑。

10 分钟实验：

```powershell
go version
go env GOPROXY
go doc net/http.Client          # 查一个类型
go doc net/http.Client.Get      # 查一个方法
go vet ./...                    # 静态检查
gofmt -l .                      # 列出没格式化的文件
```

**验收**：能解释为什么 `main.go` 和 `demo/main.go` 都写 `package main`，却不算同一个包。

---

### 单元 2：变量、类型、零值（第 13、16–22、34、39、57 行）

```go
var titleRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

type ScanResult struct {
    URL        string `json:"url"`
    StatusCode int    `json:"status_code"`
    Title      string `json:"title"`
    Error      string `json:"error,omitempty"`
}
```

必须理解：

1. **声明方式**：`var x T = v`、`x := v`（只能在函数内）、`const`。
   `var titleRe = ...` 省略类型，由右边推断；包级变量在 `main` 之前初始化。
2. **零值**：不赋值也有默认值。`string` 是 `""`，`int` 是 `0`，`bool` 是 `false`，指针/接口/map/slice 是 `nil`。
   所以 `ScanResult{URL: target, Error: err.Error()}` 里没写的字段自动是 `0` 和 `""`。
3. **struct**：字段名大写 = 导出（其他包能访问），小写 = 包内私有。**JSON 只能序列化导出字段。**
4. **struct tag**：反引号里的 `` `json:"url"` `` 不是注释，是给 `encoding/json` 看的元数据。
   - 去掉 tag，JSON 键会变成 `URL`、`StatusCode`；
   - 字段改成小写 `url`，JSON 直接忽略它；
   - `omitempty`：零值时不出现在 JSON 里（所以成功响应没有 `error` 字段）。
5. **map**：`map[string]string{"status": "ok"}` 是字面量；`map` 的零值是 `nil`，**往 nil map 写会 panic**，要先 `make`。
6. **slice**：`[]byte` 是字节切片，不是数组；`string(m[1])` 是 `[]byte` → `string` 的转换（会复制）。
7. **`any`**：第 71 行的 `v any` 等价于 `interface{}`，表示“任意类型”，Go 1.18 起的别名。

10 分钟实验：写一个 15 行的小程序，定义同样的 `ScanResult`，用 `json.Marshal` 打印，分别试：有 tag / 无 tag / 小写字段 / 加 `omitempty`。

**验收**：能解释为什么错误响应里会多出 `"status_code":0`（因为没加 `omitempty`）。这其实是个可以改进的设计点。

---

### 单元 3：函数、错误、defer（第 24、33、38、49–54、71 行）

```go
func scanHandler(w http.ResponseWriter, r *http.Request) {
    resp, err := client.Get(target)
    if err != nil {
        writeJSON(w, http.StatusBadGateway, ScanResult{URL: target, Error: err.Error()})
        return
    }
    defer resp.Body.Close()
    ...
}
```

必须理解：

1. **函数签名**：`func 名字(参数) 返回值`；参数名在前、类型在后；可以有多个返回值。
2. **多返回值**：`resp, err := client.Get(target)` 同时接两个值；`err` 是 `error` 接口，`nil` 表示成功。
3. **错误是值，不是异常**：Go 没有 `try/catch`。惯例是**最后一个返回值是 `error`**，调用方立刻检查。
   `if err != nil { ...; return }` 叫“提前返回”，比层层嵌套清晰。
4. **`defer`**：注册一个“函数返回前执行”的调用，多个 `defer` 后进先出（LIFO）。
   这里用来保证 `resp.Body.Close()` 一定执行，避免连接泄漏。
   **注意**：`log.Fatal` 内部是 `os.Exit`，**不会执行 defer**；`defer` 放在循环里也会堆到函数结束才执行。
5. **`_` 空白标识符**：表示“我知道这里有值，但故意不要”。
   `body, _ := io.ReadAll(...)` 丢弃错误；`_ = json.NewEncoder(w).Encode(v)` 显式表示忽略返回值。
   （Go 允许直接丢弃函数返回值不写 `_`，这里写出来是为了表达“我是故意的”。）
6. **函数是一等公民**：`http.HandleFunc("/health", healthHandler)` 把函数当值传进去；`demo/main.go` 里还用了匿名函数。

10 分钟实验：把 `scanHandler` 里的“发请求 + 读 body + 取 title”抽成一个函数：

```go
func fetchTitle(target string) (statusCode int, title string, err error)
```

（这里用了命名返回值，函数内可以直接 `return`；不熟就写成 `(int, string, error)`。）

要求：`scanHandler` 只负责读参数、调它、写 JSON。这一步能让你真正理解“多返回值”和“错误向上传递”。

---

### 单元 4：指针与方法（第 48–49 行）

```go
client := &http.Client{Timeout: 5 * time.Second}
resp, err := client.Get(target)
```

必须理解：

- `http.Client{...}` 是结构体字面量；`&` 取地址，得到 `*http.Client`。
- Go 有指针，但**没有指针运算**；`p.Field` 不用写 `(*p).Field`，编译器自动解引用。
- `client.Get` 的接收者是 `*http.Client`（指针接收者）；`client` 本身就是指针，直接调用。
  即使 `client` 是值变量，Go 也会自动取地址，但**不要复制 `http.Client`**（它内部有连接池等状态）。
- `time.Second` 是 `time.Duration` 常量；`5 * time.Second` 是 5 秒。`Duration` 底层是 `int64` 纳秒。
- 为什么必须设 `Timeout`：`http.Client` 零值可用，但 `Timeout` 为零表示**没有整体超时**，目标不响应会一直挂着。

**验收**：能回答“把 `&` 去掉，代码还能编译吗？能跑吗？为什么仍然不推荐？”（能编译：`client` 是可寻址变量，Go 会自动取地址；但 `http.Client` 内部有连接池等状态，按惯例用指针、避免复制。）

---

### 单元 5：接口——Go 的分水岭（第 33、57、71–74 行）

这个文件里用到了四个接口，只是你大多没看到 `interface` 关键字：

```go
func healthHandler(w http.ResponseWriter, r *http.Request)
func writeJSON(w http.ResponseWriter, code int, v any)
body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
```

1. **`http.ResponseWriter`** 是接口，大致定义如下：

   ```go
   type ResponseWriter interface {
       Header() Header
       Write([]byte) (int, error)
       WriteHeader(statusCode int)
   }
   ```

   你不需要知道谁实现了它；`net/http` 传进来一个具体类型，你只管调用这三个方法。

2. **`io.Reader`** 是接口：

   ```go
   type Reader interface {
       Read(p []byte) (n int, err error)
   }
   ```

   `resp.Body` 满足它；`io.LimitReader` 返回的东西也满足它；`io.ReadAll` 只要求参数是 `io.Reader`。**接口让函数可以接受“任何能读的东西”。**

3. **`error`** 也是接口：`resp, err := ...` 里的 `err` 就是它，只有一个方法 `Error() string`；任何实现了这个方法的类型都能当错误用。

4. **`any`** 是空接口：所有类型都满足它，所以 `writeJSON` 能接收 `map` 或 `struct`。

5. **隐式实现**：只要一个类型有这些方法，就自动满足接口，不需要声明。这是 Go 和 Java 最大的区别，也是标准库能组合得这么松的原因。

10 分钟实验（非常重要）：自己写一个类型，实现 `http.Handler`：

```go
type versionHandler struct{ version string }

func (h versionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]string{"version": h.version})
}

// 注册：http.Handle("/version", versionHandler{"0.1"})
```

写完你就懂了：`HandleFunc` 只是把普通函数适配成 `Handler` 的语法糖。

---

### 单元 6：标准库四件套（第 25–29、39、44、57、60、72–74 行）

不需要背 API，需要知道“**遇到什么问题，去哪个包找**”，然后会查 `pkg.go.dev` 或 `go doc`。

| 包 | 本文件用它做什么 | 你必须会的 |
|---|---|---|
| `net/http` | 起服务、注册路由、发请求 | `HandleFunc`、`ListenAndServe`、`Client.Get`、`ResponseWriter` 三个方法 |
| `io` | 读响应体、限制大小 | `Reader` 接口、`ReadAll`、`LimitReader` |
| `regexp` | 提取 `<title>` | `MustCompile`、`FindSubmatch`、子匹配下标 |
| `encoding/json` | 返回 JSON | struct tag、`NewEncoder().Encode`、导出字段 |
| `strings` | 判断前缀、去空白 | `HasPrefix`、`TrimSpace` |
| `time` | 超时 | `Duration`、`Second` |
| `log` | 打日志 | `Println`、`Fatal`（= 打印 + `os.Exit(1)`） |

几个容易忽略的细节：

- `r.URL.Query().Get("url")`：`Query()` 返回 `map[string][]string` 的封装，`Get` 取第一个值；参数不存在返回 `""`，不报错。
- `w.Header().Set(...)` 必须在 `w.WriteHeader(code)` **之前**；状态行一旦发出，再改 header 就无效了。
- `json.NewEncoder(w).Encode(v)` 会**多写一个换行**，并且默认转义 `<`、`>`、`&`；想关掉用 `SetEscapeHTML(false)`。
- `http.ListenAndServe(":8080", nil)` 的 `nil` 表示用 `DefaultServeMux`；`":8080"` 监听所有网卡，改成 `"127.0.0.1:8080"` 只监听本机（更安全，也是个小练习）。
- Go 1.22 起 `ServeMux` 支持 `"GET /scan"` 这种新写法；本项目 `go.mod` 写的是 1.21，按老式路由理解即可。

---

### 单元 7：并发——本文件没有，但已经在发生

`main.go` 里没有 `go` 关键字，但 `net/http` 对**每个请求都会开一个 goroutine**。也就是说：

- 两个请求同时进来，`scanHandler` 是并发执行的；
- 现在没有共享可变状态，所以没问题；
- 以后你加“全局计数器 / 缓存 map”，就必须考虑 `sync.Mutex` 或 `channel`，否则会有数据竞争（Linux/macOS 或装了 MinGW 的 Windows 可用 `go run -race .` 检测）。

W2 要学的 `goroutine`、`channel`、`context` 属于下一阶段。现在只需知道：**并发不是从你写 `go` 才开始的。**

### 补丁：本文件没出现，但写 Go 离不开的 5 个东西

- **`for`**：Go 只有 `for`，没有 `while`。三种写法：`for i := 0; i < n; i++ {}`、`for cond {}`、`for {}`（配 `break`）。
- **`range`**：遍历 slice / map / string / channel；`for i, v := range xs`，不要下标就 `for _, v := range xs`。
- **`switch`**：默认不穿透，不需要 `break`；`switch x { case 1: ... default: ... }`。
- **`make` / `append`**：`make(map[string]int)`、`make([]string, 0, 10)`、`xs = append(xs, v)`。本文件的 map/slice 都是字面量或转换，没用到它们。
- **自己定义接口**：本文件只用了标准库接口；下一步试着写 `type Fetcher interface { Fetch(string) (string, error) }`，再让一个 struct 实现它。

这 5 个点花 1 小时在 Tour 上过一遍就够，不需要展开。

---

## 4. 五段精读：把最难的几行拆开

### 片段 A：第 13 行，正则

```go
var titleRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
```

- 反引号是**原始字符串**：`\` 不转义，写正则很舒服。
- `(?is)` 是模式修饰：`i` 忽略大小写，`s` 让 `.` 能匹配换行。
- `<title[^>]*>`：`[^>]*` 吃掉 `<title id="x">` 里的属性。
- `(.*?)`：括号是**捕获组**，`*?` 是非贪婪（匹配到第一个 `</title>` 就停）。
- `MustCompile` 在包初始化时编译；正则写错会 **panic**（不是返回 error）。
- Go 的 `regexp` 是 RE2 引擎，**线性时间、不支持回溯引用**，所以不会像某些语言那样被恶意正则拖死。

### 片段 B：第 17–22 行，struct + tag + JSON

```go
type ScanResult struct {
    URL        string `json:"url"`
    StatusCode int    `json:"status_code"`
    Title      string `json:"title"`
    Error      string `json:"error,omitempty"`
}
```

- 四个字段都是**导出**的（大写），否则 `encoding/json` 看不见。
- tag 只是字符串，写错了**不会报错**，只会悄悄用字段名——这是新手常见的“为什么 JSON 键不对”。
- `omitempty` 对 `string` 是“空串就省略”，对 `int` 是“0 就省略”，对指针是“nil 就省略”。
- 成功响应：`{"url":...,"status_code":200,"title":...}`；失败响应还会多一个 `"status_code":0`（因为没加 `omitempty`）。

### 片段 C：第 24–29 行 + 第 33–35 行，服务与 handler

```go
func main() {
    http.HandleFunc("/health", healthHandler)
    http.HandleFunc("/scan", scanHandler)
    log.Println("服务启动：http://127.0.0.1:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

- `HandleFunc` 把“路径 → 函数”注册到默认路由器；`ListenAndServe` 阻塞运行。
- `log.Fatal` 只在 `ListenAndServe` 返回错误时执行（正常运行时它永远不返回）。
- `healthHandler` 的签名必须精确是 `func(http.ResponseWriter, *http.Request)`，多一个参数都不行。
- `http.HandlerFunc` 是适配器：把普通函数变成实现了 `ServeHTTP` 的类型。`HandleFunc` 内部就做了类似的事。

### 片段 D：第 48–57 行，客户端、错误、defer、io

```go
client := &http.Client{Timeout: 5 * time.Second}
resp, err := client.Get(target)
if err != nil { ...; return }
defer resp.Body.Close()

body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
```

- 先检查 `err`，再 `defer Close`：如果 `resp` 是 `nil`，`resp.Body` 会 panic。
- `LimitReader` 只是“最多读 n 字节”，不是“读满 n 字节”；超过就 EOF。
- `1<<20` 是 1,048,576 字节（1 MiB，约 1 MB）。用移位是为了让“1MB”一眼可见。
- `io.ReadAll` 返回 `[]byte` 和 `error`；这里丢弃 error 是**有意的简化**：网络中断时可能拿到半截 body。
- 真实项目里通常还会检查 `resp.StatusCode`、设置 `User-Agent`、限制重定向次数。

### 片段 E：第 59–68 行 + 第 71–74 行，解析与响应

```go
if m := titleRe.FindSubmatch(body); len(m) > 1 {
    title = strings.TrimSpace(string(m[1]))
}
```

- `if` 可以带一个初始化语句；`m` 的作用域只在 `if/else` 里。
- `FindSubmatch` 返回 `[][]byte`：`m[0]` 是整个匹配，`m[1]` 是第一个捕获组；没匹配到返回 `nil`。
- `len(m) > 1` 是防御：确保捕获组存在。
- `string(m[1])` 把字节切片转成字符串；`TrimSpace` 去掉首尾空白（包括换行）。
- `writeJSON` 里先 `Header().Set`，再 `WriteHeader`，最后 `Encode`；顺序错了 header 就不生效。
- `Encode` 的错误没法再改状态码（响应已经发出），所以显式 `_ =` 忽略。

---

## 5. 七天计划（每天 1–2 小时，可压缩可拉长）

| 天 | 做什么 | 验收标准 |
|---|---|---|
| Day 1 | 装环境、跑通、改端口/超时/正则；学 `go run/build/vet/doc/fmt` | 能独立启动两个服务并解释每个命令 |
| Day 2 | 单元 1 + 2：包、变量、零值、struct、tag、map、slice | 写 20 行小程序：struct → `json.Marshal` → 打印，能解释每个字段 |
| Day 3 | 单元 3 + 4：函数、多返回值、error、defer、指针 | 把 `scanHandler` 拆出 `fetchTitle`，行为不变 |
| Day 4 | 单元 5：接口、`ResponseWriter`、`io.Reader`、`any` | 自己写一个 `http.Handler` 类型并注册成功 |
| Day 5 | 单元 6：`net/http`、`io`、`regexp`、`json` 文档；做 README 三个改造 | `/scan?url=...&timeout=2` 生效；无 title 返回 `""`；`/version` 可用 |
| Day 6 | 不看原文重写 `main.go`；加第一个 `httptest` 测试 | 重写版能跑；`go test ./...` 通过 |
| Day 7 | 复盘 + 自测 10 题；预习 goroutine/channel（只读不写） | 能画出示意图，讲清每个请求为什么已经是并发的 |

> 如果每天只有 30 分钟：把 Day 2–5 各拆成两天，两周完成也完全正常。**不要赶进度，赶“能不能不看代码讲出来”。**

---

## 6. 练习册（从抄到改到重写）

### 热身（每个 10 分钟）

1. 把 `/health` 的返回值改成 `{"status":"ok","service":"asset-scanner"}`。
2. 给 `ScanResult` 加一个 `Server` 字段，从响应头 `Server` 读取并返回。
3. 把监听地址改成 `127.0.0.1:8080`，解释 `:8080` 和 `127.0.0.1:8080` 的区别。

### README 里的三个改造（必做）

**练习 1：`/scan?url=...&timeout=2`**

提示：

```go
timeout := 5
if s := r.URL.Query().Get("timeout"); s != "" {
    if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 30 {
        timeout = n
    }
}
client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
```

要学的点：`strconv.Atoi` 的多返回值、`time.Duration` 转换、参数校验（别让用户传 0 或 10000）。

验收：`timeout=1` 时，请求一个慢站点会在 1 秒左右报 `context deadline exceeded`。

**练习 2：没有 title 的页面返回 `title: ""`**

提示：现在的 `if len(m) > 1` 已经处理了“没匹配到”的情况；真正要验证的是“无 title 页面、空 `<title></title>`、超过 1MB 的页面”三种情况。

注意：README 里让你扫 `http://127.0.0.1:9090/health`，但 demo 用的是 `HandleFunc("/")`，这是 catch-all，`/health` 仍然会返回 `Demo Target`。这正好是坑 #13 的现场例子。

验收：扫自己（返回 JSON、没有 title）：`/scan?url=http://127.0.0.1:8080/health`，`title` 应为空字符串，而不是 500。

**练习 3：加 `/version`**

提示：照抄 `healthHandler` 的结构，注册一行 `http.HandleFunc("/version", versionHandler)`。

验收：返回 `{"version":"0.1"}`，并提交一次语义清晰的 commit。

### 进阶（选做，但强烈建议）

1. **加测试**：用 `httptest.NewRecorder()` 直接调用 `healthHandler`，断言状态码和 body。
2. **拆分函数**：`fetchTitle(target) (int, string, error)`，让 handler 只做参数解析和响应。
3. **自定义 Handler**：用 `type` + `ServeHTTP` 实现 `/version`，体会“函数 → 接口”的适配。
4. **并发预习**：读 README 的 W2，用 `goroutine + channel` 并发解析子域；先画图，再写代码，最后 `go run -race .`。
5. **安全边界**：这个 `/scan` 本质是一个 SSRF 练习；想清楚“如果部署到公网，攻击者能用它访问什么”。这是安全方向的同学必须补的一课。

---

## 7. 调试手册：报错是老师

### 常用命令

```powershell
go run .                 # 编译并运行当前目录
go run ./demo            # 运行 demo 子目录
go build -o app.exe .    # 编译成二进制
go vet ./...             # 静态检查（能发现很多蠢错误）
gofmt -l .               # 列出格式不对的文件；gofmt -w . 直接改
go doc net/http.Client   # 查文档，比搜索引擎快
go test ./...            # 跑测试
go run -race .           # 检测数据竞争
```

### 读编译错误的顺序

1. **只看第一条**：后面的错误很多是第一条的连锁反应；
2. 找到 `文件名:行号`；
3. 常见关键词：
   - `declared and not used`：局部变量声明了没用 → 用掉或删掉；
   - `imported and not used`：导入没用 → 删掉；
   - `undefined: xxx`：拼写错 / 没导入 / 作用域不对；
   - `cannot use x (type A) as type B`：类型不匹配，别硬转，先想为什么；
   - `missing return`：有返回值的方法没写全 return。

### 打印调试

```go
fmt.Printf("%T %#v\n", v, v)   // 看类型和完整结构
log.Printf("target=%q err=%v", target, err)
```

### VS Code

装官方 Go 扩展，`Go: Install/Update Tools` 装 `gopls`、`dlv`；F9 断点、F5 调试。**但先学会看编译错误和 `go doc`，不要一上来就靠补全。**

---

## 8. 这份代码相关的 14 个坑

| # | 坑 | 现象 | 正确做法 |
|---|---|---|---|
| 1 | 导入/局部变量未使用 | 编译失败 | 删掉，或 `_ = x` |
| 2 | `:=` 在 `if` 里遮蔽外层变量 | 外面变量没变 | 用 `=`，或换个名字 |
| 3 | 往 nil map 写 | panic | 先 `make(map[string]string)` |
| 4 | JSON 字段小写 | 输出里没这个字段 | 首字母大写 + tag |
| 5 | tag 拼错 | 不报错，键名不对 | 对照 `json:"..."` 检查 |
| 6 | `defer` 写在循环里 | 函数结束才执行，资源堆积 | 包成匿名函数，或手动 Close |
| 7 | `log.Fatal` 前有 `defer` | defer 不执行 | 用 `return err` 向上传递 |
| 8 | `WriteHeader` 后再 `Set` header | header 不生效 | 先 Header，再 WriteHeader |
| 9 | 正则 `.*` 贪婪 | 匹配到最后一个 `</title>` | 用 `.*?` |
| 10 | `[]byte` 和 `string` 直接 `==` | 编译错误 | `string(b) == s` 或 `bytes.Equal` |
| 11 | 忽略 `Get` 的 err 后用 `resp` | nil panic | 先 `if err != nil` |
| 12 | 不 Close body / 不设 Timeout | 连接泄漏、永久挂起 | `defer Close` + `Timeout` |
| 13 | `HandleFunc("/")` | 匹配所有未注册路径 | 用更具体的 pattern |
| 14 | 以为 handler 是串行的 | 共享变量出现随机错误 | 加锁 / 用 channel / `-race` |

---

## 9. 资源清单（按顺序用，别收藏吃灰）

**必做（约 4–6 小时）**

1. [A Tour of Go 中文版](https://tour.go-zh.org/)：只做 Basics、Methods and interfaces、Concurrency 三部分；每道练习都敲一遍。
2. [Go by Example 中文版](https://gobyexample-cn.github.io/)：对照读 `values`、`variables`、`slices`、`maps`、`structs`、`methods`、`interfaces`、`errors`、`goroutines`、`channels`、`http-client`、`http-server`、`json`。
3. [pkg.go.dev](https://pkg.go.dev/)：本文件用到的 `net/http`、`io`、`regexp`、`encoding/json` 各点进去看 10 分钟，重点看函数签名和示例。

**当字典用（不要通读）**

4. [Effective Go](https://go.dev/doc/effective_go)：讲“为什么这样写才地道”，遇到问题查对应小节。
5. [《Go 语言圣经》中文版](https://gopl-zh.github.io/)：第 1–7 章当参考书；不要从第一页硬啃。
6. [Go FAQ](https://go.dev/doc/faq)：解释“为什么没有异常/继承/重载”这类困惑。

**进阶（入门后再看）**

7. [100 Go Mistakes](https://100go.co/)：工作后用得上，现在只看前几章即可。
8. [Go 官方博客](https://go.dev/blog/)：`Errors are values`、`Slices: usage and internals` 两篇必读。

**视频**：可以看，但**不要只看**。每看 20 分钟，必须自己敲 10 行。

---

## 10. 自测 10 题（能答 8 题才算过关）

1. `package main` 和 `func main()` 各自的作用是什么？为什么 `demo/` 也能 `go run`？
2. 为什么“导入未使用”是编译错误，而包级变量未使用却没事？
3. 去掉 `json:"url"` 这个 tag 会怎样？把 `URL` 改成 `url` 又会怎样？
4. `:=` 和 `=` 的区别？`_` 是干什么的？
5. `&http.Client{...}` 里的 `&` 是什么？去掉还能编译吗？
6. `defer resp.Body.Close()` 什么时候执行？为什么放在 `if err != nil` 之后？
7. `io.LimitReader` 和 `io.ReadAll` 各做什么？`1<<20` 是多少？
8. `FindSubmatch` 返回什么？`m[0]` 和 `m[1]` 有什么区别？没匹配到返回什么？
9. `w.Header().Set` 和 `w.WriteHeader` 的顺序为什么重要？
10. `json.NewEncoder(w).Encode(v)` 前面的 `_ =` 是什么意思？不写行不行？

### 参考答案（先自己答，再对照）

1. `package main` 声明可执行包；`func main` 是入口。`demo/` 是另一个目录 = 另一个 `main` 包，单独编译。
2. 未使用的导入会拖慢编译、膨胀依赖，所以强制报错；包级变量可能被其他文件使用，编译器不轻易判定。
3. 去 tag：JSON 键变成 `URL`；字段小写：未导出，JSON 忽略，输出缺字段。
4. `:=` 声明并赋值（至少一个新变量）；`=` 只赋值；`_` 丢弃不需要的返回值。
5. `&` 取地址，得到 `*http.Client`；去掉能编译（变量可寻址，自动取地址），但复制 Client 不推荐。
6. 函数返回时执行；放在 err 检查后，避免 `resp` 为 nil 时 `resp.Body` panic。
7. `LimitReader` 限制最多读 n 字节；`ReadAll` 读到 EOF 返回 `[]byte`；`1<<20` = 1 MiB。
8. 返回 `[][]byte`；`m[0]` 整个匹配，`m[1]` 第一个捕获组；没匹配返回 `nil`。
9. `WriteHeader` 一旦调用，状态行就发出，之后改 header 无效；必须先 Header 后 WriteHeader。
10. 显式表示“故意忽略 Encode 的错误”；不写也能编译，但意图不明确。

---

## 11. 今天就做的 3 件事

1. `go run .` + `go run ./demo`，把 README 的 `/scan` 示例跑通；
2. 把 `main.go` 从头到尾抄一遍（**手打，不复制**），每行写一句中文注释；
3. 在 `go-learning-guide.md` 旁边新建 `notes.md`，记录今天遇到的 3 个报错和它们的含义。

> 学 Go 最快的路径不是“看完”，而是“跑通 → 改坏 → 修好 → 重写”。
> 这 75 行就是你的第一块试验田。
