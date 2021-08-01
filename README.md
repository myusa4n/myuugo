# myuugo

## 概要
植山類さんの [低レイヤを知りたい人のためのCコンパイラ作成入門](https://www.sigbus.info/compilerbook) を参考に実装しているGo言語のコンパイラです。

当面の目的はセルフホストできるようにすることです。

## 現状(2021/08/01更新)
```go
func fib(n) {
  if n <= 1 {
    return 1
  }
  return fib(n - 1) + fib(n - 2)
}

func main() {
  ans = 0
  for n = 1; n <= 10; n++ {
    ans = ans + fib(n)
  }
  return ans
}
```