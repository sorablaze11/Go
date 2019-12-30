// Database structure for redis DB
package models

import (
    "github.com/go-redis/redis"
)

var client *redis.Client

func Init () {
	// Redis instance initilization
    client = redis.NewClient(&redis.Options{
        Addr : "localhost:6379",
    })
}
