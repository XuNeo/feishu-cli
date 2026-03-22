package cmd

import (
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var msgP2PChatIDCmd = &cobra.Command{
	Use:   "p2p-chat-id",
	Short: "获取与指定用户的私信 chat_id",
	Long: `通过对方的 open_id 获取你与该用户之间私信会话的 chat_id。

获取到 chat_id 后可配合 msg list 读取私信历史消息:
  feishu-cli msg list --container-id <chat_id> --user-access-token <token>

注意:
  - 需要 User Access Token，返回的是你（当前授权用户）与对方的私信
  - 与 /chat/v3/p2p/id 不同，该接口返回用户间私信而非 Bot 与用户的私信
  - 支持批量查询，一次最多传入多个 open_id

示例:
  # 获取与指定用户的私信 chat_id
  feishu-cli msg p2p-chat-id --open-id ou_xxx

  # 批量查询多个用户
  feishu-cli msg p2p-chat-id --open-id ou_xxx --open-id ou_yyy

  # JSON 格式输出
  feishu-cli msg p2p-chat-id --open-id ou_xxx -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		token, err := resolveRequiredUserToken(cmd)
		if err != nil {
			return err
		}

		openIDs, _ := cmd.Flags().GetStringArray("open-id")
		if len(openIDs) == 0 {
			return fmt.Errorf("至少需要指定一个 --open-id")
		}

		output, _ := cmd.Flags().GetString("output")

		chats, err := client.GetP2PChatIDs(openIDs, token)
		if err != nil {
			return err
		}

		if output == "json" {
			return printJSON(chats)
		}

		if len(chats) == 0 {
			fmt.Println("未找到对应的私信会话")
			return nil
		}

		fmt.Printf("找到 %d 个私信会话:\n\n", len(chats))
		for i, chat := range chats {
			fmt.Printf("[%d] Chat ID:  %s\n", i+1, chat.ChatID)
			fmt.Printf("    用户 1:   %s\n", chat.ChatterID1)
			fmt.Printf("    用户 2:   %s\n", chat.ChatterID2)
			fmt.Println()
		}
		return nil
	},
}

func init() {
	msgCmd.AddCommand(msgP2PChatIDCmd)
	msgP2PChatIDCmd.Flags().StringArray("open-id", []string{}, "对方的 open_id（可多次指定）")
	msgP2PChatIDCmd.Flags().StringP("output", "o", "", "输出格式（json）")
	msgP2PChatIDCmd.Flags().String("user-access-token", "", "User Access Token（用户授权令牌）")
}
