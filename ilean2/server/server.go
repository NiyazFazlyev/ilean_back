package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"iLean/server/socket"
	socketIO "iLean/server/socketio"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	httpServer *http.Server
	gin        *gin.Engine

	log *logrus.Entry

	wg   sync.WaitGroup
	stop chan struct{}

	jwt_key string

	socket *socket.Hub
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// @title hpnnm API
// @version 1.0
// @description REST API server
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @BasePath /api/v1
// @Query.collection.format multi
func NewServer(bindAddr, jwt_key string) (*Server, error) {
	socket, err := socket.New()

	if err != nil {
		logrus.Fatal(err)
	}

	socketIO := socketIO.NewSocketIO()
	go socketIO.Serve()
	//defer socketIO.Close()

	server := &Server{
		jwt_key: jwt_key,
		log:     logrus.WithField("subsystem", "web_server"),
		socket:  socket,
	}
	server.stop = make(chan struct{})

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(CORSMiddleware())

	r := router.Group("api/v1")
	{
		r.POST("/controller/command/:n", server.Controller)

		router.GET("/socket.io/*any", gin.WrapH(socketIO))
		router.POST("/socket.io/*any", gin.WrapH(socketIO))

	}

	server.httpServer = &http.Server{
		Addr:    bindAddr,
		Handler: router,
	}
	server.wg.Add(1)

	go func() {
		defer server.wg.Done()
		for {
			select {
			case <-server.stop:
				return
			default:
			}

			server.log.Info("starting")

			if err := server.httpServer.ListenAndServe(); err != nil {
				if err == http.ErrServerClosed {
					return
				}
				server.log.WithError(err).Error("failed to start")
			}

			time.Sleep(3 * time.Second)
		}
	}()

	return server, nil
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	close(s.stop)

	err := s.httpServer.Shutdown(ctx)

	if err != nil {
		s.log.WithError(err).Error("failed to gracefull shutdown")
	}

	s.wg.Wait()
}
