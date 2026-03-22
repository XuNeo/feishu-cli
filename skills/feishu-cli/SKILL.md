---
name: feishu-cli
description: >-
  飞书 CLI 全功能技能：IM 消息/群聊/私信、文档读写创建、电子表格操作、知识库、
  文档导入导出、搜索、文件管理、画板、日历、任务、权限管理、OAuth 认证。
  当用户需要操作飞书任何功能时都应使用本技能，包括：查聊天记录、找私信、列群聊、
  读取/创建/编辑飞书文档、操作电子表格（读写行列/筛选/样式）、导出文档为 Markdown、
  从 Markdown 导入创建文档、搜索文档和消息、上传下载文件、管理日历事件、
  创建任务、添加评论、设置权限、查用户信息、登录授权、Token 过期处理。
  遇到飞书相关任何操作时都应优先使用本技能，不要尝试自己调用飞书 API。
argument-hint: <功能模块> <操作>
user-invocable: true
allowed-tools: Bash, Read, Write
---

# 飞书 CLI 全功能参考

通过 feishu-cli 操作飞书所有模块。按需读取对应的 references 文件获取详细用法。

## 快速路由

| 用户需求 | 读取 |
|---------|------|
| 查聊天记录、私信、群聊列表、Reaction/Pin | `references/chat.md` |
| 发消息、回复、转发、消息卡片 | `references/msg-send.md` |
| 读取飞书文档内容 | `references/doc-read.md` |
| 创建/编辑飞书文档 | `references/doc-write.md` |
| 文档高级操作（高亮块/画板/批量更新） | `references/doc-guide.md` |
| 导出文档为 Markdown / Word / PDF | `references/export.md` |
| 从 Markdown / Word 导入创建文档（返回值含 doc_url） | `references/import.md` |
| 电子表格读写、行列操作、样式 | `references/toolkit.md` → sheet 模块 |
| 搜索文档/消息/应用 | `references/search.md` |
| 云空间文件管理、上传下载 | `references/toolkit.md` → file 模块 |
| 日历事件查看/创建/回复 | `references/toolkit.md` → calendar 模块 |
| 任务/任务清单管理 | `references/toolkit.md` → task 模块 |
| 知识库节点/空间操作 | `references/toolkit.md` → wiki 模块 |
| 画板操作、PlantUML 图表 | `references/board.md` |
| 文档/文件权限设置 | `references/perm.md` |
| 登录、Token 管理、scope 配置 | `references/auth.md` |
| 话题群线程回复获取 | `references/thread-replies.md` |

---

## 配置（两种方式）

```bash
# 方式 1：环境变量（推荐）
export FEISHU_APP_ID="cli_xxx"
export FEISHU_APP_SECRET="xxx"

# 方式 2：配置文件
feishu-cli config init
```

大多数命令使用 **App Token**（应用身份），无需登录。
读取消息/群聊/搜索需要 **User Token**，需先登录：

```bash
# SSH 远程环境推荐 Device Flow
feishu-cli auth login --method device --scopes \
  "im:message:readonly im:message.p2p_msg:get_as_user im:message.group_msg:get_as_user \
   im:chat:read im:chat:readonly offline_access"
```

Token 过期/权限不足（99991679）→ 读 `references/auth.md`。

---

## 常用命令速查

### IM 消息与群聊
```bash
feishu-cli msg history --container-id oc_xxx --container-id-type chat -o json
feishu-cli msg p2p-chat-id --open-id ou_xxx -o json
feishu-cli msg list-chats --limit 20 -o json
feishu-cli msg send --receive-id-type email --receive-id user@co.com --text "你好"
feishu-cli user me -o json
```

### 文档
```bash
feishu-cli doc create --title "新文档"
feishu-cli doc export <doc_id> --output doc.md
feishu-cli doc import doc.md --title "从 Markdown 导入"
feishu-cli doc get <doc_id> -o json
```

### 电子表格
```bash
feishu-cli sheet read <sheet_token> --sheet-id <id> --range A1:D10
feishu-cli sheet write <sheet_token> --sheet-id <id> --range A1 --values '[[1,2],[3,4]]'
feishu-cli sheet add-rows <sheet_token> --sheet-id <id> --start-row 5 --count 3
feishu-cli sheet get <sheet_token>           # 获取表格基本信息
feishu-cli sheet list-sheets <sheet_token>   # 列出所有工作表
```

### 文件管理
```bash
feishu-cli file list --folder-token <token>
feishu-cli file upload /path/to/file.pdf --folder-token <token>
feishu-cli file download <file_token> --output ./file.pdf
```

### 搜索
```bash
feishu-cli search docs "关键词" -o json
feishu-cli search messages "关键词" --chat-type p2p_chat -o json
```

---

## 命令分工（避免混淆）

| 命令 | 身份 | 说明 |
|------|:---:|------|
| `msg send/reply/forward` | App | 发消息，不需要登录 |
| `msg history/list/get` | User | 读消息，需要 `auth login` |
| `msg p2p-chat-id` | User | 获取私信 chat_id |
| `doc create/import` | App | 创建文档 |
| `doc export/get` | App/User | 读文档内容 |
| `sheet read/write` | App | 电子表格读写 |
| `search docs/messages` | User | 搜索，需要 `auth login` |
| `chat create/link` | App | 创建群聊 |
| `chat get/member list` | User | 查看群信息 |
