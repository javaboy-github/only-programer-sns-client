package user

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/javaboy-github/only-programer-sns-client/util"
	"github.com/spf13/cobra"
)

func UserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "user",
		Short: "ユーザーを登録/更新/閲覧します。",
	}
	cmd.AddCommand(createCmd())
	return cmd
}

func createCmd() *cobra.Command {
	cmd := &cobra.Command {
		Use: "create",
		Short: "アカウントを作成します。",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ユーザー名を入力[二回改行で決定][30字]")
			name := util.GetText()
			if utf8.RuneCountInString(name) > 30 {
				color.Red("30字を超えています。")
				return
			}
			fmt.Println("自己紹介を入力[二回改行で決定][300字]")
			profile := util.GetText()
			if utf8.RuneCountInString(profile) > 300 {
				color.Red("300字を超えています")
				return
			}
			http.Post("https://versatileapi.herokuapp.com/api", "text/plain", strings.NewReader(fmt.Sprintf("{\"name\":\"%s\",\"description\":\"%s\"}", name, profile)))
			fmt.Println("完了しました")
		},
	}
	return cmd
}
