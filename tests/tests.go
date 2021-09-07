package main

import (
	"fmt"
	"strconv"
)

func main() {
	testInt("arithmetic operation test 1", 21, 5+20-4)
	testInt("arithmetic operation test 2", 41, 12+34-5)
	testInt("arithmetic operation test 3", 47, 5+6*7)
	testInt("arithmetic operation test 4", 15, 5*(9-6))
	testInt("arithmetic operation test 5", 4, (3+5)/2)
	testInt("arithmetic operation test 6", 5, 10+-5)
	testInt("arithmetic operation test 7", 4, -10+-7*-2)

	testBool("boolean operation test 1", true, 1+1 == 2)
	testBool("boolean operation test 2", true, 1-5*2 == -9)
	testBool("boolean operation test 3", false, 4*-3 != -12)
	testBool("boolean operation test 4", false, 4*-3 != -12)
	testBool("boolean operation test 5", true, 1-5*2 != 9)
	testBool("boolean operation test 6", true, 5*10*-1 < 7*-7)
	testBool("boolean operation test 7", false, 2+3 < 5)
	testBool("boolean operation test 8", true, 2+3 <= 5)
	testBool("boolean operation test 9", true, 4*3 <= 5*7)
	testBool("boolean operation test 10", false, 5*10*-1 > 7*-7)
	testBool("boolean operation test 11", false, 2+3 > 5)
	testBool("boolean operation test 12", true, 2+3 >= 5)
	testBool("boolean operation test 13", false, 4*3 >= 5*7)
	testBool("boolean operation test 14", true, !(2+3 > 5))
	testBool("boolean operation test 15", false, !(2+3 >= 5))
	testBool("boolean operation test 16", true, 1+1 == 2 && 3+4 == 7)
	testBool("boolean operation test 17", false, 2+1 == 3 && 3+5 == 7 && 1+1 == 2)
	testBool("boolean operation test 18", true, 1+1 == 2 || 3+4 == 7)
	testBool("boolean operation test 19", true, 2+1 == 4 || 3+5 == 8 || 1+1 == 2)
	testBool("boolean operation test 20", false, 2+1 == 4 || 3+5 == 7 || 1+1 == 4)
	testBool("boolean operation test 21", true, 2+1 == 4 && 3 == 0 || 1+1 == 2)

	testInt("local var test 1", 4, localVarTest1())
	testInt("local var test 2", 7, localVarTest2())
	testInt("local var test 3", 21, localVarTest3())
	testInt("local var test 4", 222, localVarTest4())
	testInt("local var test 5", 22, localVarTest5())
	testInt("local var test 6", 12, localVarTest6())

	testInt("if stmt test 1", 1, ifStmtTest1())
	testInt("if stmt test 2", 2, ifStmtTest2())
	testInt("if stmt test 3", 34, ifStmtTest3())
	testInt("if stmt test 4", 1, ifStmtTest4())

	testInt("for stmt test 1", 15, forStmtTest1())
	testInt("for stmt test 2", 15, forStmtTest2())
	testInt("for stmt test 3", 15, forStmtTest3())

	testInt("func def test 1", 3, funcDefTest1())
	testInt("func def test 2", 8, funcDefTest2())

	testInt("pointer test 1", 4, pointerTest1())
	testInt("pointer test 2", 3, pointerTest2())

	testInt("type test 1", 100, typeTest1())

	testInt("top level var test 1", 22, topLevelTest1())
	testInt("top level var test 2", 4, topLevelTest2())
	testInt("top level var test 3", 20, topLevelTest3())

	testInt("array test 1", 6, arrayTest1())
	testInt("array test 2", 0, arrayTest2())
	testInt("array test 3", 1, arrayTest3())
	// testInt("array test 4", 12, arrayTest4())
	testInt("array test 5", 22, arrayTest5())

	testInt("rune test 1", 91, runeTest1())
	testInt("rune test 2", 3, runeTest2())
	testInt("rune test 3", 2, runeTest3())

	testInt("string test 1", 0, stringTest1())
	testInt("string test 2", 0, stringTest2())
	testInt("string test 3", 0, stringTest3())
	testInt("string test 4", 0, stringTest4())
	// testInt("string test 5", 0, stringTest5(21, 0, 0, 0))

	fmt.Println("comment test 1")
	commentTest1()

	testInt("short var decl test 1", 22, shortVarDeclTest1())
	testInt("short var decl test 2", 100, shortVarDeclTest2())

	testInt("multiple value test 1", 13, multipleValueTest1())

	testInt("scope test 1", 22, scopeTest1())

	testBool("bool test 1", true, boolTest1())
	testBool("bool test 2", true, boolTest2())

	testInt("slice test 1", 17, sliceTest1())
	testInt("slice test 2", 4, sliceTest2())
	testInt("slice test 3", 1000, sliceTest3())

	testInt("struct test 1", 100, structTest1())
	testInt("struct test 2", 200, structTest2())
	testInt("struct test 3", 300, structTest3())

	testInt("multiple file test 1", 10, multipleFileTest1())

	testInt("len test 1", 7, lenTest1())

	fmt.Println("OK")
}

func localVarTest1() int {
	var a = 3
	return a + 1
}

func localVarTest2() int { var z = 20; var a = 13; var x = z - a; return x }

func localVarTest3() int {
	var a = 5
	a + 3
	return 4*a + 1
}

func localVarTest4() int {
	var hello = 5*4 + 2
	var world = hello * 20 / 2
	return world + 2
}

func ifStmtTest1() int {
	var a = 2
	if a == 2 {
		a = a * 3
		a = 1
		return a
		a = 5
	}
	return a
}

func ifStmtTest2() int {
	var a = 2
	if a != 2 {
		a = 6
	}
	return a
}

func ifStmtTest3() int {
	var test = 16
	if test < 10 {
		test = 100
		test = test + 21
	} else {
		test = test - 5
		test = test - 1
		test = 3*test + 4
	}
	return test
}

func ifStmtTest4() int {
	var test = 14
	if test < 10 {
		return 0
	} else if test < 20 {
		return 1
	} else if test < 30 {
		return 2
	} else {
		return 3
	}
}

func forStmtTest1() int {
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

func forStmtTest2() int {
	var i = 1
	var sum = 0
	for i < 6 {
		sum = sum + i
		i = i + 1
	}
	return sum
}

func forStmtTest3() int {
	var sum = 0
	var i int
	for i = 0; i < 6; i = i + 1 {
		sum = sum + i
	}
	return sum
}

func foo(a int, b int) int {
	return a + b
}

func funcDefTest1() int {
	return foo(2, 1)
}

func fib(n int) int {
	if n <= 1 {
		return 1
	}
	return fib(n-1) + fib(n-2)
}

func funcDefTest2() int {
	return fib(5)
}

func pointerTest1() int {
	var str = 4
	return *&str
}

func localVarTest5() int {
	var a int = 3
	var b int
	b = 19
	return a + b
}

func pointerTest2() int {
	var x int
	var y *int

	y = &x
	*y = 3
	return x
}

func localVarTest6() int {
	var x1 int = 1
	var x11 int = 11
	return x1 + x11
}

func Mul5(n int) int {
	return n * 5
}

func typeTest1() int {
	var n = Mul5(10) * 2
	var m int = n
	return m
}

var number int

func topLevelTest1() int {
	number = 5
	return 10*number/2 - 3
}

func bar(number int) int {
	return number
}

func topLevelTest2() int {
	number = 5
	return bar(4)
}

func baz(n int) int {
	return number * n
}

func topLevelTest3() int {
	number = 5
	return baz(4)
}

func arrayTest1() int {
	var arr [10]int
	arr[0] = 1
	arr[1] = 2
	arr[2] = 3
	return arr[0] + arr[1] + arr[2]
}

func zero(n int) int {
	return 0
}

func arrayTest2() int {
	var arr [3]int
	return zero(arr[2])
}

var memo [3]int

func arrayTest3() int {
	memo[0] = 1
	return memo[0] + memo[1]
}

/*
func arrayTest4() int {
	var memo2d [3][4]int
	memo2d[0][1] = 12
	return memo2d[0][1]
}
*/

func arrayTest5() int {
	var memo1 [3]int
	memo1[2] = 22

	var memo2 [3]int = memo1
	memo1[2] = 10

	return memo2[2]
}

func runeTest1() rune {
	var c1 rune = 40
	var c2 rune = 51
	return c1 + c2
}

func runeTest2() int {
	var x [3]rune
	x[0] = -1
	x[1] = 2
	var y int
	y = 4
	return 3
}

func runeTest3() int {
	return 'c' - 'a'
}

func stringTest1() int {
	var msg string = "hello world"
	fmt.Println(msg)
	return 0
}

func stringTest2() int {
	var a = 3
	var b = 2
	fmt.Println(strconv.Itoa(334))
	return 0
}

func stringTest3() int {
	fmt.Println(string([]rune{'h', 'e', 0}))
	fmt.Println(strconv.Itoa(1024))
	return 0
}

func stringTest4() int {
	var l = "Hello, "
	var r = "golang!"
	fmt.Println(l + r)
	fmt.Println("ok" + ", google")
	return 0
}

func commentTest1() {
	var comment = 0
	// ok?
	/*
	  comment = 10
	  日本語も行ける？
	*/
	return
	// ok！
}

func shortVarDeclTest1() int {
	number := 20
	number = number + 2
	return number
}

func shortVarDeclTest2() int {
	num2, num5 := 1, -10
	num2, num5 = 2, 50
	return num2 * num5
}

// (n番目のfibの値, n-1番目のfibの値)を返す
func fib2(n int) (int, int) {
	if n == 1 {
		return 1, 1
	}
	cur, prev := fib2(n - 1)
	return cur + prev, cur
}

func multipleValueTest1() int {
	// 1, 2, 3, 5, 8, 13
	var m int
	var _ int
	m, _ = fib2(6)
	return m
}

func scopeTest1() int {
	var n int = 10
	for n := 0; n < 10; n = n + 1 {
		// 特に何もしない
	}
	if 1+1 == 2 {
		n = 22
		var n int
		n = 3
		fmt.Println(strconv.Itoa(n))
	}
	return n
}

func boolTest1() bool {
	var n int = 10
	if true {
		n = n * 100
	}
	if false {
		n = 0
	}
	return n == 1000
}

func boolTest2() bool {
	return true
}

func sliceTest1() int {
	var a = []int{}
	a = append(a, 1)
	a = append(a, 17)
	return a[1]
}

func sliceTest2() int {
	var s = []string{}
	s = append(s, "hello")
	s = append(s, "world")
	fmt.Println(s[0])
	return 4
}

func sliceTest3() int {
	var a = []int{1, 10, 100, 1000, 10000}
	return a[3]
}

type Streamer struct {
	Name  string
	Power int
}

func structTest1() int {
	var s Streamer
	s.Name = "stylishotaku"
	s.Power = 100
	return s.Power
}

func structTest2() int {
	var s Streamer = Streamer{Name: "jun", Power: 200}
	var s2 = s
	return s2.Power
}

func structTest3() int {
	var s *Streamer = &Streamer{Name: "hello", Power: 0}
	s.Power = 300
	return s.Power
}

func multipleFileTest1() int {
	test2 = 10
	return test2
}

func lenTest1() int {
	var x = []int{1, 2, 3}
	var y = "he"
	var z [2]int
	return len(x) + len(y) + len(z)
}
