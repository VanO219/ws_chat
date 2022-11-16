package main

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
)

var (
	//lg *lg.lgger
	lg       *logrus.Logger
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Пропускаем любой запрос
		},
	}
)

func main() {
	//e := echo.New()
	//e.GET("/ws", WsUpdate)
	//lg.Fatalln(e.Start("192.168.1.3:4200"))
	l, err := os.OpenFile("errors.log", os.O_CREATE|os.O_APPEND|os.O_SYNC|os.O_TRUNC|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	}
	lg = logrus.New()
	lg.SetOutput(l)
	lg.SetFormatter(&logrus.TextFormatter{})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			lg.Println("ничё не вышло с ws upgrade")
			return
		}
		go func() {
			defer connection.Close()
			for {
				mt, message, err := connection.ReadMessage()
				if err != nil || mt == websocket.CloseMessage {
					if err != nil {
						lg.Println(errors.Wrap(err, "connection.ReadMessage()"))
					}
					return
				}
				//if  {
				//	lg.Println("Получен код закрытия соединения")
				//	return
				//}
				//if len(message) != 0 {
				//	lg.Println("Получено новое сообщение")
				//}
				err = connection.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					lg.Println(errors.Wrap(err, "connection.WriteMessage(websocket.TextMessage, message)"))
					return
				}
			}
		}()
	})
	lg.Println("Старт сервера")
	err = http.ListenAndServe("192.168.1.3:4200", nil)
	if err != nil {
		lg.Fatal("ListenAndServe: ", err)
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
