package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
)

// Enumerate the paths of Go files directly under `path`.
func EnumerateGoFilePaths(path string) []string {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	files, err := os.ReadDir(path)
	if err != nil {
		Alarm("Failed to enumerate Go files.")
	}
	paths := []string{}
	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, ".go") {
			paths = append(paths, path+file.Name())
		}
	}
	return paths
}

// ファイルの末尾に改行を付与して読み込む
func ReadFile(path string) string {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(path)
		Alarm("ファイル '" + path + "' の読み取りに失敗しました")
	}
	if len(bytes) == 0 || bytes[len(bytes)-1] != '\n' {
		bytes = append(bytes, '\n')
	}
	return string(bytes)
}

// エラーの起きた場所を報告するための関数
// 下のようなフォーマットでエラーメッセージを表示する
//
// foo.c:10: x = y + + 5;
//                   ^ 式ではありません
func ErrorAt(path string, target string, message string) {
	var content = ReadFile(path)
	// 行番号と、restがその行の何番目から始まるかを見つける
	var lineNumber = 1
	var startIndex = 0
	for _, c := range content[:len(content)-len(target)] {
		if c == '\n' {
			lineNumber += 1
			startIndex = 0
		} else if c == '\t' {
			// タブは空白8文字だと考える
			startIndex = ((startIndex + 8) / 8) * 8
		} else {
			startIndex += 1
		}
	}
	for i, line := range strings.Split(content, "\n") {
		if i+1 == lineNumber {
			// 見つかった行をファイル名と行番号と一緒に表示
			var indent, _ = fmt.Fprintf(os.Stderr, "%s:%d: ", path, lineNumber)
			fmt.Fprintln(os.Stderr, line)
			fmt.Fprintf(os.Stderr, "%*s^ %s\n", indent+startIndex, " ", message)
		}
	}
	os.Exit(1)
}

func Strtoi(s string) (int, string) {
	var res = 0
	for i, c := range s {
		if !unicode.IsDigit(c) {
			return res, s[i:]
		}
		res = res*10 + int(c) - int('0')
	}
	return res, ""
}

func IsAlnum(c rune) bool {
	return IsAlpha(c) || unicode.IsDigit(c)
}

func IsAlpha(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func RuneAt(s string, i int) rune {
	return []rune(s)[i]
}

func Alarm(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}
