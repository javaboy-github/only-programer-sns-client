package msg

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"unicode/utf8"

	"github.com/fatih/color"
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
		Run: func(cmd *cobra.Command, args []string) {
			var msg string
			cmd.Println("メッセージを入力[二回連続で改行で送信][280B]")
			var sc = bufio.NewScanner(os.Stdin)
			// 二回連続で改行が入るまで繰り返す
			for {
				sc.Scan()
				input := sc.Text()
				if input == "" {
					break
				}
				msg += input
			}

			// 280字以内でないと送信できない
			if utf8.RuneCountInString(msg) > 280 {
				color.Red("280Bを超えています！送信不可")
				os.Exit(1)
			} else {
				fmt.Println("送信します")
				// メッセージを送信
                req, _ := http.NewRequest(http.MethodPost, "https://versatileapi.herokuapp.com/api/text", bytes.NewBuffer([]byte("{\"text\":\"" + msg +"\"}")))
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

func seeMsgsCmd()*cobra.Command  {
	cmd := &cobra.Command{
		Use: "list",
		Short: "すべてのメッセージを表示します",
		Run: func(cmd *cobra.Command, args []string) {
			// メッセージ一覧を取得
			resp, _ := http.Get("https://versatileapi.herokuapp.com/api/text/all")
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var result []map[string]string;
			json.Unmarshal([]byte(body), &result)
			for _,e := range result {
				fmt.Println(color.BlueString(e["id"]) + "[" + color.YellowString(e["_created_at"]) + "]")
				fmt.Println(e["text"])
			}
		},
	}
	return cmd
}
