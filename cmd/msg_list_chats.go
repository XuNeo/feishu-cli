package cmd

import (
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var msgListChatsCmd = &cobra.Command{
	Use:   "list-chats",
	Short: "枚举当前用户所有群聊",
	Long: `获取当前授权用户加入的所有群聊列表。

与 search-chats 的区别:
  search-chats  按关键词搜索，不能全量枚举
  list-chats    全量枚举，支持按活跃时间排序和 --limit 快速截断

性能建议:
  不加 --limit 时会翻页拉取全部群聊，群多时较慢。
  加上 --limit N 后自动切换为按活跃时间降序排序，拿够 N 个即停止，
  适合"最近活跃的 N 个群"场景，速度远快于全量枚举。

示例:
  # 全量枚举（可能较慢）
  feishu-cli msg list-chats -o json

  # 最近活跃的 20 个群（快速）
  feishu-cli msg list-chats --limit 20 -o json

  # 跳过超过 150 人的大群
  feishu-cli msg list-chats --max-members 150 -o json

  # 最近活跃的 50 个群，跳过大群
  feishu-cli msg list-chats --limit 50 --max-members 150 -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		token, err := resolveRequiredUserToken(cmd)
		if err != nil {
			return err
		}

		maxMembers, _ := cmd.Flags().GetInt("max-members")
		limit, _ := cmd.Flags().GetInt("limit")
		sortType, _ := cmd.Flags().GetString("sort-type")
		output, _ := cmd.Flags().GetString("output")

		opts := client.ListAllChatsOptions{
			MaxMembers: maxMembers,
			Limit:      limit,
			SortType:   sortType,
		}

		result, err := client.GetAllChats(opts, token)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(result)
		}

		fmt.Printf("找到 %d 个群聊", len(result.Items))
		if result.SkippedLarge > 0 {
			fmt.Printf("（已跳过 %d 个超过 %d 人的大群）", result.SkippedLarge, maxMembers)
		}
		fmt.Println(":")

		for i, chat := range result.Items {
			countStr := ""
			if maxMembers > 0 {
				if chat.UserCount == -1 {
					countStr = "  人数未知（外部群）"
				} else {
					countStr = fmt.Sprintf("  %d 人", chat.UserCount)
				}
			}
			externalMark := ""
			if chat.External {
				externalMark = "  [外部群]"
			}
			fmt.Printf("\n[%d] %s%s\n", i+1, chat.Name, externalMark)
			fmt.Printf("    群聊 ID:  %s\n", chat.ChatID)
			if countStr != "" {
				fmt.Printf("    成员数:  %s\n", countStr)
			}
			fmt.Printf("    类型:     %s\n", chat.ChatMode)
			if chat.Description != "" {
				fmt.Printf("    描述:     %s\n", chat.Description)
			}
		}

		if result.SkippedLarge > 0 {
			fmt.Printf("\n提示: 已跳过 %d 个成员数超过 %d 的大群，使用 --max-members 0 可包含所有群\n",
				result.SkippedLarge, maxMembers)
		}
		return nil
	},
}

func init() {
	msgCmd.AddCommand(msgListChatsCmd)
	msgListChatsCmd.Flags().Int("max-members", 0, "跳过成员数超过此值的大群（0 表示不过滤）")
	msgListChatsCmd.Flags().Int("limit", 0, "最多返回 N 个群（0 表示全量，指定时自动按活跃时间排序）")
	msgListChatsCmd.Flags().String("sort-type", "", "排序方式：ByCreateTimeAsc（默认）/ ByActiveTimeDesc（指定 --limit 时自动使用）")
	msgListChatsCmd.Flags().StringP("output", "o", "", "输出格式（json）")
	msgListChatsCmd.Flags().String("user-access-token", "", "User Access Token（用户授权令牌）")
}
