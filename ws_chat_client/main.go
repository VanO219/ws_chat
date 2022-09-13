package main

import (
	"context"
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
	addr         = "192.168.1.3:4200"
	numOfGors    = 100000
	numOfMess    = 1000
	errorsLogger *log.Logger
)

func main() {
	l, err := os.OpenFile("./errors.log", os.O_CREATE|os.O_APPEND|os.O_SYNC|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	errorsLogger = log.New(l, "error", 0)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ctx, done := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(numOfGors)
	t := time.Now()
	for i := 0; i < numOfGors; i++ {
		fmt.Println(i)
		go client(ctx, wg, addr)
		//time.Sleep(50 * time.Microsecond)
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

	var (
		c   *websocket.Conn
		err error
	)

	for {
		c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			errorsLogger.Println("попытка установить конект: ", err)
			continue
		}
		break
	}

	//c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	//if err != nil {
	//	log.Println("попытка установить конект: ", err)
	//	return
	//}
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
					//log.Println("read:", err)
					return
				}
				//fmt.Println(string(mess))
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
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					errorsLogger.Println("write close:", err)
					return
				}
				return
			}
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				errorsLogger.Println("write:", err)
				return
			}
			nm--
		case <-ctx.Done():
			errorsLogger.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				errorsLogger.Println("write close:", err)
				return
			}
			return
		}
	}
}
