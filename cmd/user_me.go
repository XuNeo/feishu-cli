package cmd

import (
	"fmt"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var userMeCmd = &cobra.Command{
	Use:   "me",
	Short: "获取当前授权用户信息",
	Long: `获取当前 User Access Token 对应的用户信息。

输出字段:
  open_id     用户的 Open ID（ou_xxx 格式，用于消息 API）
  union_id    用户的 Union ID
  name        用户姓名
  en_name     英文名
  email       邮箱
  avatar_url  头像 URL
  tenant_key  所在租户 Key

示例:
  # 获取当前用户信息
  feishu-cli user me

  # JSON 格式输出（适合脚本使用）
  feishu-cli user me -o json

  # 手动指定 Token
  feishu-cli user me --user-access-token u-xxx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		token, err := resolveRequiredUserToken(cmd)
		if err != nil {
			return err
		}

		user, err := client.GetCurrentUser(token)
		if err != nil {
			return err
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			return printJSON(user)
		}

		fmt.Printf("当前用户信息:\n\n")
		fmt.Printf("  姓名:      %s\n", user.Name)
		if user.EnName != "" {
			fmt.Printf("  英文名:    %s\n", user.EnName)
		}
		fmt.Printf("  Open ID:   %s\n", user.OpenID)
		if user.UnionID != "" {
			fmt.Printf("  Union ID:  %s\n", user.UnionID)
		}
		if user.Email != "" {
			fmt.Printf("  邮箱:      %s\n", user.Email)
		}
		if user.Mobile != "" {
			fmt.Printf("  手机:      %s\n", user.Mobile)
		}
		if user.TenantKey != "" {
			fmt.Printf("  租户 Key:  %s\n", user.TenantKey)
		}
		return nil
	},
}

func init() {
	userCmd.AddCommand(userMeCmd)
	userMeCmd.Flags().String("user-access-token", "", "User Access Token（用户授权令牌）")
	userMeCmd.Flags().StringP("output", "o", "", "输出格式（json）")
}
