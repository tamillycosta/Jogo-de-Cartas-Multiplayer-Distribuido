package pubsub


//Estrutura genérica de mensagem Pub/Sub
type Message struct {
    Type  string      `json:"type"`
    Topic string      `json:"topic,omitempty"`
    Data  interface{} `json:"data,omitempty"`
}

//  Representa uma inscrição de cliente em tópico
type Subscription struct {
    ClientID string
    Topic    string
}