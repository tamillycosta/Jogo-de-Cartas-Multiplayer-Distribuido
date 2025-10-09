package pubsub


// interface ara lidar com os topicos da aplicação
type HandleTopics interface {
    HandleTopic(clientID string, topic string, data interface{}) error
    GetTopics() []string
}

