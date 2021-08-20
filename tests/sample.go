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
	memo[n] = fib(n-1) + fib(n-2)
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

type FibonacciNumber struct {
	Nth   int
	Value int
}

/*
  1番目から10番目までの
  フィボナッチ数の総和を計算するプログラム
*/
func main() {
	for n := 0; n <= 10; n = n + 1 {
		setMinusOne(&memo[n]) // 初期化
	}

	var fibs []FibonacciNumber = []FibonacciNumber{}
	for n := 1; n <= 10; n = n + 1 {
		fibs = append(fibs, FibonacciNumber{Nth: n, Value: fib(n)})
		printf("%d: %d\n", n, fibs[n-1].Value)
	}

	return
}
