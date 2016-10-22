package job

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/oikomi/FishChatServer2/dao/kafka"
	"github.com/oikomi/FishChatServer2/server/msg_server/conf"
	"github.com/oikomi/FishChatServer2/server/msg_server/job/model"
	"golang.org/x/net/context"
)

type KafkaProducer struct {
	producer       *kafka.Producer
	sendP2PMsgChan chan *model.SendP2PMsgKafka
}

func NewKafkaProducer() (kafkaProducer *KafkaProducer) {
	producer := kafka.NewProducer(conf.Conf.KafkaProducer.Producer)
	kafkaProducer = &KafkaProducer{
		producer:       producer,
		sendP2PMsgChan: make(chan *model.SendP2PMsgKafka, 1),
	}
	return
}

func (kp *KafkaProducer) SendP2PMsg(data *model.SendP2PMsgKafka) {
	kp.sendP2PMsgChan <- data
}

func (kp *KafkaProducer) HandleSuccess() {
	var (
		pm *sarama.ProducerMessage
	)
	for {
		pm = <-kp.producer.Successes()
		if pm != nil {
			glog.Info("producer message success, partition:%d offset:%d key:%v valus:%s", pm.Partition, pm.Offset, pm.Key, pm.Value)
		}
	}
}

func (kp *KafkaProducer) HandleError() {
	var (
		err *sarama.ProducerError
	)
	for {
		err = <-kp.producer.Errors()
		if err != nil {
			glog.Error("producer message error, partition:%d offset:%d key:%v valus:%s error(%v)", err.Msg.Partition, err.Msg.Offset, err.Msg.Key, err.Msg.Value, err.Err)
		}
	}
}

func (kp *KafkaProducer) Process() {
	var sendP2PMsg *model.SendP2PMsgKafka
	for {
		select {
		case sendP2PMsg = <-kp.sendP2PMsgChan:
			var (
				err    error
				vBytes []byte
			)
			if vBytes, err = json.Marshal(sendP2PMsg); err != nil {
				glog.Error(err)
				return
			}
			if err := kp.producer.Input(context.Background(), &sarama.ProducerMessage{
				Topic: conf.Conf.KafkaProducer.Topic,
				Key:   sarama.StringEncoder(model.SendP2PMsgKey),
				Value: sarama.ByteEncoder(vBytes),
			}); err != nil {
				glog.Error(err)
			}
		}
	}
}