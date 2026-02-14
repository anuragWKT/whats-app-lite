package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	serverAddr = "localhost:8080"
	roomName   = "load_test_room"
)

func main() {
	clients := flag.Int("clients", 50, "Number of concurrent clients")
	msgs := flag.Int("msgs", 20, "Messages per client")
	flag.Parse()

	log.Printf(" Starting Load Test with %d clients sending %d messages each...", *clients, *msgs)

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < *clients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runClient(id, *msgs)
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	log.Println("------------------------------------------------")
	log.Printf("Load Test Completed in %v", duration)
	log.Printf("Total Messages Sent: %d", *clients**msgs)
	log.Printf("Throughput: %.2f msg/sec", float64(*clients**msgs)/duration.Seconds())
	log.Println("------------------------------------------------")
}

func runClient(id int, count int) {
	u := url.URL{
		Scheme:   "ws",
		Host:     serverAddr,
		Path:     "/ws",
		RawQuery: fmt.Sprintf("room=%s&username=bot_%d", roomName, id),
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("Client %d failed to connect: %v", id, err)
		return
	}
	defer c.Close()

	go func() {
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				return 
			}
		}
	}()

	for i := 0; i < count; i++ {
		msg := fmt.Sprintf(`{"type":"message","content":"Load test msg %d from bot %d","room":"%s","sender":"bot_%d"}`, i, id, roomName, id)
		
		err := c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Printf("Client %d write error: %v", id, err)
			return
		}
		
		
		time.Sleep(10 * time.Millisecond)
	}
}