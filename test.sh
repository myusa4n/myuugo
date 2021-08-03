#!/bin/bash

assert() {
  expected="$1"
  input="$2"

  ./main "$input" > tmp.s
  gcc -o tmp tmp.s
  ./tmp
  actual="$?"

  if [ "$actual" = "$expected" ]; then
    echo "$input => $actual"
  else
    echo "$input => $expected expected, but got $actual"
    exit 1
  fi
}

assert 21 "
func main() {
  5+20-4
}"
assert 41 "
func main() {
  12 + 34 - 5
}"
assert 47 '
func main() {
  5 + 6*7
}'
assert 15 '
func main() {
  5*(9-6)
}'
assert 4 '
func main() {
  (3+5) / 2
}'
assert 5 '
func main() {
  10 + -5
}'
assert 4 '
func main() {
  -10 + -7 * -2
}'
assert 1 '
func main() {
  1 + 1 == 2
}'
assert 0 '
func main() {
  1 - 5 * 2 == 9
}
'
assert 0 '
func main() {
  4 * -3 != -12
}'
assert 1 '
func main() {
  1 - 5 * 2 != 9
}'
assert 1 '
func main() {
  5 * 10 * -1 < 7 * -7
}'
assert 0 '
func main() {
  2+3 < 5
}'
assert 1 '
func main() {
  2+3<=5
}'
assert 1 '
func main() {
  4*3<=5*7
}'
assert 0 '
func main() {
  5 * 10 * -1 > 7 * -7
}'
assert 0 '
func main() {
  2+3 > 5
}'
assert 1 '
func main() {
  2+3>=5
}'
assert 0 '
func main() {
  4*3>=5*7
}'
assert 4 "
func main() {
  a = 3
  a + 1
}
"
assert 7 "func main() { z = 20; a = 13; x = z - a; x }"
assert 21 "
func main() {
  a = 5
  a + 3;
  4 * a+1
}"
assert 222 "
func main() {
  hello = 5 * 4 + 2
  world = hello * 20 / 2
  world + 2
}
"
assert 5 "
func main() {
  return 5
}"
assert 10 "
func main() {
  abc = 2
  return 5*abc
}
"
assert 1 "
func main() {
  a = 2
  if a == 2 {
    a = a * 3
    a = 1
    return a
    a = 5
  }
  a
}
"
assert 2 "
func main() {
  a = 2
  if a != 2 {
    a = 6
  }
  a
}
"
assert 34 "
func main() {
  test = 16
  if test < 10 {
    test = 100
    test = test + 21
  } else {
    test = test - 5
    test = test - 1
    test = 3 * test + 4
  }
  test
}"
assert 15 "
func main() {
  i = 1
  sum = 0
  for {
    sum = sum + i
    if i == 5 {
      return sum
    }
    i = i + 1
  }
}
"
assert 15 "
func main() {
  i = 1
  sum = 0
  for i < 6 {
    sum = sum + i
    i = i + 1
  }
  sum
}
"
assert 15 "
func main() {
  sum = 0
  for i = 0; i < 6; i = i+1 {
    sum = sum + i
  }
  sum
}
"
assert 3 "
func foo(a, b) {
  return a + b
}

func main() {
  foo(2, 1)
}
"
assert 8 "
func fib(n) {
  if n <= 1 {
    return 1
  }
  return fib(n - 1) + fib(n - 2)
}

func main() {
  return fib(5)
}
"
assert 4 "func main() {
  str = 4
  *&str
}
"
assert 10 "func main() {
  a = 10
  b = 4
  *(&b + 8)
}"

echo OK