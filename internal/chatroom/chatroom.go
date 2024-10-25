package chatroom

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type client struct {
	id   int
	conn *websocket.Conn
}

type ChatRoom struct {
	registeredConnections map[int]*websocket.Conn
	toUnregister          chan int
	toRegister            chan client
	messages              chan []byte
	shouldDelete          chan struct{}
}

func NewCR() *ChatRoom {
	cr := &ChatRoom{
		registeredConnections: make(map[int]*websocket.Conn),
		toUnregister:          make(chan int),
		toRegister:            make(chan client),
		messages:              make(chan []byte, 100),
		shouldDelete:          make(chan struct{}),
	}

	go cr.startPolling()
	return cr
}

func (c *ChatRoom) startPolling() {
	defer func() {
		c.shouldDelete <- struct{}{}
	}()

	for {
		select {
		case id := <-c.toUnregister:
			delete(c.registeredConnections, id)
			if len(c.registeredConnections) == 0 {
				close(c.shouldDelete)
				return
			}
		case newClient := <-c.toRegister:
			c.registeredConnections[newClient.id] = newClient.conn
			go c.connReader(newClient)
		default:
		}

		select {
		case msg := <-c.messages:
			for clid, conn := range c.registeredConnections {
				err := conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Println("failed to write json to a conn of a user: ", clid, err)
					c.toUnregister <- clid
				}
			}
		case <-time.After(time.Minute * 3):
			return
		}
	}
}

func (c *ChatRoom) AddClient(userid int, conn *websocket.Conn) {
	c.toRegister <- client{
		id:   userid,
		conn: conn,
	}
}

func (c *ChatRoom) RemoveClient(id int) {
	c.toUnregister <- id
}

func (c *ChatRoom) GetDoneChan() chan struct{} {
	return c.shouldDelete
}

func (c *ChatRoom) connReader(cl client) {
	for {
		_, msg, err := cl.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			c.toUnregister <- cl.id
			break
		}

		c.messages <- msg
	}
}
