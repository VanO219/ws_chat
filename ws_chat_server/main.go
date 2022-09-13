package main

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
)

var (
	errorsLogger *log.Logger
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Пропускаем любой запрос
		},
	}
)

func main() {
	//e := echo.New()
	//e.GET("/ws", WsUpdate)
	//log.Fatalln(e.Start("192.168.1.3:4200"))
	l, err := os.OpenFile("./errors.log", os.O_CREATE|os.O_APPEND|os.O_SYNC|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	errorsLogger = log.New(l, "error", 0)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			errorsLogger.Println("ничё не вышло с ws upgrade")
		}
		go func() {
			defer connection.Close()
			for {
				mt, message, err := connection.ReadMessage()
				if err != nil {
					errorsLogger.Println(errors.Wrap(err, "connection.ReadMessage()"))
					return
				}
				if mt == websocket.CloseMessage {
					errorsLogger.Println("Получен код закрытия соединения")
					return
				}
				err = connection.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					errorsLogger.Println(errors.Wrap(err, "connection.WriteMessage(websocket.TextMessage, message)"))
					return
				}
			}
		}()
	})
	err = http.ListenAndServe("192.168.1.3:4200", nil)
	if err != nil {
		errorsLogger.Fatal("ListenAndServe: ", err)
	}
}

func WsUpdate(c echo.Context) (err error) {
	conn, _, _, err := ws.UpgradeHTTP(c.Request(), c.Response())
	if err != nil {
		return c.String(200, "ничё не вышло с ws upgrade")
	}
	go func() {
		defer conn.Close()

		for {
			msg, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				return
			}
			err = wsutil.WriteServerMessage(conn, op, msg)
			if err != nil {
				return
			}
		}
	}()
	return nil
}
