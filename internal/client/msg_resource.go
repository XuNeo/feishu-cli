package client

import (
	"fmt"
	"io"
	"net/http"
)

// DownloadMessageResource 下载消息中的图片或文件附件
// API: GET /im/v1/messages/{message_id}/resources/{file_key}?type=image|file
// 与 media download 不同，此接口专门用于下载消息中的附件资源
func DownloadMessageResource(messageID, fileKey, resourceType, userAccessToken string) ([]byte, error) {
	cfg := getBaseURL()
	reqURL := fmt.Sprintf("%s/open-apis/im/v1/messages/%s/resources/%s?type=%s",
		cfg, messageID, fileKey, resourceType)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("下载消息附件失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+userAccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载消息附件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("下载消息附件失败: HTTP %d, body=%s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("下载消息附件失败: 读取响应失败: %w", err)
	}

	return data, nil
}
