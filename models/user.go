package models

import (
    "fmt"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"github.com/go-redis/redis"
)

var (
	ErrUserNotFound = errors.New("User not found")
	ErrInvalidLogin = errors.New("Wrong password")	
)

type User struct {
    key string
}

func NewUser(username string, hash []byte) (*User, error) {
    // Returs a new id for a given key if the given key doesn't exist in redis.
    // If the key already exist then it will return its incremented value as id.
    // In this case key = "user:next-id"
    id, err := client.Incr("user:next-id").Result()
    if err != nil {
        return nil, err
    }
    key := fmt.Sprintf("user:%d", id)

    // Creates a pipeline for more optimized code.
    // Sends all the commands in one round trip instead of sending each command one at a time and wait for response
    pipe := client.Pipeline()
    pipe.HSet(key, "id", id)
    pipe.HSet(key, "username", username)
    pipe.HSet(key, "hash", hash)
    pipe.HSet("user:by-username", username, id)
    _, err = pipe.Exec()
    if err != nil {
        return nil, err
    }
    return &User{key}, nil
}

func (user *User) GetUsername() (string, error) {
    return client.HGet(user.key, "username").Result()
}

func (user *User) GetHash() ([]byte, error) {
    return client.HGet(user.key, "hash").Bytes()
}

func (user *User) Authenticate(password string) error {
    hash, err := user.GetHash()
    if err != nil {
        return err
    }
    err = bcrypt.CompareHashAndPassword(hash, []byte(password))
    if err == bcrypt.ErrMismatchedHashAndPassword {
        return ErrInvalidLogin
    }
    return err
}

func GetUserByUsername(username string) (*User, error) {
    id, err := client.HGet("user:by-username", username).Int64()
    if err == redis.Nil {
        return nil, ErrUserNotFound
    } else if err != nil {
        return nil, err
    } 
    key := fmt.Sprintf("user:%d", id)
    return &User{key}, nil
}


func AuthenticateUser(username, password string) error {
    user, err := GetUserByUsername(username)
    if err != nil {
        return err
    }
    return user.Authenticate(password)

}

func RegisterUser(username, password string) error {
    cost := bcrypt.DefaultCost
    hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
    if err != nil {
        return err
    }
    _, err = NewUser(username, hash)
    return err
}