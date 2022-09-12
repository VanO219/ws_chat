package main

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

func main() {
	//e := echo.New()
	//e.GET("/ws", WsUpdate)
	//log.Fatalln(e.Start("192.168.1.3:4200"))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			fmt.Println("ничё не вышло с ws upgrade")
		}
		go func() {
			defer conn.Close()

			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = wsutil.WriteServerMessage(conn, op, msg)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}()
	})
	err := http.ListenAndServe("192.168.1.3:4200", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
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
