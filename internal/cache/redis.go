package cache

import (
	"context"
	"time"
	"usergenerator/internal"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	TTL time.Duration
	Exmpl *cache.Cache
	CTX context.Context
}

type Object struct {
    Str string
    Num int
}

func New(cfg internal.Config, ctx context.Context) (*Cache,error) {
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"server1": cfg.RedisURL,
		},
	})
	 che := cache.New(&cache.Options{
        Redis:      ring,
        LocalCache: cache.NewTinyLFU(1000, cfg.TTL),
    })
	
	ctx = context.TODO()
    key := "mykey"
    obj := &Object{
        Str: "mystring",
        Num: 42,
    }

    if err := che.Set(&cache.Item{
        Ctx:   ctx,
        Key:   key,
        Value: obj,
        TTL:   time.Hour,
    }); err != nil {
        return nil,err
    }
	var val Object

	che.Get(ctx,key,val)

	return &Cache{cfg.TTL,che,ctx},nil
}