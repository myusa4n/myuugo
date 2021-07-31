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

assert 0 0
assert 42 42
assert 21 "5+20-4"
assert 41 "12 + 34 - 5"
assert 47 '5 + 6*7'
assert 15 '5*(9-6)'
assert 4 '(3+5) / 2'
assert 5 '10 + -5'
assert 4 '-10 + -7 * -2'
assert 1 '1 + 1 == 2'
assert 0 '1 - 5 * 2 == 9'
assert 0 '4 * -3 != -12'
assert 1 '1 - 5 * 2 != 9'
assert 1 '5 * 10 * -1 < 7 * -7'
assert 0 '2+3 < 5'
assert 1 '2+3<=5'
assert 1 '4*3<=5*7'
assert 0 '5 * 10 * -1 > 7 * -7'
assert 0 '2+3 > 5'
assert 1 '2+3>=5'
assert 0 '4*3>=5*7'
assert 4 "
a = 3
a + 1
"
assert 7 "z = 20; a = 13; x = z - a; x"
assert 21 "a = 5
a + 3;
4 * a+1
"
assert 222 "hello = 5 * 4 + 2
world = hello * 20 / 2
world + 2
"
assert 5 "return 5"
assert 10 "abc = 2
return 5*abc
"
assert 1 "a = 2
if a == 2 {
  a = a * 3
  a = 1
  return a
  a = 5
}
a"
assert 2 "a = 2
if a != 2 {
  a = 6
}
a"
assert 34 "test = 16
if test < 10 {
  test = 100
  test = test + 21
} else {
  test = test - 5
  test = test - 1
  test = 3 * test + 4
}
test"
assert 15 "i = 1
sum = 0
for {
  sum = sum + i
  if i == 5 {
    return sum
  }
  i = i + 1
}
"
assert 15 "i = 1
sum = 0
for i < 6 {
  sum = sum + i
  i = i + 1
}
sum
"

echo OK