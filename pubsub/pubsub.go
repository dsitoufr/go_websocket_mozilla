package pubsub

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

var (
	PUBLISH     = "publish"
	SUBSCRIBE   = "subscribe"
	UNSUBSCRIBE = "unsubscribe"
)

type PubSub struct {
	clients       []Client
	Subscriptions []Subscription
}

type Client struct {
	Id         string
	Connection *websocket.Conn
}

type Message struct {
	Action  string          `json:action`
	Topic   string          `json:topic`
	Message json.RawMessage `json:"message"`
}

type Subscription struct {
	Topic  string
	Client *Client
}

func (p *PubSub) AddClient(client *Client) *PubSub {
	p.clients = append(p.clients, *client)

	//server send message to client when adding to list
	msg := []byte("hello client id: " + client.Id)
	client.Connection.WriteMessage(1, msg)
	//fmt.Println("adding new client to list", client.Id)
	return p
}


func (client *Client) Send(message []byte) error {
	return client.Connection.WriteMessage(1, message)
}

//remove subscription by client
func (p *PubSub) RemoveClient(client *Client) *PubSub {

	for index, subscription := range p.Subscriptions {
		if subscription.Client.Id == client.Id {
			//remove all subscriptions of that client
			p.Subscriptions = append(p.Subscriptions[:index], p.Subscriptions[index+1:]...)
		}
	}

	//remove client from list
	for index, cli := range p.clients {
		if cli.Id == client.Id {
			p.clients = append(p.clients[:index], p.clients[index+1:]...)
		}
	}

	return p
}

func (p *PubSub) GetSubscription(topic string, client *Client) []Subscription {
	var subscriptionList []Subscription

	for _, subscription := range p.Subscriptions {
		if client != nil {
			if subscription.Client.Id == client.Id {
				subscriptionList = append(subscriptionList, subscription)
			}
		} else {
			subscriptionList = append(subscriptionList, subscription)
		}
	}
	return subscriptionList
}

func (p *PubSub) Subscribe(client *Client, topic string) *PubSub {

	clientSub := p.GetSubscription(topic, client)
	if len(clientSub) > 0 {
		//client already subscribed
		fmt.Println("client already subscribed at topic: ", topic)
		return p
	}

	s := &Subscription{Topic: topic,
		Client: client,
	}
	p.Subscriptions = append(p.Subscriptions, *s)
	fmt.Println("new subscriber at topic: ", topic, client.Id)
	return p
}

func (p *PubSub) Unsubscribe(client *Client, topic string) *PubSub {

	for index, sub := range p.Subscriptions {
		if (sub.Client.Id == client.Id) && (sub.Topic == topic) {
			p.Subscriptions = append(p.Subscriptions[:index], p.Subscriptions[index+1:]...)
		}
	}

	return p
}

func (p *PubSub) HandleReceiveMessage(client *Client, messageType int, payload []byte) *PubSub {
	m := &Message{}
	err := json.Unmarshal(payload, &m)
	if err != nil {
		fmt.Println("this is incorrect message payload.")
		return p
	}

	switch m.Action {

	case PUBLISH:
		p.Publish(m.Topic, m.Message, nil)
		fmt.Println("this is a publish new message")
		break
	case SUBSCRIBE:
		p.Subscribe(client, m.Topic)
		fmt.Println("new subscriber to topic: ", m.Topic)
		break
	case UNSUBSCRIBE:
		p.Unsubscribe(client, m.Topic)
		fmt.Println("client wants unsubscribe to topic: ", m.Topic, client.Id)
		break
	default:
		break
	}

	return p
}

func (p *PubSub) Publish(topic string, message []byte, excludeClient *Client) {
	subscriptions := p.GetSubscription(topic, nil)

	for _, sub := range subscriptions {
		fmt.Printf("sending to client id %s message is %s \n", sub.Client.Id, message)
		//sub.Client.Connection.WriteMessage(1, message)
		sub.Client.Send(message)
	}
}
