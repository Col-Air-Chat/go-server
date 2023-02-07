package common

import (
	"col-air-go/model"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/streadway/amqp"
)

var (
	wg       sync.WaitGroup
	rdClient *redis.Client
	conn     *amqp.Connection
	ch       *amqp.Channel
	pool     chan struct{}
	host     = "localhost:6379"
	password = ""
	db       = 0
	amqpUrl  = ""
)

func SetRedisConfig(h string, p string, d int, a string) {
	host = h
	password = p
	db = d
	amqpUrl = a
}

func InitRedis() {
	rdClient = redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})
	var err error
	conn, err = amqp.Dial(amqpUrl)
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
					value := msg.Headers["value"]
					expire := int64(1 * 24 * time.Hour)
					if msg.Headers["expire"] != nil {
						expire = msg.Headers["expire"].(int64)
					}
					ctx, cancel := GetContextWithTimeout(5 * time.Second)
					defer cancel()
					rdClient.Set(ctx, key, value, time.Duration(expire))
					msg.Ack(false)
				case "del":
					key := msg.Headers["key"].(string)
					ctx, cancel := GetContextWithTimeout(5 * time.Second)
					defer cancel()
					rdClient.Del(ctx, key)
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
							ctx, cancel := GetContextWithTimeout(5 * time.Second)
							defer cancel()
							pipe.Set(ctx, action["key"].(string), action["value"], expire)
						case "del":
							ctx, cancel := GetContextWithTimeout(5 * time.Second)
							defer cancel()
							pipe.Del(ctx, action["key"].(string))
						}
					}
					ctx, cancel := GetContextWithTimeout(5 * time.Second)
					defer cancel()
					pipe.Exec(ctx)
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

func GetDataFromRedis(resultChan chan interface{}, resultErrChan chan error, key string) {
	defer close(resultChan)
	defer close(resultErrChan)
	ctx, cancel := GetContextWithTimeout(5 * time.Second)
	defer cancel()
	val, err := rdClient.Get(ctx, key).Result()
	if err != nil {
		resultErrChan <- err
	} else {
		resultErrChan <- nil
		resultChan <- val
	}
}

func PublishRedisMessage(ctx context.Context, channel string, msg model.WebsocketMsg) error {
	err := rdClient.Publish(ctx, channel, msg).Err()
	return err
}

func SubscribeRedisMessage(ctx context.Context, channel string) (string, error) {
	sub := rdClient.PSubscribe(ctx, channel)
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		return "", err
	}
	return msg.Payload, nil
}
