package cache

import (
	"context"
	"fmt"
	"testing"
)

func TestRedis(t *testing.T) {
	V8Example()
	sc := rdb.Get(context.Background(), "key")
	fmt.Print(sc)
	fmt.Println()
}
