#!/bin/bash

assert() {
  expected="$1"
  input="$2"

  ./main "$input" > tmp.s
  gcc -no-pie -o tmp tmp.s
  ./tmp
  actual="$?"

  if [ "$actual" = "$expected" ]; then
    echo "$input => $actual"
  else
    echo "$input => $expected expected, but got $actual"
    exit 1
  fi
}

assert 0 "tests/tests.go_"