# 飞书 IM 会话与认证

通过 feishu-cli 读取飞书消息、管理群聊、查询用户身份。

> 认证问题（Token 过期、scope 不足、99991679 错误）→ 读 `references/auth.md`

---

## 前置条件：登录

所有读取/管理命令需要 User Token。先检查登录状态：

```bash
feishu-cli auth status -o json
# logged_in: false 或 access_token_valid: false → 需要重新登录
```

SSH 远程环境推荐 Device Flow（不需要配置回调地址）：

```bash
feishu-cli auth login --method device --scopes \
  "im:message:readonly im:message.p2p_msg:get_as_user im:message.group_msg:get_as_user \
   im:chat:read im:chat:readonly offline_access"
```

详细登录方式、scope 配置、Token 刷新机制 → 读 `references/auth.md`。

---

## 常用场景速查

| 目标 | 命令 |
|------|------|
| 看某群最近消息 | `msg history --container-id oc_xxx --container-id-type chat` |
| 看和某人的私信 | `msg p2p-chat-id --open-id ou_xxx` → `msg history --container-id oc_xxx` |
| 找群聊 | `msg search-chats --query "关键词"` |
| 列出最近活跃的群 | `msg list-chats --limit 20` |
| 查自己的 open_id | `user me -o json` |
| 查对方 open_id | `user search --email xxx@company.com` |
| 下载消息图片/文件 | `msg download-resource --message-id om_xxx --file-key img_xxx --type image` |
| 群成员列表 | `chat member list oc_xxx` |
| 消息加 Reaction | `msg reaction add om_xxx --emoji-type THUMBSUP` |

---

## 场景一：查看群聊消息

### 找到群聊

```bash
# 按名字搜索
feishu-cli msg search-chats --query "群名关键词" -o json

# 列出最近活跃的群（快，推荐）
feishu-cli msg list-chats --limit 20 -o json

# 全量枚举（群多时慢）
feishu-cli msg list-chats -o json

# 按成员数过滤（会额外调详情接口，稍慢）
feishu-cli msg list-chats --limit 50 --max-members 150 -o json
```

`--limit N` 会自动按活跃时间排序，拿够 N 个即停止，比全量枚举快很多。
外部群（`external=true`）无法获取成员数（API 限制 232033），`user_count=-1`，不参与 `--max-members` 过滤。

### 获取消息

```bash
# 最近 50 条
feishu-cli msg history \
  --container-id oc_xxx \
  --container-id-type chat \
  --page-size 50 --sort-type ByCreateTimeDesc \
  -o json

# 指定时间范围（毫秒时间戳）
feishu-cli msg history \
  --container-id oc_xxx --container-id-type chat \
  --start-time 1773778860000 \
  -o json

# 翻页
feishu-cli msg history \
  --container-id oc_xxx --container-id-type chat \
  --page-token "上次返回的token" \
  -o json
```

话题群需要额外获取线程回复 → 读 `references/thread-replies.md`

### 解析消息内容

`body.content` 是 JSON 字符串：

| msg_type | content |
|----------|---------|
| text | `{"text":"内容"}` |
| image | `{"image_key":"img_xxx"}` |
| file | `{"file_key":"xxx","file_name":"..."}` |
| interactive | 卡片 JSON |

```python
import json
content = json.loads(msg['body']['content'])
text = content.get('text', '')
```

---

## 场景二：查看私聊记录

飞书私信需要先拿到 chat_id，再读历史。

```bash
# 步骤 1：通过对方 open_id 获取私信 chat_id
feishu-cli msg p2p-chat-id --open-id ou_xxx -o json
# 返回: {"chat_id": "oc_xxx", "chatter_id1": "...", "chatter_id2": "..."}

# 步骤 2：读取私信历史
feishu-cli msg history \
  --container-id oc_xxx --container-id-type chat \
  --page-size 50 --sort-type ByCreateTimeDesc \
  -o json
```

`msg p2p-chat-id` 返回的是**你与对方之间**的真实私信 chat_id（不是 Bot 与对方的），底层调用 `POST /im/v1/chat_p2p/batch_query`。

不知道对方 open_id？用邮箱查：
```bash
feishu-cli user search --email user@company.com -o json
```

不知道自己的 open_id？
```bash
feishu-cli user me -o json
# 返回: {"open_id": "ou_xxx", "name": "你的名字", "email": "..."}
```

---

## 场景三：下载消息附件

消息里的图片、文件需要用 `msg download-resource` 下载（不是 `media download`，两者是不同接口）。

```bash
# 先获取消息详情，从 body.content 拿到 image_key 或 file_key
feishu-cli msg get om_xxx -o json

# 下载图片（--type image）
feishu-cli msg download-resource \
  --message-id om_xxx \
  --file-key img_xxx \
  --type image \
  --output ./photo.png

# 下载文件/音频/视频（--type file）
feishu-cli msg download-resource \
  --message-id om_xxx \
  --file-key file_xxx \
  --type file \
  --output ./document.pdf
```

---

## 场景四：群聊信息管理

```bash
# 查群详情
feishu-cli chat get oc_xxx -o json

# 查群成员
feishu-cli chat member list oc_xxx --page-size 100

# 改群名/描述/群主
feishu-cli chat update oc_xxx --name "新群名"
feishu-cli chat update oc_xxx --owner-id ou_xxx

# 添加/移除成员
feishu-cli chat member add oc_xxx --id-list ou_xxx,ou_yyy
feishu-cli chat member remove oc_xxx --id-list ou_xxx

# 解散群（不可恢复）
feishu-cli chat delete oc_xxx
```

> `chat create` 和 `chat link` 仅支持 App Token（Bot 身份），不需要 User Token。

---

## 场景五：消息互动

```bash
# 获取单条消息
feishu-cli msg get om_xxx -o json

# 查置顶消息
feishu-cli msg pins --chat-id oc_xxx -o json

# 置顶/取消置顶
feishu-cli msg pin om_xxx
feishu-cli msg unpin om_xxx

# Reaction（常用类型：THUMBSUP SMILE HEART JIAYI OK FIRE）
feishu-cli msg reaction add om_xxx --emoji-type THUMBSUP
feishu-cli msg reaction remove om_xxx --reaction-id <id>
feishu-cli msg reaction list om_xxx

# 删除消息（仅 Bot 自己发的，不可恢复）
feishu-cli msg delete om_xxx
```

---

## 处理大量消息

获取 100+ 条消息时：
1. 每页 `-o json` 输出重定向到文件，用 `page_token` 翻页直到 `has_more=false`
2. `create_time` 是毫秒时间戳，除以 1000 转秒
3. `deleted: true` 的撤回消息内容是 `"This message was recalled"`，汇总时跳过
4. 搜索 API 有频率限制，大量请求间加 1s 延迟；`msg history` 限频宽松，可安全并发

```python
from datetime import datetime
ts = int(msg['create_time']) / 1000
time_str = datetime.fromtimestamp(ts).strftime('%Y-%m-%d %H:%M')
content = json.loads(msg['body']['content'])
text = content.get('text', '')
```

---

## 权限速查

| 命令 | 需要的 scope |
|------|-------------|
| `msg history/get/list` | `im:message:readonly` + `im:message.group_msg:get_as_user` |
| `msg p2p-chat-id` | `im:message.p2p_msg:get_as_user` |
| `msg list-chats` / `msg search-chats` | `im:chat:read` |
| `msg download-resource` | `im:message:readonly` |
| `user me` | `contact:user.base:readonly` |
| `chat get` / `chat member list` | `im:chat:readonly` |
| `chat update/delete` | `im:chat` |
| `msg pin/unpin` | `im:message.pins` |
| `msg reaction` | `im:message.reactions` |
| `search messages` | `search:message` |

scope 不足报 99991679 → 重新 `auth login` 并在 `--scopes` 里加上缺失的权限。完整说明 → `references/auth.md`。

---

## 与其他技能分工

| 场景 | 技能 |
|------|------|
| 发消息、回复、转发 | feishu-cli-msg |
| 搜索文档、高级消息搜索 | feishu-cli-search |
| 表格、日历、任务、知识库 | feishu-cli-toolkit |
