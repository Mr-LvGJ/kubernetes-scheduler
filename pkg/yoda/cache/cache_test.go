package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-redis/redis/v8"
)

func TestRedis(t *testing.T) {
	V8Example()
	sc := rdb.Get(context.Background(), "key")
	fmt.Print(sc)
	fmt.Println()
}

func TestRedis2(t *testing.T) {
	c := New()
	c.Set(context.Background(), "hello", "world", 0)
	s, err := c.Get(context.Background(), "hello").Result()
	if err == redis.Nil {
		fmt.Println("no key")
	} else {
		fmt.Println(s + "=====")
	}

}
