package main

import (
	"fmt"
	"github.com/javaboy-github/only-programer-sns-client/msg"
	"github.com/javaboy-github/only-programer-sns-client/user"
	"github.com/spf13/cobra"
	"os"
)

var RootCmd = &cobra.Command{
	Use:   "opsc",
	Short: "Programer only SNS のクライアント。",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("使い方が間違ってます。--helpを見てください。")
	},
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
