#!/bin/bash

assert() {
  expected="$1"
  input="$2"

  ./main "library/fmt/" > tmp_fmt.s
  ./main "library/os/" > tmp_os.s
  ./main "library/strconv/" > tmp_strconv.s
  ./main "$input" > tmp.s
  gcc -no-pie -o tmp tmp.s tmp_fmt.s tmp_os.s tmp_strconv.s
  ./tmp
  actual="$?"

  if [ "$actual" = "$expected" ]; then
    echo "$input => $actual"
  else
    echo "$input => $expected expected, but got $actual"
    exit 1
  fi
}

assert 0 "tests/"
