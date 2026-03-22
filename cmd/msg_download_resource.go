package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var msgDownloadResourceCmd = &cobra.Command{
	Use:   "download-resource",
	Short: "下载消息中的图片或文件附件",
	Long: `下载飞书消息中的附件资源（图片、文件、音频、视频）。

与 media download 的区别:
  media download         下载飞书云空间中的素材文件（需要 file_token）
  msg download-resource  下载消息中的附件（需要 message_id + file_key）

如何获取 file_key:
  通过 msg get <message_id> 获取消息详情，从 body.content 中提取:
  - 图片: image_key 字段（img_xxx 格式）
  - 文件: file_key 字段（file_xxx 格式）
  - 音频/视频: file_key 字段

资源类型 (--type):
  image   图片
  file    文件、音频、视频

示例:
  # 下载消息中的图片（保存到当前目录）
  feishu-cli msg download-resource --message-id om_xxx --file-key img_xxx --type image

  # 下载消息中的文件
  feishu-cli msg download-resource --message-id om_xxx --file-key file_xxx --type file

  # 指定输出路径
  feishu-cli msg download-resource --message-id om_xxx --file-key img_xxx --type image --output ./photo.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		token, err := resolveRequiredUserToken(cmd)
		if err != nil {
			return err
		}

		messageID, _ := cmd.Flags().GetString("message-id")
		fileKey, _ := cmd.Flags().GetString("file-key")
		resourceType, _ := cmd.Flags().GetString("type")
		outputPath, _ := cmd.Flags().GetString("output")

		data, err := client.DownloadMessageResource(messageID, fileKey, resourceType, token)
		if err != nil {
			return err
		}

		// 确定输出路径
		if outputPath == "" {
			ext := ""
			if resourceType == "image" {
				ext = ".png"
			}
			outputPath = fileKey + ext
		}

		// 确保目录存在
		dir := filepath.Dir(outputPath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}
		}

		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}

		fmt.Printf("已下载到: %s（%d bytes）\n", outputPath, len(data))
		return nil
	},
}

func init() {
	msgCmd.AddCommand(msgDownloadResourceCmd)
	msgDownloadResourceCmd.Flags().String("message-id", "", "消息 ID（om_xxx 格式）")
	msgDownloadResourceCmd.Flags().String("file-key", "", "附件 key（img_xxx 或 file_xxx 格式）")
	msgDownloadResourceCmd.Flags().String("type", "image", "资源类型（image/file）")
	msgDownloadResourceCmd.Flags().StringP("output", "o", "", "输出文件路径（默认使用 file_key 作为文件名）")
	msgDownloadResourceCmd.Flags().String("user-access-token", "", "User Access Token（用户授权令牌）")
	mustMarkFlagRequired(msgDownloadResourceCmd, "message-id", "file-key")
}
