# 📊 mog — Microsoft Ops Gadget

> **Microsoft 365 命令行工具** — 支持邮件、日历、云盘、联系人、任务、Word、PowerPoint、Excel、OneNote

[![Go Reference](https://pkg.go.dev/badge/github.com/iskycc/mogcli.svg)](https://pkg.go.dev/github.com/iskycc/mogcli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`mog` 是 [gog](https://github.com/visionik/gog)（Google Ops Gadget）的微软版。相同的操作习惯，不同的云服务。

---

## 📝 关于本仓库

**本仓库是 [visionik/mogcli](https://github.com/visionik/mogcli) 的一个分支。**

We would like to express our sincere gratitude to the original authors — **[visionik](mailto:visionik@pobox.com)** and **Vinston** — for creating the excellent `mogcli` project and open-sourcing it under the MIT license.

**本分支的主要改动：**
- ✅ **新增中国区（世纪互联 / 21Vianet）Office 365 支持**：通过 `--region china` 自动切换 Graph API 和 OAuth2 认证端点
- ✅ **模块路径迁移**：从 `github.com/visionik/mogcli` 迁移至 `github.com/iskycc/mogcli`
- ✅ 保留原有的多账户、Slug 系统、别名系统等全部功能

**开源许可证**：本项目继续沿用原项目的 **MIT License**。

---

## ✨ 功能模块

| 模块 | 说明 |
|------|------|
| 📧 **邮件 (Mail)** | 搜索、发送、草稿、附件、文件夹 |
| 📅 **日历 (Calendar)** | 事件、创建、响应、忙闲查询、ACL |
| 📁 **云盘 (Drive)** | OneDrive 文件 — 列出、上传、下载、移动 |
| 👥 **联系人 (Contacts)** | 个人联系人 + 组织目录搜索 |
| ✅ **任务 (Tasks)** | Microsoft To-Do — 列表、添加、完成、清空 |
| 📝 **Word** | 文档 — 列出、导出、复制 |
| 📊 **PowerPoint** | 演示文稿 — 列出、导出、复制 |
| 📈 **Excel** | 电子表格 — 读取、写入、表格、导出 |
| 📓 **OneNote** | 笔记本、分区、页面、搜索 |

**特色功能：**
- 🔗 **Slug 系统** — 用 8 位短码替代微软超长的 GUID
- 🤖 **AI 友好** — `--ai-help` 输出完整的命令文档，方便 LLM 调用
- 🔄 **gog 兼容** — 与 `gog` 相同的命令结构和参数习惯

---

## 🚀 快速开始

```bash
# 安装
go install github.com/iskycc/mogcli/cmd/mog@latest

# 认证（国际版 / 默认）
mog auth login --client-id YOUR_CLIENT_ID

# 认证（中国区 / 世纪互联）
mog auth login --client-id YOUR_CLIENT_ID --region china

# 查看邮件
mog mail search "*" --max 10

# 发送邮件
mog mail send --to bob@example.com --subject "Hello" --body "Hi Bob!"

# 列出日历事件
mog calendar list

# 创建会议
mog calendar create --summary "会议" \
  --from 2025-01-15T10:00:00 --to 2025-01-15T11:00:00 \
  --attendees "alice@example.com"

# 上传文件到 OneDrive
mog drive upload ./report.pdf

# 添加任务
mog tasks add "Review PR" --due tomorrow --important

# 读取 Excel
mog excel get myworkbook.xlsx Sheet1 A1:D10

# 搜索 OneNote
mog onenote search "meeting notes"
```

---

## 📦 安装

```bash
# 通过 Go 安装（推荐）
go install github.com/iskycc/mogcli/cmd/mog@latest

# 或克隆后本地编译
git clone https://github.com/iskycc/mogcli.git
cd mogcli
go build -o mog ./cmd/mog
```

---

## ⚙️ 配置 Azure AD 应用

### 1. 注册应用

1. 访问对应门户：
   - **国际版**：[Azure Portal](https://portal.azure.com)
   - **中国区（世纪互联）**：[Azure China Portal](https://portal.azure.cn)
2. 进入 **App registrations** → **New registration**
3. **Name**：`mog CLI`（任意名称）
4. **Supported account types**：
   - 国际版选择 **"Accounts in any organizational directory (Any Azure AD directory - Multitenant)"**
   - 中国区（世纪互联）通常为单租户环境，选择 **"Accounts in this organizational directory only (Single tenant)"** 即可
5. **Redirect URI**：留空（使用 Device Code Flow）

### 2. 启用公共客户端流

1. 在应用注册页中，进入 **Authentication** → **Advanced settings**
2. 将 **"Allow public client flows"** 设为 **Yes**
3. 点击 **Save**

> ⚠️ Device Code Flow 必须开启 **"Allow public client flows"**，否则无法登录。

### 3. 添加 API 权限

添加以下 **Delegated** 权限：

| 权限 | 说明 |
|------|------|
| `User.Read` | 登录并读取用户配置文件 |
| `offline_access` | 维持访问（获取 Refresh Token） |
| `Mail.ReadWrite` | 读写邮件 |
| `Mail.Send` | 发送邮件 |
| `Calendars.ReadWrite` | 完整日历权限 |
| `Files.ReadWrite.All` | 完整 OneDrive 权限 |
| `Contacts.Read` | 读取联系人 |
| `Contacts.ReadWrite` | 完整联系人权限 |
| `People.Read` | 读取组织人员 |
| `Tasks.ReadWrite` | 读写任务 |
| `Notes.ReadWrite` | 读写 OneNote |

> **中国区提示**：世纪互联版 Graph API 对部分权限的支持可能有限（如 `People.Read`、`Notes.ReadWrite`）。若授权时遇到权限错误，可先在 Azure Portal 中移除对应权限后重试。

### 4. 登录

```bash
# 国际版（默认）
mog auth login --client-id YOUR_CLIENT_ID

# 中国区（世纪互联 / 21Vianet）
mog auth login --client-id YOUR_CLIENT_ID --region china
```

- **Client ID 不可混用**：在 `portal.azure.cn` 注册的应用只能用于 `--region china`；在 `portal.azure.com` 注册的应用只能用于默认的全球版。
- 登录时会自动打开浏览器进行微软认证。Token 保存在 `~/.config/mog/tokens.json`。

### 5. 验证

```bash
mog auth status
```

---

## 👥 多账户支持

支持同时管理多个 Microsoft 365 账户（个人、工作、国际版、中国区等）：

### 配置多个账户

```bash
# 登录并命名账户
mog auth login --client-id GLOBAL_CLIENT_ID --account global --region global
mog auth login --client-id CHINA_CLIENT_ID --account china --region china
mog auth login --client-id YOUR_CLIENT_ID --account work
mog auth login --client-id YOUR_CLIENT_ID --account personal
```

### 切换账户使用

```bash
# 单条命令临时切换
mog mail search "*" --account china
mog calendar list --account global

# 或通过环境变量（当前会话生效）
export MOG_ACCOUNT=china
mog mail search "*"
```

### 列出已配置账户

```bash
mog auth list
```

### 账户存储结构

每个账户的配置完全隔离：

```
~/.config/mog/
  default/
    settings.json    # 配置（含 region）
    tokens.json      # OAuth Token
    slugs.json       # Slug 缓存
    aliases.json     # 别名
  china/
    settings.json
    tokens.json
    slugs.json
    aliases.json
  work/
    ...
```

旧版单账户配置会自动迁移到 `default` 目录下。

---

## 🌏 中国区（世纪互联）说明

由 21Vianet 运营的 Office 365 中国版与国际版在认证端点和 Graph API 域名上不同：

| 环境 | Graph API | OAuth2 认证 |
|------|-----------|-------------|
| 全球版（默认） | `graph.microsoft.com` | `login.microsoftonline.com` |
| 中国区 | `microsoftgraph.chinacloudapi.cn` | `login.partner.microsoftonline.cn` |

使用 `--region china` 登录后，`region` 会按账户保存在 `settings.json` 中，后续所有命令（邮件、日历、云盘等）都会自动走中国区端点，无需每次手动指定。

### 中国区已知 API 限制

以下命令在世纪互联环境中**可能不可用或功能受限**：

| 模块 | 命令 | 说明 |
|------|------|------|
| **OneNote** | `mog onenote ...` | `Notes.ReadWrite` 及 OneNote Graph API 在 21Vianet 环境中**常不可用**，可能返回 `NotImplemented` 或权限错误。 |
| **联系人** | `mog contacts directory ...` | `People.Read`（组织通讯录搜索）支持有限，可能返回空结果或报错。 |
| **Excel** | 复杂表格操作 | 部分高级 Workbook/Table API 行为可能不一致。基础读写通常正常。 |

**通常可正常使用的模块**：邮件、日历、云盘、任务（To-Do）、Word、PowerPoint、基础 Excel。

---

## 📖 命令参考

### 全局选项

| 选项 | 说明 |
|------|------|
| `--account`, `-a` | 账户名称（默认：`default`，环境变量：`MOG_ACCOUNT`） |
| `--json` | 输出 JSON（适合脚本） |
| `--plain` | 输出纯文本（TSV，无颜色） |
| `--verbose`, `-v` | 显示完整 ID 和额外信息 |
| `--force` | 跳过确认提示 |
| `--no-input` | 禁止交互式输入（CI 模式） |
| `--ai-help` | 输出完整的 AI/LLM 参考文档 |

---

### 📧 邮件 (Mail)

```bash
mog mail search <query>              # 搜索邮件
mog mail search "*" --max 10         # 最近 10 封邮件
mog mail get <id>                    # 读取指定邮件

mog mail send --to X --subject Y --body Z
mog mail folders                     # 列出文件夹

# 草稿
mog mail drafts list
mog mail drafts create --to X --subject Y --body Z
mog mail drafts send <draftId>

# 附件
mog mail attachment list <messageId>
mog mail attachment download <messageId> <attachmentId> --out ./file.pdf
```

---

### 📅 日历 (Calendar)

```bash
mog calendar list                    # 即将发生的事件
mog calendar list --from 2025-01-01 --to 2025-01-31
mog calendar calendars               # 列出日历列表

mog calendar create --summary "会议" \
  --from 2025-01-15T10:00:00 \
  --to 2025-01-15T11:00:00

mog calendar get <eventId>
mog calendar update <eventId> --summary "新标题"
mog calendar delete <eventId>

# 响应邀请
mog calendar respond <eventId> accept
mog calendar respond <eventId> decline --comment "无法参加"

# 查看忙闲
mog calendar freebusy alice@example.com bob@example.com \
  --start 2025-01-15T09:00:00 --end 2025-01-15T17:00:00

# 查看权限
mog calendar acl
```

别名：`mog cal` → `mog calendar`

---

### 📁 云盘 (Drive / OneDrive)

```bash
mog drive ls                         # 根目录
mog drive ls /Documents              # 指定路径
mog drive search "report"            # 搜索文件

mog drive download <id> --out ./file.pdf
mog drive upload ./doc.pdf
mog drive upload ./doc.pdf --folder <folderId> --name "renamed.pdf"

mog drive mkdir "新文件夹"
mog drive move <id> <destinationId>
mog drive rename <id> "new-name.pdf"
mog drive copy <id> --name "copy.pdf"
mog drive rm <id>                    # 删除文件
```

---

### ✅ 任务 (Tasks / Microsoft To-Do)

```bash
mog tasks lists                      # 列出任务列表
mog tasks list                       # 默认列表中的任务
mog tasks list <listId>              # 指定列表中的任务
mog tasks list --all                 # 包含已完成任务

mog tasks add "买牛奶"
mog tasks add "给妈妈打电话" --due tomorrow --notes "生日"
mog tasks add "Review PR" --list Work --due monday --important

mog tasks done <taskId>
mog tasks undo <taskId>
mog tasks delete <taskId>
mog tasks clear                      # 清空已完成任务
mog tasks clear <listId>             # 清空指定列表的已完成任务
```

别名：`mog todo` → `mog tasks`

---

### 👥 联系人 (Contacts)

```bash
mog contacts list
mog contacts search "john"
mog contacts get <id>

mog contacts create --name "John Doe" --email "john@example.com"
mog contacts update <id> --email "new@example.com"
mog contacts delete <id>

mog contacts directory "john"        # 组织目录搜索
```

---

### 📈 Excel

```bash
mog excel list                       # 列出工作簿
mog excel metadata <id>              # 列出工作表

# 读取数据
mog excel get <id>                   # 第一个工作表，已用范围
mog excel get <id> Sheet1 A1:D10     # 指定范围

# 写入数据（按行填充）
mog excel update <id> Sheet1 A1:B2 val1 val2 val3 val4

# 追加到表格
mog excel append <id> TableName col1 col2 col3

# 创建与管理
mog excel create "预算 2025"
mog excel add-sheet <id> --name "Q2"
mog excel tables <id>
mog excel clear <id> Sheet1 A1:C10   # 清空数值（保留格式）
mog excel copy <id> "预算副本"

# 导出
mog excel export <id> --out ./data.xlsx
mog excel export <id> --format csv --out ./data.csv
```

---

### 📓 OneNote

```bash
mog onenote notebooks                # 列出笔记本
mog onenote sections <notebookId>    # 列出分区
mog onenote pages <sectionId>        # 列出页面
mog onenote get <pageId>             # 获取页面内容（文本）
mog onenote get <pageId> --html      # 获取原始 HTML

mog onenote create-notebook "Work Notes"
mog onenote create-section <notebookId> "January"
mog onenote create-page <sectionId> "Meeting Notes" "Content here"

mog onenote delete <pageId>
mog onenote search "meeting"
```

---

### 📝 Word

```bash
mog word list                        # 列出文档
mog word export <id> --out ./doc.docx
mog word export <id> --format pdf --out ./doc.pdf
mog word copy <id> "Copy of Report"
```

---

### 📊 PowerPoint

```bash
mog ppt list                         # 列出演示文稿
mog ppt export <id> --out ./deck.pptx
mog ppt export <id> --format pdf --out ./deck.pdf
mog ppt copy <id> "Copy of Deck"
```

---

## 🔗 Slug 系统

Microsoft Graph 的 ID 非常长（100+ 字符）。`mog` 自动生成 8 字符的短码（Slug）：

```
完整 ID:  AQMkADAwATMzAGZmAS04MDViLTRiNzgtMDA...
Slug:     a3f2c891
```

- 所有命令默认输出 Slug
- 所有命令同时接受 Slug 或完整 ID
- 使用 `--verbose` 可同时查看完整 ID
- Slug 缓存在 `~/.config/mog/slugs.json`
- `mog auth logout` 会清空缓存

---

## 🏷️ 别名 (Aliases)

为常用 ID 或 Slug 创建易记的名称：

```bash
# 创建别名
mog alias set @standup f1a2b3c4
mog alias set @budget "AQMkADAwATMz..."

# 在任何需要 ID 的地方使用别名
mog calendar get @standup
mog excel get @budget Sheet1 A1:D10

# 管理别名
mog alias list
mog alias get @standup
mog alias rm @standup
```

- `@` 前缀在 Shell 中更安全，无需额外引号
- 别名解析链路：`@standup` → `f1a2b3c4` → 完整 ID
- 按账户保存在 `~/.config/mog/{account}/aliases.json`

---

## 🤖 AI 友好

运行 `mog --ai-help` 可查看完整文档，包括：

- 所有命令及其参数
- 日期/时间格式规范
- 正例和反例
- 退出码和管道模式
- 故障排除指南

遵循 [dashdash](https://github.com/visionik/dashdash) 规范。

---

## 🔄 gog 兼容性

`mog` 遵循 [gog](https://github.com/visionik/gog) 的命令风格，方便跨云服务形成肌肉记忆：

| 模式 | mog | gog |
|------|-----|-----|
| 日历事件 | `--summary`, `--from`, `--to` | 相同 |
| 任务备注 | `--notes` | 相同 |
| 输出格式 | `--json`, `--plain` | 相同 |
| 最大结果数 | `--max` | 相同 |
| Excel 读取 | `mog excel get <id> Sheet1 A1:D10` | `gog sheets get <id> Sheet1!A1:D10` |
| 表格写入 | `mog excel update <id> ...` | `gog sheets update <id> ...` |

---

## 🗂️ 配置文件

| 文件 | 用途 |
|------|------|
| `~/.config/mog/tokens.json` | OAuth Token（敏感信息） |
| `~/.config/mog/settings.json` | Client ID、Region 等配置 |
| `~/.config/mog/slugs.json` | ID 与 Slug 的映射缓存 |
| `~/.config/mog/aliases.json` | 自定义别名 |

**环境变量：**

| 变量 | 说明 |
|------|------|
| `MOG_CLIENT_ID` | Azure AD Client ID（替代 `--client-id`） |
| `MOG_REGION` | 登录区域：`global` 或 `china` |
| `MOG_ACCOUNT` | 当前使用的账户名 |

---

## 🛠️ 开发

```bash
# 使用 Taskfile（推荐）
task test              # 运行测试
task test:coverage     # 带覆盖率报告
task lint              # 静态检查
task fmt               # 格式化代码
task check             # 全部检查

# 或直接通过 Go
go test ./...
go build ./cmd/mog
```

---

## 📄 许可证

MIT License

---

## 👨‍💻 原作者

原项目由 **[visionik](mailto:visionik@pobox.com)** 和 **Vinston** 开发，遵循 MIT 协议开源。
