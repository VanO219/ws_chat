package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	addr      = "localhost:8080"
	numOfGors = 1000
	numOfMess = 1000
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ctx, done := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(numOfGors)
	t := time.Now()
	for i := 0; i < numOfGors; i++ {
		go client(ctx, wg, addr)
	}

	gosDone := make(chan struct{})
	go func() {
		wg.Wait()
		gosDone <- struct{}{}
	}()

	for {
		select {
		case <-interrupt:
			done()
			return
		case <-gosDone:
			fmt.Println("Все клиенты отработали за ", time.Since(t))
			return
		}
	}

}

func client(ctx context.Context, wg *sync.WaitGroup, addr string) {
	defer wg.Done()
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	//done := make(chan struct{})

	go func() {
		//defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, _, err := c.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}
			}
		}
	}()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	nm := numOfMess
	for {
		select {
		case t := <-ticker.C:
			if nm == 0 {
				return
			}
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
			nm--
		case <-ctx.Done():
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			return
		}
	}
}
