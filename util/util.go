package util

import (
	"bufio"
	"os"
	"strings"
)

// 改行が入れられるまで文字を受け付ける
func GetText() string {
	var result string
	var sc = bufio.NewScanner(os.Stdin)
	for {
		sc.Scan()
		input := sc.Text()
		if input == "" {
			result = result[:len(result) -1 ]
			break
		}
		result += input + "\n"
	}
	return result
}
// 文字列をjson文字列へエスケープする
func StringToJsonString(str string) string {
	str = strings.Replace(str, "\\", "\\\\", -1)// \　-> \\
	str = strings.Replace(str, "\"", "\\\"", -1)// " -> \"
	str = strings.Replace(str, "/", "\\/", -1)  // / -> \/
	str = strings.Replace(str, "\n", "\\n", -1) // <改行> -> \n
	str = strings.Replace(str, "\t", "\\t", -1) // <タブ> -> \t
	return str
}
