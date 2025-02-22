package server

import (
	"encoding/json"
	"iLean/entity"
	"net/http"
	"strconv"
	"math"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	typeCommand = 2
)

func (s *Server) Controller(c *gin.Context) {
	commandNum, err := strconv.Atoi(c.Param("n"))
	if err != nil {
		c.JSON(http.StatusBadRequest, entity.Response{Status: http.StatusBadRequest, Message: "Status Bad Request"})
		return
	}
	// settings command
	switch commandNum {
	case 1:
		command := new(entity.CommandTemperature)
		c.ShouldBindJSON(&command)

		validate := validator.New()
		err := validate.Struct(command)
		if err != nil {
			c.JSON(http.StatusBadRequest, entity.Response{Status: http.StatusBadRequest, Message: "Status Bad Request"})
			return
		}

		mes, err := entity.NewCommand(*command, typeCommand)
		if err != nil {
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		dataBytes, err := json.Marshal(mes)

		if err != nil {
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		s.socket.Send("controller", command.SerialNumber, dataBytes)

		c.JSON(http.StatusOK, entity.Response{Status: http.StatusOK, Message: "ok"})
	case 2:
		command := new(entity.CommandTemperatureBySensor)

		c.ShouldBindJSON(&command)

		validate := validator.New()

		err := validate.Struct(command)

		if err != nil {
			c.JSON(http.StatusBadRequest, entity.Response{Status: http.StatusBadRequest, Message: "Status Bad Request"})
			return
		}

		mes, err := entity.NewCommand(*command, typeCommand)
		if err != nil {
			s.log.Error(err)
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		dataBytes, err := json.Marshal(mes)

		if err != nil {
			s.log.Error(err)
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		s.socket.Send("controller", command.SerialNumber, dataBytes)

		c.JSON(http.StatusOK, entity.Response{Status: http.StatusOK, Message: "ok"})

	case 3:
		command := new(entity.CommandDataVentModule)

		c.ShouldBindJSON(&command)

		validate := validator.New()

		err := validate.Struct(command)

		if err != nil {
			c.JSON(http.StatusBadRequest, entity.Response{Status: http.StatusBadRequest, Message: "Status Bad Request"})
			return
		}

		mes, err := entity.NewCommand(*command, typeCommand)
		if err != nil {
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		dataBytes, err := json.Marshal(mes)

		if err != nil {
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		s.socket.Send("controller", command.SerialNumber, dataBytes)

		c.JSON(http.StatusOK, entity.Response{Status: http.StatusOK, Message: "ok"})

	case 4:
		command := new(entity.CommandDataVentModule)

		c.ShouldBindJSON(&command)

		validate := validator.New()

		err := validate.Struct(command)

		if err != nil {
			c.JSON(http.StatusBadRequest, entity.Response{Status: http.StatusBadRequest, Message: "Status Bad Request"})
			return
		}

		mes, err := entity.NewCommand(*command, typeCommand)
		if err != nil {
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		dataBytes, err := json.Marshal(mes)

		if err != nil {
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		s.socket.Send("controller", command.SerialNumber, dataBytes)

		c.JSON(http.StatusOK, entity.Response{Status: http.StatusOK, Message: "ok"})

	case 5:
		command := new(entity.CommandDataVentModuleForAll)

		c.ShouldBindJSON(&command)

		validate := validator.New()

		err := validate.Struct(command)

		if err != nil {
			c.JSON(http.StatusBadRequest, entity.Response{Status: http.StatusBadRequest, Message: "Status Bad Request"})
			return
		}

		mes, err := entity.NewCommand(*command, typeCommand)
		if err != nil {
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		dataBytes, err := json.Marshal(mes)

		if err != nil {
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		s.socket.Send("controller", command.SerialNumber, dataBytes)

		c.JSON(http.StatusOK, entity.Response{Status: http.StatusOK, Message: "ok"})

	case 6:
		command := new(entity.CommandDataVentByZone)

		c.ShouldBindJSON(&command)

		validate := validator.New()

		err := validate.Struct(command)

		if err != nil {
			c.JSON(http.StatusBadRequest, entity.Response{Status: http.StatusBadRequest, Message: "Status Bad Request"})
			return
		}

		mes, err := entity.NewCommand(*command, typeCommand)
		if err != nil {
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		dataBytes, err := json.Marshal(mes)

		if err != nil {
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		s.socket.Send("controller", command.SerialNumber, dataBytes)

		c.JSON(http.StatusOK, entity.Response{Status: http.StatusOK, Message: "ok"})

	case 9:
		command := new(entity.CommandHysteresisOnHumidityModule)
		c.ShouldBindJSON(&command)

		validate := validator.New()

		err := validate.Struct(command)

		if err != nil {
			c.JSON(http.StatusBadRequest, entity.Response{Status: http.StatusBadRequest, Message: "Status Bad Request"})
			return
		}

		
		if command.Humidity == 0.0 {
			command.Humidity = math.MaxFloat32
		}
		if command.Hysteresis == 0 {
			command.Hysteresis = math.MaxUint8
		}

		rawRq, _ := json.Marshal(command)
		s.log.Infof("RAW REQUEST: %s", string(rawRq))

		mes, err := entity.NewCommand(*command, typeCommand)
		if err != nil {
			s.log.Error(err)
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}

		dataBytes, err := json.Marshal(mes)
		if err != nil {
			s.log.Error(err)
			c.JSON(http.StatusInternalServerError, entity.Response{Status: http.StatusInternalServerError, Message: "Status Internal Server Error"})
			return
		}
		
		s.log.Info(string(dataBytes))
		s.socket.Send("controller", command.SerialNumber, dataBytes)

		c.JSON(http.StatusOK, entity.Response{Status: http.StatusOK, Message: "ok"})

	default:
		c.JSON(http.StatusBadRequest, entity.Response{Status: http.StatusBadRequest, Message: "Status Bad Request"})

	}

}
