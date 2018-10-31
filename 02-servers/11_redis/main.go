package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/garyburd/redigo/redis"
)

var (
	redisAddr   = flag.String("addr", "redis://user:@localhost:6379/0", "redis addr")
	sessManager *SessionManager
	users       = map[string]string{
		"maxprosper":       "test",
		"maxim.proskurnia": "100500",
	}
)

func main() {
	flag.Parse()

	var err error
	redisConn, err := redis.DialURL(*redisAddr)
	if err != nil {
		log.Fatalf("can't connect to redis")
	}

	sessManager = NewSessionManager(redisConn)

	http.HandleFunc("/", innerPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/logout", logoutPage)

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}

func NewSessionManager(conn redis.Conn) *SessionManager {
	return &SessionManager{
		redisConn: conn,
	}
}

func (sm *SessionManager) Create(in *Session) (*SessionID, error) {
	id := SessionID{RandStringRunes(sessKeyLen)}
	dataSerialized, _ := json.Marshal(in)
	mkey := "sessions:" + id.ID
	result, err := redis.String(sm.redisConn.Do("SET", mkey, dataSerialized, "EX", 86400))
	if err != nil {
		return nil, err
	}
	if result != "OK" {
		return nil, fmt.Errorf("result no OK")
	}

	return &id, nil
}

func (sm *SessionManager) Check(in *SessionID) *Session {
	mkey := "session:" + in.ID
	data, err := redis.Bytes(sm.redisConn.Do("GET", mkey))
	if err != nil {
		log.Println("can't get data:", err)
		return nil
	}
	sess := &Session{}
	err = json.Unmarshal(data, sess)
	if err != nil {
		log.Println("can't unpack session data:", err)
		return nil
	}
	return sess
}

func (sm *SessionManager) Delete(in *SessionID) {
	mkey := "sessions:" + in.ID
	_, err := redis.Int(sm.redisConn.Do("DEL", mkey))
	if err != nil {
		log.Println("redis error:", err)
	}
}
