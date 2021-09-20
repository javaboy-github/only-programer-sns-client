package main

import (
	"fmt"
	"os"

	"github.com/javaboy-github/only-programer-sns-client/global"
	"github.com/javaboy-github/only-programer-sns-client/msg"
	"github.com/javaboy-github/only-programer-sns-client/user"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   global.AppName,
	Short: global.AppName + ": Programer only SNS のクライアント。",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("使い方が間違ってます。--helpを見てください。")
	},
	Version: global.Version,
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize()
	RootCmd.AddCommand(msg.MsgCmd())
	RootCmd.AddCommand(user.UserCmd())
}
