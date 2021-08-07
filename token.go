package main

import (
	"strconv"
	"unicode"
)

// (先頭の識別子, 識別子を切り出して得られた残りの文字列)  を返す
func getIdentifier(s string) (string, string) {
	var res = ""
	for i, c := range s {
		if (i == 0 && unicode.IsDigit(c)) || !(isAlnum(c) || (c == '_')) {
			return res, s[i:]
		}
		res += string(c)
	}
	return res, ""
}

type TokenKind string

const (
	TokenReserved   TokenKind = "RESERVED"
	TokenNumber     TokenKind = "NUMBER"
	TokenIdentifier TokenKind = "IDENTIFIER"
	TokenEof        TokenKind = "EOF"
	TokenReturn     TokenKind = "return"
	TokenIf         TokenKind = "if"
	TokenElse       TokenKind = "else"
	TokenFor        TokenKind = "for"
	TokenFunc       TokenKind = "func"
	TokenVar        TokenKind = "var"
	TokenPackage    TokenKind = "package"
)

type Token struct {
	kind TokenKind // トークンの型
	val  int       // kindがNumberの場合、その数値
	str  string    // トークン文字列
	rest string    // 自信を含めた残りすべてのトークン文字列
}

func (t Token) Test(kind TokenKind, str string) bool {
	return t.kind == kind && t.str == str
}

func NewToken(kind TokenKind, str string, rest string) Token {
	return Token{kind: kind, str: str, rest: rest}
}

type Tokenizer struct {
	tokens []Token // inputをトークナイズした結果
	pos    int     // 現在着目しているトークンの添数
	input  string  // ユーザーからの入力プログラム
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{}
}

func (t *Tokenizer) Tokenize(input string) {
	t.tokens = []Token{}
	t.pos = 0
	t.input = input

	/*
		var keywords = []string{
			"package",
			"return",
			"func", "else",
			"for", "var",
			"if", "==", "!=", "<=", ">=",
			"+", "-", "*", "/", "(", ")", "<", ">", ";", "\n", "=", "{", "}", ",", "&",
		}
	*/

	for input != "" {
		if len(input) >= 2 {
			var head2 = input[:2]
			if head2 == "==" || head2 == "!=" || head2 == "<=" || head2 == ">=" {
				t.tokens = append(t.tokens, NewToken(TokenReserved, head2, input))
				input = input[2:]
				continue
			}
		}

		var c = runeAt(input, 0)
		if isAlpha(c) || (c == '_') {
			// input から 識別子を取り出す
			var token = NewToken(TokenIdentifier, "", input)
			token.str, input = getIdentifier(input)
			if token.str == string(TokenReturn) {
				token.kind = TokenReturn
			} else if token.str == string(TokenIf) {
				token.kind = TokenIf
			} else if token.str == string(TokenElse) {
				token.kind = TokenElse
			} else if token.str == string(TokenFor) {
				token.kind = TokenFor
			} else if token.str == string(TokenFunc) {
				token.kind = TokenFunc
			} else if token.str == string(TokenVar) {
				token.kind = TokenVar
			} else if token.str == string(TokenPackage) {
				token.kind = TokenPackage
			}
			t.tokens = append(t.tokens, token)
			continue
		}
		if c == '+' || c == '-' || c == '*' || c == '/' || c == '(' || c == ')' || c == '<' ||
			c == '>' || c == ';' || c == '\n' || c == '=' || c == '{' || c == '}' || c == ',' || c == '&' {
			t.tokens = append(t.tokens, NewToken(TokenReserved, string(c), input))
			input = input[1:]
			continue
		}
		if unicode.IsSpace(c) {
			input = input[1:]
			continue
		}
		if unicode.IsDigit(c) {
			var token = NewToken(TokenNumber, "", input)
			token.val, input = strtoi(input)
			token.str = strconv.Itoa(token.val)
			t.tokens = append(t.tokens, token)
			continue
		}
		errorAt(input, "トークナイズできません")
	}
	t.tokens = append(t.tokens, NewToken(TokenEof, "", ""))
}

// 現在のトークンを返す
func (t *Tokenizer) Fetch() Token {
	return t.tokens[t.pos]
}

func (t *Tokenizer) Prefetch(n int) Token {
	return t.tokens[t.pos+n]
}

func (t *Tokenizer) Succ() {
	t.pos = t.pos + 1
}
