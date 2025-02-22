package socket

import (
	"flag"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var addr = flag.String("addr", ":63240", "http service address")


func New() (*Hub, error)  {

	flag.Parse()
	hub := newHub()
	go hub.run()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})


	go func() {
		//defer server.wg.Done()
		for {
			logrus.Info("start ws")
			if err := http.ListenAndServe(*addr, nil); err != nil{
				if err == http.ErrServerClosed {
					return
				}
				logrus.Info("failed to start")
			}

			time.Sleep(3 * time.Second)

			logrus.Info("restarting ws")
		}
	}()


	return hub, nil

}