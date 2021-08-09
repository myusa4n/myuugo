# myuugo

## 概要
植山類さんの [低レイヤを知りたい人のためのCコンパイラ作成入門](https://www.sigbus.info/compilerbook) を参考に実装しているGo言語のコンパイラです。

当面の目的はセルフホストできるようにすることです。

## 現状(2021/08/08更新)
```go
package main

var memo [11]int

// 
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

func setMinusOne(n *int) {
  *n = -1
}

func main() {
  var n int
  for n = 0; n <= 10; n = n + 1 {
    setMinusOne(&memo[n])
  }

  var ans = 0
  for n = 1; n <= 10; n = n + 1 {
    ans = ans + fib(n)
  }
  printf("ans is %d\n", ans)
  return
}
```