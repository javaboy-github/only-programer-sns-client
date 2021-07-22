package msg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/javaboy-github/only-programer-sns-client/util"
	"github.com/javaboy-github/only-programer-sns-client/user"
	"github.com/spf13/cobra"
)

func MsgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "msg",
		Short: "メッセージを送信/閲覧/返信します。",
	}
	cmd.AddCommand(sendMsg())
	cmd.AddCommand(seeMsgsCmd())
	return cmd
}

func sendMsg() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send msg",
		Short: "メセージを送信します。",
		Long:  "メッセージを送信します。第一引数にreply先のIDを入れることも可能です",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("メッセージを入力[二回連続で改行で送信][280B]")
			// テキストを取得
			msg := util.GetText()
			msg = util.StringToJsonString(msg)

			// 280字以内でないと送信できない
			if utf8.RuneCountInString(msg) > 280 {
				color.Red("280Bを超えています！送信不可")
				os.Exit(1)
			} else {
				fmt.Println("送信します")
				// メッセージを送信
				var data string
				if len(args) == 0 {
					data = fmt.Sprintf("{\"text\":\"%s\"}", msg)
				} else {
					data = fmt.Sprintf("{\"text\":\"%s\",\"in_reply_to_text_id\":\"%s\"}", msg, args[0])
				}
				req, _ := http.NewRequest(http.MethodPost, "https://versatileapi.herokuapp.com/api/text", bytes.NewBuffer([]byte(data)))
				req.Header.Set("Authorization", "HelloWorld")
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					color.Red("エラーが発生しました")
				} else {
					fmt.Println("送信が完了しました！")
				}
				defer resp.Body.Close()
			}
		},
	}
	return cmd
}

func seeMsgsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "すべてのメッセージを表示します",
		Run: func(cmd *cobra.Command, args []string) {
			// ユーザーのリストを取得
			userList := user.ReadUsers()

			// メッセージ一覧を取得
			resp, _ := http.Get("https://versatileapi.herokuapp.com/api/text/all")
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var result []map[string]string
			n := 0 // 匿名ナンバー
			json.Unmarshal([]byte(body), &result)
			texts := map[string][]string{} // リプライ実装用
			for _, e := range result {
				if _, ok := userList[e["_user_id"]]; !ok {
					userList[e["_user_id"]] = "匿名" + strconv.Itoa(n)
					n++
				}
				texts[e["id"]] = []string{e["_user_id"], e["text"]}
				// リプライに関する文字列
				reply := ""
				replayToText, ok1 := e["in_reply_to_text_id"]
				replayToUser, ok2 := e["in_reply_to_user_id"]
				if ok1 && ok2 {
					// ユーザーとテキストにリプライ
					// リプライ先のユーザー名とリプライ先のテキストのユーザー名が同じ場合
					if texts[replayToText][0] == replayToUser {
						reply = fmt.Sprintf("%s %s>", color.BlueString("@"+userList[replayToUser]), texts[replayToText][1])
					} else {
						// ただの地獄
						reply = fmt.Sprintf("%s,%s %s>", color.BlueString("@"+userList[replayToUser]), color.BlueString("@"+userList[texts[replayToText][0]]), texts[replayToText][1])
					}
				} else if ok2 {
					// ユーザーのみにリプライ
					reply = fmt.Sprintf("%s>", color.BlueString("@"+userList[replayToUser]))
				} else if ok1 {
					// テキストのみにリプライ
					reply = fmt.Sprintf("%s %s>", color.BlueString("@"+userList[texts[replayToText][0]]), texts[replayToText][1])
				}
				fmt.Printf("%s[%s][%s] %s\n", color.BlueString(userList[e["_user_id"]]), color.GreenString(e["id"]), color.YellowString(e["_created_at"]), reply)
				fmt.Println(e["text"])
			}
		},
	}
	return cmd
}
