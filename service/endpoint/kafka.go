package endpoint

import (
	"github.com/Shopify/sarama"
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/kataras/golog"
	"strconv"
	"strings"
	"time"
)

type KafkaEndpoint struct {
	client sarama.Client
	// 改用阻塞式同步生产者
	producer sarama.SyncProducer

	//retryLock sync.Mutex
}

func newKafkaEndpoint() *KafkaEndpoint {
	r := &KafkaEndpoint{}
	return r
}

func (k *KafkaEndpoint) String() string {
	return "kafka , version:" + k.client.Config().Version.String()
}

func (k *KafkaEndpoint) Connect() error {
	cfg := sarama.NewConfig()
	cfg.Producer.Partitioner = sarama.NewRandomPartitioner

	// 同步消息必须为true
	cfg.Producer.Return.Successes = true
	cfg.Net.SASL.Enable = false

	if len(config.GlobalConfig.KafkaVersion) > 0 {
		v, err := sarama.ParseKafkaVersion(config.GlobalConfig.KafkaVersion)
		if err != nil {
			golog.Warnf("解析kafka版本号失败 %v", config.GlobalConfig.KafkaVersion)
		} else {
			cfg.Version = v
		}
	}

	if config.GlobalConfig.KafkaSASLUser != "" && config.GlobalConfig.KafkaSASLPassword != "" {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.User = config.GlobalConfig.KafkaSASLUser
		cfg.Net.SASL.Password = config.GlobalConfig.KafkaSASLPassword
	}

	var err error
	var client sarama.Client
	ls := strings.Split(config.GlobalConfig.KafkaAddr, ",")
	client, err = sarama.NewClient(ls, cfg)
	if err != nil {
		golog.Errorf("kafka client 产生失败 %s", err)
		return err
	}

	var producer sarama.SyncProducer
	producer, err = sarama.NewSyncProducerFromClient(client)
	if err != nil {
		golog.Errorf("kafka producer 生成失败 %s", err)
		return err
	}

	k.producer = producer
	k.client = client

	return nil
}

func (k *KafkaEndpoint) Ping() error {
	return k.client.RefreshMetadata()
}

func (k *KafkaEndpoint) Produce(data []byte) error {
	//var ms []*sarama.ProducerMessage

	err := k.sendRowMessage(data)
	if err != nil {
		golog.Error("send message error: ", err)
		return err
	}

	return nil
}

func (k *KafkaEndpoint) sendRowMessage(data []byte) (err error) {

	value := sarama.ByteEncoder(data)
	//golog.Infof("kafka body: %s",body)
	if value.Length() > 1000000 {
		golog.Warnf("kafka message过大 ,抛弃消息,  message: %s", data)
		return err
	}

	// kafka 消息的 key
	// 如果用相同的key 会造成分区不均匀
	key := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

	m := &sarama.ProducerMessage{
		Topic:     config.GlobalConfig.KafkaTopic,
		Value:     value,
		Key:       sarama.StringEncoder(key),
		Timestamp: time.Now(),
	}

	_, _, err = k.producer.SendMessage(m)

	if err != nil {
		golog.Errorf("kafka 发送失败: %s \n topic: %v. key: %v. content: %s", err, config.GlobalConfig.KafkaTopic, key, data)
		return err
	}
	return nil
}

func (k *KafkaEndpoint) Close() {
	if k.producer != nil {
		k.producer.Close()
	}
	if k.client != nil {
		k.client.Close()
	}

}

//
//func (k *KafkaEndpoint) buildMessages(row *model.RowMsg, meta *schema.Table) ([]*sarama.ProducerMessage, error) {
//	kvm := rowMap(row, rule, true)
//	ls, err := luaengine.DoMQOps(kvm, row.Action, rule)
//	if err != nil {
//		return nil, errors.Errorf("lua 脚本执行失败 : %s ", err)
//	}
//
//	var ms []*sarama.ProducerMessage
//	for _, resp := range ls {
//		m := &sarama.ProducerMessage{
//			Topic: resp.Topic,
//			Value: sarama.ByteEncoder(resp.ByteArray),
//		}
//		golog.Infof("topic: %s, message: %s", resp.Topic, string(resp.ByteArray))
//		ms = append(ms, m)
//	}
//
//	return ms, nil
//}
