package signaling

import ("io"
        "log"
)


func (self *Message) NewBuddy() *Message {
	self.Type = "newbuddy"
	data := map[string]string {"uid": self.From, "from": self.From, "to": self.From, "type": "newbuddy"}
	self.Data = ToJsonString(&data)
	return self
}


func (self *Message) Uid() *Message {
	data := map[string]string {"uid": self.From, "from": self.From, "to": self.From, "type": "uid"}
	self.Data = ToJsonString(&data)
	return self
}


func (self *Message) Dropped() *Message {
	self.Type = "dropped"
	data := map[string]string {"from": self.From, "to": "", "type": "dropped"}
	self.Data = ToJsonString(&data)
	return self
}


func (self *RestrictedMsg) ReadJson(reader io.Reader) error {
	return ReadJson(reader, self)
}


type Broker struct {

	// room name -> client uid -> client chanel
	clients map[string]map[string]chan *Message

	// Channel into which messages are pushed to be broadcast out
	// to attached clients.
	//
	messages chan *Message
}


func NewBroker() *Broker {
	b := &Broker{
		make(map[string]map[string]chan *Message),
		make(chan *Message),
	}
	return b
}


func pushMessage(msg *Message, broker *Broker){
	if msg.To != "" {
		//Concrete message destination
		room, ok := broker.clients[msg.Room]
		if !ok {
				log.Printf("No such room %s", msg.Room)
				return
		}
		client_channel, ok := room[msg.To]
		if !ok {
				log.Printf("No such partcipant %s in room %s", msg.To, msg.Room)
			return
		}
		client_channel <- msg
	} else {
		//Global notification for all
		for name, q := range broker.clients[msg.Room] {
			if msg.From != name {
				q <- msg
			}
		}
	}
	log.Printf("Broadcast message to %s clients", msg.Room)
}
