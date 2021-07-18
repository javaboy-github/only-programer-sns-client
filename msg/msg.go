package msg

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
    "net/http"
    "bytes"
)

func MsgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "msg",
		Short: "メッセージを送信/閲覧/返信します。",
	}
	cmd.AddCommand(sendMsg())
	return cmd
}

func sendMsg() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send msg",
		Short: "メセージを送信します。",
		Run: func(cmd *cobra.Command, args []string) {
			var msg string
			cmd.Println("メッセージを入力[二回連続で改行で送信][280B]")
			var isEntered bool
			var sc = bufio.NewScanner(os.Stdin)
			for {
				sc.Scan()
				input := sc.Text()
				if isEntered && input == "" {
					break
				}
				isEntered = input == ""
				msg += input
			}

			if len(msg) > 280 {
				color.Red("280Bを超えています！送信不可")
				os.Exit(1)
			} else {
				fmt.Println("送信します")
                req, _ := http.NewRequest(http.MethodPost, "https://versatileapi.herokuapp.com/api/text", bytes.NewBuffer([]byte("{text:\"" + msg +"\"}")))
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
