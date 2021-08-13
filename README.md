# myuugo

## 概要
植山類さんの [低レイヤを知りたい人のためのCコンパイラ作成入門](https://www.sigbus.info/compilerbook) を参考に実装しているGo言語のコンパイラです。

当面の目的はセルフホストできるようにすることです。

## 現状(2021/08/10更新)
```go
package main

var memo [11]int

// フィボナッチ数をメモ化再帰で計算する
func fib(n int) int {
  if memo[n] != -1 {
    return memo[n]
  }
  if n <= 1 {
    memo[n] = 1
    return memo[n]
  }
  memo[n] = fib(n - 1) + fib(n - 2)
  return memo[n]
}

// フィボナッチ数をO(n)で計算する
func fib2(n int) (int, int) {
  if n == 1 {
    return 1, 1
  }
  cur, prev := fib2(n - 1)
  return cur + prev, cur
}

func setMinusOne(n *int) {
  *n = -1
}

/*
  1番目から10番目までの
  フィボナッチ数の総和を計算するプログラム
*/
func main() {
  for n := 0; n <= 10; n = n + 1 {
    setMinusOne(&memo[n]) // 初期化
  }

  var ans = 0
  for n = 1; n <= 10; n = n + 1 {
    ans = ans + fib(n)
  }
  printf("ans is %d\n", ans)
  return
}
```