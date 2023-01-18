package redis

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/streadway/amqp"
)

var (
	wg       sync.WaitGroup
	rdClient *redis.Client
	conn     *amqp.Connection
	ch       *amqp.Channel
	pool     chan struct{}
)

func InitRedis() {
	rdClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "Col_Air_Redis#20230116",
		DB:       0,
	})
	_, err := rdClient.Ping().Result()
	if err != nil {
		panic(err)
	}
	conn, err = amqp.Dial("amqp://ColAirRedis:ColAirRedis@localhost:5672/ColRedis")
	if err != nil {
		panic(err)
	}
	ch, err = conn.Channel()
	if err != nil {
		panic(err)
	}
	q, _ := ch.QueueDeclare("redis", true, false, false, false, nil)
	ch.Qos(5, 0, false)
	msgs, _ := ch.Consume(q.Name, "", false, false, false, false, nil)
	pool = make(chan struct{}, 500)
	go func() {
		for msg := range msgs {
			pool <- struct{}{}
			wg.Add(1)
			go func(msg amqp.Delivery) {
				defer func() {
					<-pool
					wg.Done()
				}()
				actionType := msg.Headers["type"].(string)
				switch actionType {
				case "set":
					key := msg.Headers["key"].(string)
					value := msg.Headers["value"].(string)
					expire := int64(1 * 24 * time.Hour)
					if msg.Headers["expire"] != nil {
						expire = msg.Headers["expire"].(int64)
					}
					rdClient.Set(key, value, time.Duration(expire))
					msg.Ack(false)
				case "del":
					key := msg.Headers["key"].(string)
					rdClient.Del(key)
					msg.Ack(false)
				case "pipe":
					pipe := rdClient.Pipeline()
					var actions []map[string]interface{}
					json.Unmarshal(msg.Body, &actions)
					for _, action := range actions {
						switch action["type"] {
						case "set":
							expire := 1 * 24 * time.Hour
							if action["expire"] != nil {
								expireFloat64 := action["expire"].(float64)
								expire = time.Duration(expireFloat64)
							}
							pipe.Set(action["key"].(string), action["value"].(string), expire)
						case "del":
							pipe.Del(action["key"].(string))
						}
					}
					pipe.Exec()
					msg.Ack(false)
				}
			}(msg)
		}
	}()
}

func GetRedisClient() *redis.Client {
	if rdClient == nil {
		InitRedis()
	}
	return rdClient
}

func GetRedisChannel() *amqp.Channel {
	if ch == nil {
		InitRedis()
	}
	return ch
}

func GetDataFromRedis(resultChan chan string, key string) {
	defer close(resultChan)
	val, err := rdClient.Get(key).Result()
	if err != nil {
		resultChan <- err.Error()
	} else {
		resultChan <- val
	}
}
