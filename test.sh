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

assert 0 "
package main

func main() {
  22
}
"
assert 21 "
package main

func main() {
  return 5+20-4
}"
assert 41 "
package main

func main() {
  return 12 + 34 - 5
}"
assert 47 '
package main

func main() {
  return 5 + 6*7
}'
assert 15 '
package main

func main() {
  return 5*(9-6)
}'
assert 4 '
package main

func main() {
  return (3+5) / 2
}'
assert 5 '
package main

func main() {
  return 10 + -5
}'
assert 4 '
package main

func main() {
  return -10 + -7 * -2
}'
assert 1 '
package main

func main() {
  return 1 + 1 == 2
}'
assert 0 '
package main

func main() {
  return 1 - 5 * 2 == 9
}
'
assert 0 '
package main

func main() {
  return 4 * -3 != -12
}'
assert 1 '
package main

func main() {
  return 1 - 5 * 2 != 9
}'
assert 1 '
package main

func main() {
  return 5 * 10 * -1 < 7 * -7
}'
assert 0 '
package main

func main() {
  return 2+3 < 5
}'
assert 1 '
package main

func main() {
  return 2+3<=5
}'
assert 1 '
package main

func main() {
  return 4*3<=5*7
}'
assert 0 '
package main

func main() {
  return 5 * 10 * -1 > 7 * -7
}'
assert 0 '
package main

func main() {
  return 2+3 > 5
}'
assert 1 '
package main

func main() {
  return 2+3>=5
}'
assert 0 '
package main

func main() {
  return 4*3>=5*7
}'
assert 4 "
package main

func main() {
  var a = 3
  return a + 1
}
"
assert 7 "package main; func main() { var z = 20; var a = 13; var x = z - a; return x }"
assert 21 "
package main

func main() {
  var a = 5
  a + 3;
  return 4 * a+1
}"
assert 222 "
package main

func main() {
  var hello = 5 * 4 + 2
  var world = hello * 20 / 2
  return world + 2
}
"
assert 5 "
package main

func main() {
  return 5
}"
assert 10 "
package main

func main() {
  var abc = 2
  return 5*abc
}
"
assert 1 "
package main

func main() {
  var a = 2
  if a == 2 {
    a = a * 3
    a = 1
    return a
    a = 5
  }
  return a
}
"

assert 2 "
package main

func main() {
  var a = 2
  if a != 2 {
    a = 6
  }
  return a
}
"
assert 34 "
package main

func main() {
  var test = 16
  if test < 10 {
    test = 100
    test = test + 21
  } else {
    test = test - 5
    test = test - 1
    test = 3 * test + 4
  }
  return test
}"
assert 15 "
package main

func main() {
  var i = 1
  var sum = 0
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
package main

func main() {
  var i = 1
  var sum = 0
  for i < 6 {
    sum = sum + i
    i = i + 1
  }
  return sum
}
"
assert 15 "
package main

func main() {
  var sum = 0
  var i int
  for i = 0; i < 6; i = i+1 {
    sum = sum + i
  }
  return sum
}
"
assert 3 "
package main

func foo(a int, b int) int {
  return a + b
}

func main() {
  return foo(2, 1)
}
"
assert 8 "
package main

func fib(n int) int {
  if n <= 1 {
    return 1
  }
  return fib(n - 1) + fib(n - 2)
}

func main() {
  return fib(5)
}
"
assert 4 "package main
func main() {

  var str = 4
  return *&str
}
"
assert 10 "
package main

func main() {
  var a = 10
  var b = 4
  return *(&b + 8)
}"
assert 22 "
package main
func main() {
  var a int = 3
  var b int
  b = 19
  return a + b
}"
assert 3 "
package main

func main() {
  var x int
  var y *int

  y = &x
  *y = 3
  return x
}"
assert 12 "
package main

func main() {
  var x1 int = 1
  var x11 int = 11
  return x1 + x11
}"

echo OK