package signaling

import (
	"github.com/go-martini/martini"
	"summoners"
)

import "log"
import "fmt"
import "encoding/json"
import "bytes"
import "net/http"


func ClientStream(resp http.ResponseWriter, req *http.Request, params martini.Params, b *Broker) {
	f, ok := resp.(http.Flusher)
	if !ok {
		http.Error(resp, "Streaming unsupported!",
			http.StatusInternalServerError)
		return
	}
	c, ok := resp.(http.CloseNotifier)
	if !ok {
		http.Error(resp, "close notification unsupported",
			http.StatusInternalServerError)
		return
	}
	var streamSend = func(m *Message){
		if m.Id != "" {
			fmt.Fprintf(resp, "id: %s\n", m.Id)
		}
		fmt.Fprintf(resp, "event: %s\n", m.Type)
		fmt.Fprintf(resp, "data: %s\n", m.Data)
		f.Flush()
	}
	// Create a new channel, over which the broker can
	// send this client messages.
	messageChan := make(chan *Message)
	// Add this client to the map of those that should
	// receive updates
	var roomName = params["room"]
	// todo: add max members checking
	room := b.Room(roomName)

	members := len(room)
	uid := "Maybe " + summoners.NewName(members)
	message := &Message{"", "", "uid", uid, "", roomName}

	if members >= MaxMembers{
		streamSend(message.Rejected())
		return
	}
	headers := resp.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Connection", "keep-alive")
	f.Flush()
	closer := c.CloseNotify()
	msg := message.Uid()

	streamSend(msg)

	b.PushMessage(message.NewBuddy())
	room[uid] = messageChan
	// Remove this client from the map of attached clients
	// when `ClientStream` exits.
	defer func() {
		b.Release(roomName, uid)
		b.PushMessage(message.Dropped())
	}()

	for {
		select {
		case msg := <-messageChan:
			streamSend(msg)
		case <-closer:
			log.Println("Closing connection")
			return
		}
	}
}

func UpdateHandler(resp http.ResponseWriter, req *http.Request, params martini.Params, b *Broker) {
	buf := new(bytes.Buffer)
	bytes_read, _ := buf.ReadFrom(req.Body)
	defer req.Body.Close()
	var roomName = params["room"]
	log.Printf("Readed %d bytes from response", bytes_read)

	var data map[string]string
	json.Unmarshal(buf.Bytes(), &data)

	if len(data) == 0 || data["type"] == ""{
		http.Error(resp, "Bad Request", http.StatusBadRequest)
		return
	}
	log.Println("Ok", data)

	message := &Message{"", buf.String(), data["type"], data["from"], data["to"], roomName}
	b.PushMessage(message)

	resp.WriteHeader(200)
}


func OptionsHandler(resp http.ResponseWriter, req *http.Request){
	resp.WriteHeader(200)
}
