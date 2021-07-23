package user

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/javaboy-github/only-programer-sns-client/util"
	"github.com/spf13/cobra"
)

func UserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "ユーザーを登録/更新/閲覧します。",
	}
	cmd.AddCommand(createCmd())
	cmd.AddCommand(printAllUserCmd())
	cmd.AddCommand(updateUsersCmd())
	return cmd
}

func createCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "アカウントを作成します。",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ユーザー名を入力[二回改行で決定][30字]")
			name := util.GetText()
			name = util.StringToJsonString(name)
			if utf8.RuneCountInString(name) > 30 {
				color.Red("30字を超えています。")
				return
			}
			fmt.Println("自己紹介を入力[二回改行で決定][300字]")
			profile := util.GetText()
			profile = util.StringToJsonString(profile)
			if utf8.RuneCountInString(profile) > 300 {
				color.Red("300字を超えています")
				return
			}
			resp, _ := http.Post("https://versatileapi.herokuapp.com/api/user/create_user", "text/plain", strings.NewReader(fmt.Sprintf("{\"name\":\"%s\",\"description\":\"%s\"}", name, profile)))
			defer resp.Body.Close()
			fmt.Println("完了しました")
		},
	}
	return cmd
}

func updateUsers() map[string]string {
	users := map[string]string{}
	// ユーザー一覧を取得
	{
		resp, _ := http.Get("https://versatileapi.herokuapp.com/api/user/all")
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		io.ReadAll(resp.Body)
		var result []map[string]string
		json.Unmarshal([]byte(body), &result)
		for _, e := range result {
			users[e["id"]] = e["name"]
		}
	}
	{
		// メッセージ一覧からユーザーを追加
		resp, _ := http.Get("https://versatileapi.herokuapp.com/api/text/all")
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		io.ReadAll(resp.Body)
		var result []map[string]string
		json.Unmarshal([]byte(body), &result)
		n := 0
		for _, e := range result {
			if _, ok := users[e["_user_id"]]; !ok {
				users[e["_user_id"]] = "匿名" + strconv.Itoa(n)
				n++
			}
		}
	}
	file, err := os.OpenFile("user-datas.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		color.Red("キャッシュファイルの作成に失敗!")
		os.Exit(1)
	}

	usersJsonData, _ := json.Marshal(users)
	file.Write(usersJsonData)
	defer file.Close()
	return users
}

func ReadUsers() map[string]string {
	file, err := os.Open("user-datas.json")
	if err != nil {
		fmt.Println("ユーザーデータが存在しないので、作成します")
		updateUsers()
	}
	content, _ := ioutil.ReadAll(file)
	defer file.Close()
	var result map[string]string
	json.Unmarshal([]byte(content), &result)
	return result
}

func printAllUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "すべてのユーザーを表示します",
		Run: func(cmd *cobra.Command, args []string) {
			// ユーザー一覧を取得
			resp, _ := http.Get("https://versatileapi.herokuapp.com/api/user/all")
			var result []map[string]string
			body, _ := io.ReadAll(resp.Body)
			json.Unmarshal([]byte(body), &result)
			for _, e := range result {
				fmt.Printf("[%s]%s:%s\n", color.GreenString(e["id"]), color.BlueString(e["name"]), e["description"])
			}
		},
	}
	return cmd
}
func updateUsersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "ユーザーデータのキャッシュを更新します",
		Run: func(cmd *cobra.Command, args []string) {
			updateUsers()
		},
	}
	return cmd
}
