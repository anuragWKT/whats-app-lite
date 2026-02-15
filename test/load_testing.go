package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var totalSent int64
var totalFailed int64

const (
	serverAddr = "localhost:8080"
	roomName   = "load_test_room"
	botPass    = "loadtest_pass_123"
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
	attempted := int64(*clients * *msgs)
	sent := atomic.LoadInt64(&totalSent)
	failed := atomic.LoadInt64(&totalFailed)

	log.Println("------------------------------------------------")
	log.Printf("Load Test Completed in %v", duration)
	log.Printf("Messages Attempted: %d", attempted)
	log.Printf("Messages Sent: %d", sent)
	log.Printf("Messages Failed: %d", failed)
	log.Printf("Throughput (successful): %.2f msg/sec", float64(sent)/duration.Seconds())
	log.Println("------------------------------------------------")
}

func runClient(id int, count int) {
	username := fmt.Sprintf("bot_%d", id)
	if err := registerUser(username, botPass); err != nil {
		log.Printf("Client %d failed to register: %v", id, err)
		return
	}

	u := url.URL{
		Scheme:   "ws",
		Host:     serverAddr,
		Path:     "/ws",
		RawQuery: fmt.Sprintf("room=%s&username=%s&password=%s", roomName, username, botPass),
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
		msg := fmt.Sprintf(`{"type":"message","content":"Load test msg %d from bot %d","room":"%s","sender":"%s"}`, i, id, roomName, username)

		err := c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			atomic.AddInt64(&totalFailed, 1)
			if !isDisconnectError(err) {
				log.Printf("Client %d unexpected write error: %v", id, err)
			}
			return
		}
		atomic.AddInt64(&totalSent, 1)

		time.Sleep(10 * time.Millisecond)
	}
}

func isDisconnectError(err error) bool {
	if err == nil {
		return false
	}

	if websocket.IsCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseAbnormalClosure,
	) {
		return true
	}

	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "broken pipe") ||
		strings.Contains(errText, "connection reset by peer") ||
		strings.Contains(errText, "websocket: close sent")
}

func registerUser(username, password string) error {
	payload := map[string]string{
		"username": username,
		"password": password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/register", serverAddr),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict {
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("register failed: status=%d body=%s", resp.StatusCode, string(respBody))
}
