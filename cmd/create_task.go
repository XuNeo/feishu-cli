package cmd

import (
	"fmt"
	"time"

	"github.com/riba2534/feishu-cli/internal/client"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var createTaskCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新任务",
	Long: `创建新的飞书任务。

参数:
  --summary, -s       任务标题（必填）
  --description, -d   任务描述
  --due               截止时间（格式: 2006-01-02 15:04:05 或 2006-01-02）
  --assignee          负责人 open_id（可多次指定）
  --tasklist          加入的任务清单 GUID（可多次指定；若清单已关联群聊则成为群任务）
  --origin-href       任务来源链接
  --origin-platform   任务来源平台名称（默认: feishu-cli）
  --output, -o        输出格式（json）

群任务说明:
  飞书"群任务"本质是把任务加入一个与群聊关联的任务清单。
  先用 tasklist create --chat-id <oc_xxx> 创建关联群聊的清单，
  再用 task create --tasklist <tasklist_guid> 把任务加入该清单。

示例:
  # 创建简单任务
  feishu-cli task create --summary "完成项目文档"

  # 创建带负责人和截止时间的任务
  feishu-cli task create --summary "代码审查" --assignee ou_xxx --due "2024-12-31 18:00:00"

  # 创建群任务（加入已关联群聊的清单）
  feishu-cli task create --summary "群任务标题" --tasklist <tasklist_guid>

  # 创建带来源链接的任务
  feishu-cli task create --summary "处理 Issue" --origin-href "https://github.com/example/repo/issues/1"

  # JSON 格式输出
  feishu-cli task create --summary "测试任务" --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		token := resolveOptionalUserToken(cmd)

		summary, _ := cmd.Flags().GetString("summary")
		description, _ := cmd.Flags().GetString("description")
		dueStr, _ := cmd.Flags().GetString("due")
		originHref, _ := cmd.Flags().GetString("origin-href")
		originPlatform, _ := cmd.Flags().GetString("origin-platform")
		assignees, _ := cmd.Flags().GetStringArray("assignee")
		tasklists, _ := cmd.Flags().GetStringArray("tasklist")

		opts := client.CreateTaskOptions{
			Summary:        summary,
			Description:    description,
			OriginHref:     originHref,
			OriginPlatform: originPlatform,
			Assignees:      assignees,
			TasklistGuids:  tasklists,
		}

		// Parse due time
		if dueStr != "" {
			dueTime, err := parseTime(dueStr)
			if err != nil {
				return fmt.Errorf("解析截止时间失败: %w", err)
			}
			opts.DueTimestamp = dueTime.UnixMilli()
		}

		task, err := client.CreateTask(opts, token)
		if err != nil {
			return err
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			if err := printJSON(task); err != nil {
				return err
			}
		} else {
			fmt.Printf("任务创建成功！\n")
			fmt.Printf("  任务 ID: %s\n", task.Guid)
			fmt.Printf("  标题: %s\n", task.Summary)
			if task.Description != "" {
				fmt.Printf("  描述: %s\n", task.Description)
			}
			if task.DueTime != "" {
				fmt.Printf("  截止时间: %s\n", task.DueTime)
			}
			if task.OriginHref != "" {
				fmt.Printf("  来源链接: %s\n", task.OriginHref)
			}
		}

		return nil
	},
}

// parseTime parses a time string in various formats
func parseTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006/01/02",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法识别的时间格式: %s", s)
}

func init() {
	taskCmd.AddCommand(createTaskCmd)
	createTaskCmd.Flags().StringP("summary", "s", "", "任务标题（必填）")
	createTaskCmd.Flags().StringP("description", "d", "", "任务描述")
	createTaskCmd.Flags().String("due", "", "截止时间（格式: 2006-01-02 15:04:05）")
	createTaskCmd.Flags().StringArray("assignee", []string{}, "负责人 open_id（可多次指定，如 --assignee ou_xxx --assignee ou_yyy）")
	createTaskCmd.Flags().StringArray("tasklist", []string{}, "加入的任务清单 GUID（可多次指定；若清单关联了群聊则成为群任务）")
	createTaskCmd.Flags().String("origin-href", "", "任务来源链接")
	createTaskCmd.Flags().String("origin-platform", "", "任务来源平台名称")
	createTaskCmd.Flags().StringP("output", "o", "", "输出格式（json）")
	createTaskCmd.Flags().String("user-access-token", "", "User Access Token（用户授权令牌）")
	mustMarkFlagRequired(createTaskCmd, "summary")
}
