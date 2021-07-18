package main

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command {
	Use: "SNSを使う",
	Short: "Programer only SNS のクライアント",
	Run: func(cmd *cobra.Command, args [] string) {
		fmt.Println("Root command")
	},
}

func main() {
    if err := RootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
        os.Exit(-1)
    }
}
