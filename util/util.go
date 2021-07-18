package util

import (
	"bufio"
	"os"
)

// 改行が入れられるまで文字を受け付ける
func GetText() string {
	var result string
	var sc = bufio.NewScanner(os.Stdin)
	for {
		sc.Scan()
		input := sc.Text()
		if input == "" {
			break
		}
		result += input
	}
	return result
}
