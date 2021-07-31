// test.shで呼び出すための関数定義をするためのもの
#include <stdio.h>

int foo(int x, int y) {
    printf("テスト用関数fooの出力: %d + %d = %d\n", x, y, x + y);
    return x + y;
}

int bar(int x) {
    printf("テスト用関数barの出力: %d + 5 = %d\n", x, x + 5);
    return 0;
}