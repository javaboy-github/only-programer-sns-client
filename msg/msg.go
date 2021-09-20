package msg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
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
	jsonData, _ := json.Marshal(data)
	req, err := http.NewRequest(http.MethodPost, "https://versatileapi.herokuapp.com/api/text", bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Authorization", "HelloWorld")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
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

// キャッシュを更新する。
// @param newData 新しいデータ。このデータを元に更新する。また、データが足りない場合は重い処理になる。
func updateMsgs(newData []map[string]string) (int, []map[string]string) {
	fmt.Println("キャッシュを更新します...")
	// IDを取り出す
	ids := []string{}
	for _, e := range newData {
		ids = append(ids, e["id"])
	}

	// キャッシュを読み込む
	{
		cacheFile, err := os.OpenFile("text-datas.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			// キャッシュがない場合->更新
			goto updateAllData
		}

		cache := []map[string]string{}
		content, _ := ioutil.ReadAll(cacheFile)
		err = json.Unmarshal([]byte(content), &cache)
		if err != nil {
			log.Fatal(err)
		}
		// キャッシュ内を検索し、補完できる場合は補完する。
		for i, e := range cache {
			if e["id"] == newData[0]["id"] {
				// 補完できる
				// 補完
				for j, v := range newData {
					v["number"] = strconv.Itoa(i + j)
					cache[i+j] = v
				}
				// 保存
				jsonCacheData, _ := json.Marshal(cache)
				cacheFile.Write([]byte(jsonCacheData))
				defer cacheFile.Close()
				return i, cache
			}
		}
		defer cacheFile.Close()
		// 補完できない場合はそのまま下記が実行される
	}

updateAllData:
	// データをすべて更新
	resp, _ := http.Get("https://versatileapi.herokuapp.com/api/text/all")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result []map[string]string
	json.Unmarshal([]byte(body), &result)
	// キャッシュの更新
	file, err := os.OpenFile("text-datas.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		color.Red("キャッシュの更新に失敗!")
		fmt.Println(err)
		os.Exit(1)
	}
	for i, e := range result {
		e["number"] = strconv.Itoa(i)
	}
	jsonCacheData, _ := json.Marshal(result)
	file.Write(jsonCacheData)
	defer file.Close()
	for i, e := range result {
		if e["id"] == newData[0]["id"] {
			return i, result
		}
	}
	return -1, nil
}

func seeMsgsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "すべてのメッセージを表示し、キャッシュを作成します",
		Run: func(cmd *cobra.Command, args []string) {
			// フラグを取得
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
			// ユーザーのリストを取得
			userList := user.ReadUsers()

			// メッセージ一覧を取得
			resp, err := http.Get(fmt.Sprintf("https://versatileapi.herokuapp.com/api/text/all?$orderby=_created_at+desc&$skip=%d&$limit=%d", page*limit, limit))
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var result []map[string]string
			n := 0 // 匿名ナンバー
			json.Unmarshal([]byte(body), &result)

			// 順番を戻す
			for i := 0; i < len(result)/2; i++ {
				result[i], result[len(result)-i-1] = result[len(result)-i-1], result[i]
			}
			// キャッシュの更新
			index, texts := updateMsgs(result)
			getElement := func(id string) map[string]string {
				for _, e := range texts {
					if e["id"] == id {
						return e
					}
				}
				return nil
			}
			if index == -1 {
				color.Red("キャッシュの更新でエラー!")
			}

			// テキストIDを追加
			for i, e := range result {
				e["text_id"] = fmt.Sprint(index + i)
			}
			for _, e := range result {
				if _, ok := userList[e["_user_id"]]; !ok {
					userList[e["_user_id"]] = "不明" + strconv.Itoa(n)
					n++
				}
				// リプライに関する文字列
				reply := ""
				replayToText, ok1 := e["in_reply_to_text_id"]
				replayToUser, ok2 := e["in_reply_to_user_id"]
				if ok1 && ok2 {
					// ユーザーとテキストにリプライ
					// リプライ先のユーザー名とリプライ先のテキストのユーザー名が同じ場合
					val := getElement(replayToText)
					if val == nil {
						reply = "不明>"
					} else if val["_user_id"] == replayToUser {
						reply = fmt.Sprintf("%s %s>", color.BlueString("@"+userList[replayToUser]), getElement(replayToText)["text"])
					} else {
						// ただの地獄
						reply = fmt.Sprintf("%s,%s %s>", color.BlueString("@"+userList[replayToUser]), color.BlueString("@"+userList[getElement(replayToText)["name"]]), getElement(replayToText)["text"])
					}
				} else if ok2 {
					// ユーザーのみにリプライ
					reply = fmt.Sprintf("%s>", color.BlueString("@"+userList[replayToUser]))
				} else if ok1 {
					// テキストのみにリプライ
					val := getElement(replayToText)
					if val == nil {
						reply = "不明 不明>"
					} else {
						reply = fmt.Sprintf("%s %s>", color.BlueString("@"+userList[val["name"]]), val["text"])
					}
				}
				date, _ := time.Parse("2006-01-02T15:04:05.000+00:00", e["_created_at"])
				fmt.Printf("%s[#%s][%s] %s\n", color.BlueString(userList[e["_user_id"]]), color.GreenString(e["text_id"]), color.YellowString(date.In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format("2006-01-02 15:04:05")), reply)
				fmt.Println(e["text"])
			}
		},
	}
	cmd.Flags().IntP("limit", "l", 20, "1ページに表示する数")
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
				text = ReadTexts()[i]["id"]
				// IDが入らなかった場合
				if len(text) != 36 {
					color.Red("テキストを見つけられません!")
					os.Exit(1)
				}
			}

			cmd.Println("メッセージを入力[二回連続で改行で送信][280B]")
			// テキストを取得
			msg := util.GetText()

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

func ReadTexts() []map[string]string {
	file, err := os.Open("text-datas.json")
	if err != nil {
		fmt.Println("テキストデータが存在しません。msg listで更新してください。")
		os.Exit(0)
	}
	content, _ := ioutil.ReadAll(file)
	defer file.Close()
	var result []map[string]string
	json.Unmarshal([]byte(content), &result)
	return result
}
