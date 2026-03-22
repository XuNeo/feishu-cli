package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/riba2534/feishu-cli/internal/config"
)

// CurrentUser 当前授权用户信息
type CurrentUser struct {
	OpenID    string `json:"open_id"`
	UnionID   string `json:"union_id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	EnName    string `json:"en_name"`
	Email     string `json:"email"`
	Mobile    string `json:"mobile"`
	AvatarURL string `json:"avatar_url"`
	TenantKey string `json:"tenant_key"`
}

// GetCurrentUser 获取当前授权用户信息（User Token）
// API: GET /authen/v1/user_info
func GetCurrentUser(userAccessToken string) (*CurrentUser, error) {
	cfg := getBaseURL()
	reqURL := fmt.Sprintf("%s/open-apis/authen/v1/user_info", cfg)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("获取当前用户信息失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+userAccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取当前用户信息失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("获取当前用户信息失败: 读取响应失败: %w", err)
	}

	var result struct {
		Code int         `json:"code"`
		Msg  string      `json:"msg"`
		Data CurrentUser `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("获取当前用户信息失败: 解析响应失败: %w", err)
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("获取当前用户信息失败: code=%d, msg=%s", result.Code, result.Msg)
	}

	return &result.Data, nil
}

// ChatItem 群聊信息
type ChatItem struct {
	ChatID      string `json:"chat_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
	ChatMode    string `json:"chat_mode"`
	ChatType    string `json:"chat_type"`
	UserCount   int    `json:"user_count"` // 仅在指定 --max-members 时填充，否则为 0
	External    bool   `json:"external"`
}

// getChatUserCount 获取单个群的成员数
// API: GET /im/v1/chats/{chat_id}
func getChatUserCount(baseURL, chatID, userAccessToken string) (int, error) {
	reqURL := fmt.Sprintf("%s/open-apis/im/v1/chats/%s?user_id_type=open_id", baseURL, chatID)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+userAccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var raw struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			UserCount string `json:"user_count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0, err
	}
	if raw.Code != 0 {
		return 0, fmt.Errorf("code=%d, msg=%s", raw.Code, raw.Msg)
	}
	count, _ := strconv.Atoi(raw.Data.UserCount)
	return count, nil
}

// ListAllChatsOptions 枚举群聊选项
type ListAllChatsOptions struct {
	MaxMembers int    // 跳过成员数超过此值的群（0 表示不过滤）
	Limit      int    // 最多返回 N 个群（0 表示不限制，全量返回）；指定时自动按活跃时间降序排列
	SortType   string // 排序方式：ByCreateTimeAsc（默认）/ ByActiveTimeDesc
	PageSize   int    // 每页数量（最大 100）
}

// ListAllChatsResult 枚举群聊结果
type ListAllChatsResult struct {
	Items        []ChatItem `json:"items"`
	SkippedLarge int        `json:"skipped_large"` // 因成员数过多被跳过的群数量
}

// GetAllChats 获取当前用户所有群聊（自动翻页）
// API: GET /im/v1/chats
// 当 opts.Limit > 0 时，自动切换为 ByActiveTimeDesc 排序，拿够即停，避免全量翻页
func GetAllChats(opts ListAllChatsOptions, userAccessToken string) (*ListAllChatsResult, error) {
	cfg := getBaseURL()
	pageSize := opts.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 100
	}

	// 指定 Limit 时按活跃时间降序，拿到足够数量即停止翻页
	sortType := opts.SortType
	if sortType == "" {
		if opts.Limit > 0 {
			sortType = "ByActiveTimeDesc"
		} else {
			sortType = "ByCreateTimeAsc"
		}
	}

	result := &ListAllChatsResult{}
	pageToken := ""

	for {
		params := url.Values{}
		params.Set("page_size", strconv.Itoa(pageSize))
		params.Set("user_id_type", "open_id")
		params.Set("sort_type", sortType)
		if pageToken != "" {
			params.Set("page_token", pageToken)
		}

		reqURL := fmt.Sprintf("%s/open-apis/im/v1/chats?%s", cfg, params.Encode())
		req, err := http.NewRequest(http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("获取群聊列表失败: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+userAccessToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("获取群聊列表失败: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("获取群聊列表失败: 读取响应失败: %w", err)
		}

		var raw struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data struct {
				HasMore   bool   `json:"has_more"`
				PageToken string `json:"page_token"`
				Items     []struct {
					ChatID      string `json:"chat_id"`
					Name        string `json:"name"`
					Description string `json:"description"`
					OwnerID     string `json:"owner_id"`
					ChatMode    string `json:"chat_mode"`
					ChatType    string `json:"chat_type"`
					External    bool   `json:"external"`
				} `json:"items"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &raw); err != nil {
			return nil, fmt.Errorf("获取群聊列表失败: 解析响应失败: %w", err)
		}
		if raw.Code != 0 {
			return nil, fmt.Errorf("获取群聊列表失败: code=%d, msg=%s", raw.Code, raw.Msg)
		}

		for _, item := range raw.Data.Items {
			chatItem := ChatItem{
				ChatID:      item.ChatID,
				Name:        item.Name,
				Description: item.Description,
				OwnerID:     item.OwnerID,
				ChatMode:    item.ChatMode,
				ChatType:    item.ChatType,
				External:    item.External,
			}

			// 仅在指定 MaxMembers 时才拉详情获取 user_count
			if opts.MaxMembers > 0 {
				if item.External {
					// 外部群无法通过详情接口获取 user_count（返回 232033 权限错误）
					// user_count 标记为 -1（未知），不参与过滤，始终保留
					chatItem.UserCount = -1
				} else {
					userCount, err := getChatUserCount(cfg, item.ChatID, userAccessToken)
					if err == nil {
						chatItem.UserCount = userCount
					}
					if userCount > opts.MaxMembers {
						result.SkippedLarge++
						continue
					}
				}
			}

			result.Items = append(result.Items, chatItem)
			// 达到 Limit 立即停止
			if opts.Limit > 0 && len(result.Items) >= opts.Limit {
				return result, nil
			}
		}

		if !raw.Data.HasMore || raw.Data.PageToken == "" {
			break
		}
		pageToken = raw.Data.PageToken
	}

	return result, nil
}

// P2PChatItem 私信会话信息
type P2PChatItem struct {
	ChatID     string `json:"chat_id"`
	ChatterID1 string `json:"chatter_id1"`
	ChatterID2 string `json:"chatter_id2"`
}

// GetP2PChatIDs 通过对方 open_id 批量获取私信 chat_id
// API: POST /im/v1/chat_p2p/batch_query
// 注意：返回的是当前授权用户与对方的私信，不是 Bot 与对方的私信
func GetP2PChatIDs(chatterIDs []string, userAccessToken string) ([]P2PChatItem, error) {
	cfg := getBaseURL()
	reqURL := fmt.Sprintf("%s/open-apis/im/v1/chat_p2p/batch_query?user_id_type=open_id", cfg)

	bodyData := map[string]any{
		"chatter_ids": chatterIDs,
	}
	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return nil, fmt.Errorf("获取私信 chat_id 失败: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("获取私信 chat_id 失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+userAccessToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取私信 chat_id 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("获取私信 chat_id 失败: 读取响应失败: %w", err)
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			P2PChats []P2PChatItem `json:"p2p_chats"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("获取私信 chat_id 失败: 解析响应失败: %w", err)
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("获取私信 chat_id 失败: code=%d, msg=%s", result.Code, result.Msg)
	}

	return result.Data.P2PChats, nil
}

// getBaseURL 获取飞书 API 基础 URL
func getBaseURL() string {
	cfg := config.Get()
	if cfg.BaseURL != "" {
		return cfg.BaseURL
	}
	return "https://open.feishu.cn"
}
