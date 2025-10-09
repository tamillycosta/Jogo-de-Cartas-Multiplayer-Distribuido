package pubsub


// Message - Estrutura genérica de mensagem Pub/Sub
type Message struct {
    Type  string      `json:"type"`
    Topic string      `json:"topic,omitempty"`
    Data  interface{} `json:"data,omitempty"`
}

// Subscription - Representa uma inscrição de cliente em tópico
type Subscription struct {
    ClientID string
    Topic    string
}