package handlers

import (
	"log"
	"encoding/json"
	"sync"
	"net/http"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn
	sendClient     chan [3]int	// (x,y,value)
}

var (
	clients      = make(map[*Client]struct{})
	broadcast    = make(chan [3]int,1024)
	centralGrid = [640][480]int{}

	clientsMu sync.Mutex

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func HandleWebSockets(w http.ResponseWriter, r *http.Request) {
	
	// upgrade
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		conn:     conn,
		sendClient:     make(chan [3]int,256),
	}

	clientsMu.Lock()
	clients[client] = struct{}{}
	clientsMu.Unlock()
	
	data, err := json.Marshal(centralGrid)
	if err != nil {
		log.Println(err)
		clientsMu.Lock()
		delete(clients, client)
		clientsMu.Unlock()
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Println(err)
		clientsMu.Lock()
		delete(clients, client)
		clientsMu.Unlock()
		return
	}

	go client.read()
	go client.write()

}

type incomingPixel struct {
	Pixel [3]int `json:"pixel"`
}

func (c *Client) read() {
	// message from our client
	// reads the messages from curr grid, sends it to central hub
	
	defer func() {
		clientsMu.Lock()
		delete(clients,c)
		clientsMu.Unlock()
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var currPixelJSON incomingPixel
		err = json.Unmarshal(data, &currPixelJSON)
		if err != nil {
			log.Println(err)
			return
		}
		currPixel := currPixelJSON.Pixel
		
		broadcast <- currPixel
	}
}

func (c *Client) write() {
	// reads the messages from central hub, sends it to the client

	defer c.conn.Close()

	for currPixel := range c.sendClient {
		data, err := json.Marshal(currPixel)
		if err != nil {
			log.Println(err)
			return
		}
		err = c.conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func HandleBroadcast() {
	// reads central hub channel for new pixel update
	// if update, send the pixel to all client's channel and update central grid
	for {
		coord := <- broadcast
		
		// only this goroutine modifies centralGrid
		// so no mutex needed
		// send to other clients only if curr message changes the grid
		if (centralGrid[coord[0]][coord[1]]!=coord[2]){
			centralGrid[coord[0]][coord[1]]=coord[2]
			
			clientsMu.Lock()
			for client := range clients {
				select {
				case client.sendClient <- coord:
				default:
					close(client.sendClient)
					delete(clients, client)
				}
			}
			clientsMu.Unlock()
		}	
	}
}