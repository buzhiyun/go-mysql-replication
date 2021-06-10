package endpoint

import (
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/kataras/golog"
)

type Endpoint interface {
	Connect() error
	Ping() error
	Produce([]byte) error
	Close()
	String() string
}

func NewEndpoint() Endpoint {
	cfg := config.GlobalConfig

	//if cfg.IsRedis() {
	//	return newRedisEndpoint()
	//}
	//
	//if cfg.IsMongodb() {
	//	return newMongoEndpoint()
	//}
	//
	//if cfg.IsRocketmq() {
	//	return newRocketEndpoint()
	//}
	//
	//if cfg.IsRabbitmq() {
	//	return newRabbitEndpoint()
	//}

	if cfg.IsKafka() {
		golog.Info("endpoint 设置为 kafka")
		return newKafkaEndpoint()
	}

	//if cfg.IsEls() {
	//	if cfg.ElsVersion == 6 {
	//		return newElastic6Endpoint()
	//	}
	//	if cfg.ElsVersion == 7 {
	//		return newElastic7Endpoint()
	//	}
	//}

	//if cfg.IsScript() {
	//	return newScriptEndpoint()
	//}

	return nil
}
