# 话题群消息获取

话题群（每条消息都有 `thread_id`）的消息获取需要两步，因为主消息流只包含每个话题的首条消息，回复内容在线程里。

## 判断是否是话题群

获取几条消息后检查字段：
- 话题群：**每条**消息都有 `thread_id`（形如 `omt_xxx`）
- 普通群：独立消息**无** `thread_id`，只有被回复/回复消息才有

| 字段 | 说明 |
|------|------|
| `thread_id` | 线程 ID，形如 `omt_xxx` |
| `root_id` | 线程根消息 ID（首条消息） |
| `parent_id` | 直接上级消息 ID |

普通群不需要这一步——`msg history` 会把所有消息（含线程回复）平铺返回。

## 获取话题群完整内容

```bash
# 步骤 1：获取主消息流（每个话题的首条消息）
feishu-cli msg history \
  --container-id oc_xxx \
  --container-id-type chat \
  --page-size 50 \
  --sort-type ByCreateTimeDesc \
  -o json

# 步骤 2：对每个 thread_id 获取回复（并行执行更快）
feishu-cli msg history \
  --container-id omt_xxx \
  --container-id-type thread \
  --page-size 50 \
  --sort-type ByCreateTimeAsc \
  -o json
```

话题群中活跃话题可能有 10-20 个，并行获取比串行快很多。飞书对 `msg history` 的频率限制较宽松，10-20 个并发没有问题。
