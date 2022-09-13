package main

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"os"
)

var (
	errorsLogger *log.Logger
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
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			errorsLogger.Println("ничё не вышло с ws upgrade")
		}
		go func() {
			defer conn.Close()
			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					errorsLogger.Println(err)
					return
				}
				err = wsutil.WriteServerMessage(conn, op, msg)
				if err != nil {
					errorsLogger.Println(err)
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
