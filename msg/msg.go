package msg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/javaboy-github/only-programer-sns-client/user"
	"github.com/javaboy-github/only-programer-sns-client/util"
	"github.com/spf13/cobra"
)

func MsgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "msg",
		Short: "メッセージを送信/閲覧/返信します。",
	}
	cmd.AddCommand(sendMsgCmd())
	cmd.AddCommand(seeMsgsCmd())
	cmd.AddCommand(replyCmd())
	return cmd
}

// メッセージを送信する
// @param text 送信するコンテンツ
// @param replayToText リプライ先のテキストID。ない場合はnull文字列
// @param replayToUser リプライ先のユーザーID。ない場合はnull文字列
func sendMsg(text string, replyToText string, replyToUser string) {
	data := map[string]string{"text": text}
	if replyToText != "" {
		data["in_reply_to_text_id"] = replyToText
	}
	if replyToUser != "" {
		data["in_reply_to_user_id"] = replyToUser
	}
	// data := fmt.Sprintf("{\"text\":\"%s\"}", text)
	jsonData, _        := json.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, "https://versatileapi.herokuapp.com/api/text", bytes.NewBuffer([]byte(jsonData)))
	req.Header.Set("Authorization", "HelloWorld")
	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
}
func sendMsgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send msg",
		Short: "メセージを送信します。",
		Long:  "メッセージを送信します。",
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
				sendMsg(msg, "", "")
				fmt.Println("送信が完了しました！")
			}
		},
	}
	return cmd
}

func seeMsgsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "すべてのメッセージを表示し、キャッシュを作成します",
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
			// キャッシュの更新
			file, err := os.OpenFile("text-datas.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
			if err != nil {
				color.Red("キャッシュの更新に失敗!")
				fmt.Println(err)
				os.Exit(1)
			}
			cacheData := []string{}
			for _, e := range result {
				cacheData = append(cacheData, e["id"])
			}
			jsonCacheData  , _ := json.Marshal(cacheData)
			file.Write(jsonCacheData)
			defer file.Close()

			// テキストIDを追加
			for i, e := range result {
				e["text_id"] = fmt.Sprint(i)
			}
			// 降順にする
			for i := 0; i < len(result)/2; i++ {
				result[i], result[len(result)-i-1] = result[len(result)-i-1], result[i]
			}
			// ページとリミット分だけ表示する
			limit, err1 := cmd.Flags().GetInt("limit")
			page, err2 := cmd.Flags().GetInt("page")
			if err1 != nil {
				color.Red("フラグlimitでエラー!")
			}
			if err2 != nil {
				color.Red("フラグpageでエラー!")
			}
			if err1 != nil || err2 != nil {
				os.Exit(1)
			}
			result = result[limit*page : limit*page+limit]
			// 順番を戻す
			for   i := 0; i < len(result)/2; i++ {
				result[i], result[len(result)-i-1] = result[len(result)-i-1], result[i]
			}

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
					val, err := texts[replayToText]
					if !err {
						reply = "不明>"
					} else if val[0] == replayToUser {
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
				fmt.Printf("%s[#%s][%s] %s\n", color.BlueString(userList[e["_user_id"]]), color.GreenString(e["text_id"]), color.YellowString(e["_created_at"]), reply)
				fmt.Println(e["text"])
			}
		},
	}
	cmd.Flags().IntP("limit", "l", 20, "人ページに表示する数")
	cmd.Flags().IntP("page", "p", 0, "表示するページ")
	return cmd
}

func replyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reply",
		Short: "リプライします。",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			text, _ := cmd.Flags().GetString("text")
			if name != "" && len(name) != 40 {
				// 名前がIDでないとき
				// 名前を適切に入れる
				users := user.ReadUsers()
				for id, user := range users {
					if user == name {
						name = id
						break
					}
				}
				// IDが入らなかった場合
				if len(name) != 40 {
					color.Red("ユーザーを見つけられません！")
					os.Exit(0)
				}
			}
			if text != "" && len(text) != 36 {
				// テキストがIDでないとき
				i, err := strconv.Atoi(text)
				if err != nil {
					color.Red("フラグtextを整数に変換できません！")
					os.Exit(1)
				}
				text = ReadTexts()[i]
				// IDが入らなかった場合
				fmt.Println(text)
				if len(text) != 36 {
					color.Red("テキストを見つけられません!")
					os.Exit(1)
				}
			}

			cmd.Println("メッセージを入力[二回連続で改行で送信][280B]")
			// テキストを取得
			msg := util.GetText()
			msg = util.StringToJsonString(msg)

			// 280字以内でないと送信できない
			if utf8.RuneCountInString(msg) > 280 {
				color.Red("280Bを超えています！送 信不可")
				os.Exit(1)
			} else {
				fmt.Println("送信します")
				// メッセージを送信
				sendMsg(msg, text, name)
				fmt.Println("送信が完了しました！")
			}
		},
	}
	cmd.Flags().StringP("name", "n", "", "リプレイ先の名前。idでも、ユーザー名でも可")
	cmd.Flags().StringP("text", "t", "", "リプレイ先のテキスト。idでも、テキストIDでも可")
	return cmd
}

func ReadTexts() []string {
	file, err := os.Open("text-datas.json")
	if err != nil {
		fmt.Println("テキストデータが存在しません。msg listで更新してください。")
		os.Exit(0)
	}
	content, _ := ioutil.ReadAll(file)
	defer file.Close()
	var result []string
	json.Unmarshal([]byte(content), &result)
	return result
}
