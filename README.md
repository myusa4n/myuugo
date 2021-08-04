# myuugo

## 概要
植山類さんの [低レイヤを知りたい人のためのCコンパイラ作成入門](https://www.sigbus.info/compilerbook) を参考に実装しているGo言語のコンパイラです。

当面の目的はセルフホストできるようにすることです。

## 現状(2021/08/05更新)
```go
package main

func fib(n int) int {
  if n <= 1 {
    return 1
  }
  return fib(n - 1) + fib(n - 2)
}

func add10(n int) {
  *n = *n + 10
}

func main() {
  var ans = 0
  var n int

  for n = 1; n <= 10; n = n + 1 {
    ans = ans + fib(n)
  }

  add10(&ans)
  return ans
}
```