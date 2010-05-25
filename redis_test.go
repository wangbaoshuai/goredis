package redis

import (
    "os"
    "reflect"
    "runtime"
    "strconv"
    "strings"
    "testing"
)

const (
    // the timeout config property in redis.conf. used to test
    // connection retrying
    serverTimeout = 5
)

var client Client

func init() {
    runtime.GOMAXPROCS(2)
    client.Addr = "127.0.0.1:8379"
    client.Db = 13
}

func TestBasic(t *testing.T) {

    var val []byte
    var err os.Error

    err = client.Set("a", []byte("hello"))

    if err != nil {
        t.Fatal("set failed", err.String())
    }

    if val, err = client.Get("a"); err != nil || string(val) != "hello" {
        t.Fatal("get failed")
    }

    if typ, err := client.Type("a"); err != nil || typ != "string" {
        t.Fatal("type failed", typ)
    }

    //if keys, err := client.Keys("*"); err != nil || len(keys) != 1 {
    //    t.Fatal("keys failed", keys)
    //}

    client.Del("a")

    if ok, _ := client.Exists("a"); ok {
        t.Fatal("Should be deleted")
    }
}

func setget(t *testing.T, i int) {

    s := strconv.Itoa(i)
    err := client.Set(s, []byte(s))
    if err != nil {
        t.Fatal("Concurrent set", err.String())
    }

    s2, err := client.Get(s)

    if err != nil {
        t.Fatal("Concurrent get", err.String())
    }

    if s != string(s2) {
        t.Fatal("Concurrent: value not the same")
    }

    client.Del(s)
}

func TestEmptyGet(t *testing.T) {
    _, err := client.Get("failerer")

    if err == nil {
        t.Fatal("Expected an error")
    }
}

func TestConcurrent(t *testing.T) {
    for i := 0; i < 20; i++ {
        go setget(t, i)
    }
}


func TestSet(t *testing.T) {
    var err os.Error

    vals := []string{"a", "b", "c", "d", "e"}

    for _, v := range vals {
        client.Sadd("s", []byte(v))
    }

    var members [][]byte

    if members, err = client.Smembers("s"); err != nil || len(members) != 5 {
        t.Fatal("Set setup failed", err.String())
    }

    for _, v := range vals {
        if ok, err := client.Sismember("s", []byte(v)); err != nil || !ok {
            t.Fatal("Sismember test failed")
        }
    }

    for _, v := range vals {
        if ok, err := client.Srem("s", []byte(v)); err != nil || !ok {
            t.Fatal("Sismember test failed")
        }
    }

    if members, err = client.Smembers("s"); err != nil || len(members) != 0 {
        t.Fatal("Set setup failed", err.String())
    }

    client.Del("s")

}


func TestList(t *testing.T) {
    //var err os.Error

    vals := []string{"a", "b", "c", "d", "e"}

    for _, v := range vals {
        client.Rpush("l", []byte(v))
    }

    if l, err := client.Llen("l"); err != nil || l != 5 {
        t.Fatal("Llen failed", err.String())
    }

    for i := 0; i < len(vals); i++ {
        if val, err := client.Lindex("l", i); err != nil || string(val) != vals[i] {
            t.Fatal("Lindex failed", err.String())
        }
    }

    for i := 0; i < len(vals); i++ {
        if err := client.Lset("l", i, []byte("a")); err != nil {
            t.Fatal("Lset failed", err.String())
        }
    }

    for i := 0; i < len(vals); i++ {
        if val, err := client.Lindex("l", i); err != nil || string(val) != "a" {
            t.Fatal("Lindex failed", err.String())
        }
    }

    client.Del("l")

}

func verifyHash(t *testing.T, key string, expected map[string][]byte) {
    //test Hget
    m1 := make(map[string][]byte)
    for k, _ := range expected {
        actual, err := client.Hget("h", k)
        if err != nil {
            t.Fatal("verifyHash Hget failed", err.String())
        }
        m1[k] = actual
    }
    if !reflect.DeepEqual(m1, expected) {
        t.Fatal("verifyHash Hget failed")
    }


    //test Hkeys
    keys, err := client.Hkeys(key)
    if err != nil {
        t.Fatal("verifyHash Hkeys failed", err.String())
    }
    if len(keys) != len(expected) {
        t.Fatal("verifyHash Hkeys failed")
    }
    for _, key := range keys {
        if expected[key] == nil {
            t.Fatal("verifyHash Hkeys failed")
        }
    }

    //test Hvals
    vals, err := client.Hvals(key)
    if err != nil {
        t.Fatal("verifyHash Hvals failed", err.String())
    }
    if len(vals) != len(expected) {
        t.Fatal("verifyHash Hvals failed")
    }

    //test Hgetall
    m2, err := client.Hgetall(key)
    if err != nil {
        t.Fatal("verifyHash Hgetall failed", err.String())
    }
    if !reflect.DeepEqual(m2, expected) {
        t.Fatal("verifyHash Hgetall failed")
    }
}

func TestHash(t *testing.T) {
    //test cast
    keys := []string{"a", "b", "c", "d", "e"}
    test := make(map[string][]byte)
    for _, v := range keys {
        test[v] = []byte(strings.Repeat(v, 5))
    }

    //set with hset
    for k, v := range test {
        client.Hset("h", k, v)
    }
    //test hset
    verifyHash(t, "h", test)

    //set with hmset
    client.Hmset("h2", test)
    //test hset
    verifyHash(t, "h2", test)

    client.Del("h")
    client.Del("h2")
}

/*
func TestTimeout(t *testing.T) {
    client.Set("a", []byte("hello world"))

	time.Sleep((serverTimeout+10) * 1e9)
    val, err := client.Get("a")

    if err != nil {
        t.Fatal(err.String())
    }

    println(string(val))
}
*/
