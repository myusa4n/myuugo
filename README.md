# myuugo

## 概要
植山類さんの [低レイヤを知りたい人のためのCコンパイラ作成入門](https://www.sigbus.info/compilerbook) を参考に実装しているGo言語のコンパイラです。

当面の目的はセルフホストできるようにすることです。

## 現状(2021/08/04更新)
```go
func fib(n) {
  if n <= 1 {
    return 1
  }
  return fib(n - 1) + fib(n - 2)
}

func main() {
  var ans = 0
  var n int
  for n = 1; n <= 10; n = n + 1 {
    ans = ans + fib(n)
  }
  return ans
}
```