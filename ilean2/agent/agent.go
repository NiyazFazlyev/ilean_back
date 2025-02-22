package agent

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"iLean/config"
	"iLean/entity"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jacobsa/go-serial/serial"
	"github.com/sirupsen/logrus"
)

const (
	preamOneByte byte = 85
	preamTwoByte byte = 170
	typeCommand  int  = 1
)

type Agent struct {
	connect *websocket.Conn
	buffer  []byte
	port    io.ReadWriteCloser
	mutex   sync.Mutex
	Config  config.Config
	err     chan error
	ctx     context.Context
	cancel  context.CancelFunc
	read    bool
}

func NewAgent(conf *config.Config) (*Agent, error) {

	agent := new(Agent)
	agent.Config = *conf
	agent.mutex = sync.Mutex{}
	agent.err = make(chan error)
	agent.ctx, agent.cancel = context.WithCancel(context.TODO())

	logrus.Info("connect to server via web-socket")
	agent.connectToSocket()

	logrus.Info("connect to device")
	agent.connectToDevice()

	go agent.Connector()
	go agent.ProcessingStream()

	return agent, nil
}

func (a *Agent) Connector() {
	for {
		<-a.err
		a.read = false

		a.Close()

		logrus.Info("lost connection")

		logrus.Info("reconnect")

		a.connectToSocket()

		a.connectToDevice()

		go a.ProcessingStream()

	}
}

func (a *Agent) connectToSocket() {
	for {
		connect, _, err := websocket.DefaultDialer.Dial("ws://185.27.192.21:63240/ws", nil)
		if err != nil {
			logrus.WithError(err).Error("failed to connect to server")
			time.Sleep(time.Second * 4)
			logrus.Info("reconnect")
			continue
		}

		type Greet struct {
			SerialNumber int    `json:"serial_number"`
			TypeClient   string `json:"type_client"`
		}

		greet := Greet{SerialNumber: a.Config.Serial, TypeClient: "controller"}

		byteGreet, err := json.Marshal(greet)

		if err != nil {
			logrus.WithError(err).Error("failed marshal json")
			time.Sleep(time.Second * 4)
			logrus.Info("reconnect")
			continue
		}

		err = connect.WriteMessage(websocket.TextMessage, byteGreet)

		if err != nil {
			logrus.WithError(err).Error("failed marshal json")
			time.Sleep(time.Second * 4)
			logrus.Info("reconnect")
			continue
		}

		_, mes, err := connect.ReadMessage()

		if err != nil {
			logrus.WithError(err).Error("failed read messages from server")
			time.Sleep(time.Second * 4)
			logrus.Info("reconnect")
			continue
		}

		if string(mes) == "ok" {
			logrus.Info("success connect to server")
			a.connect = connect
			return
		} else {
			logrus.WithError(err).Error("failed messages from server")
			time.Sleep(time.Second * 4)
			logrus.Info("reconnect")
			continue
		}

	}
}

func (a *Agent) connectToDevice() {

	for {

		options := serial.OpenOptions{
			PortName:        "/dev/ttyAMA0",
			DataBits:        8,
			BaudRate:        115200,
			StopBits:        1,
			MinimumReadSize: 2,
		}

		// Open the port.
		port, err := serial.Open(options)

		if err != nil {
			logrus.Error("failed connect to device")
			time.Sleep(time.Second * 4)
			logrus.Info("reconnect")
			continue

		} else {
			logrus.Info("success connect to device")
			a.port = port
			a.read = true
			return
		}

	}

}

func (a *Agent) ProcessingStream() {
	logrus.Info("start processing data")

	go func(agent *Agent) {
		// команды прилетают с сервера и отправляются в малинку
		for {
			_, mes, err := agent.connect.ReadMessage()

			logrus.Info("Received an income command")
			logrus.Info(string(mes))

			if err != nil {
				logrus.Error(err)
				a.err <- err

				return
			}

			var commands entity.Command
			err = json.Unmarshal(mes, &commands)
			if err != nil {
				logrus.Error(err)

				continue
			}

			logrus.Info("command: ", commands.TypeCommand)

			buf := bytes.NewBuffer([]byte{})
			if err := binary.Write(buf, binary.LittleEndian, []byte{preamOneByte, preamTwoByte, preamOneByte, preamTwoByte}); err != nil {
				logrus.Error(err)
			}
			commandToLength := map[int]int16{
				1: 14,
				2: 14,
				3: 13,
				4: 13,
				5: 13,
				6: 7,
				9: 14,
			}
			lenDataBytes := commandToLength[commands.TypeCommand]
			if err := binary.Write(buf, binary.LittleEndian, lenDataBytes); err != nil {
				logrus.Error(err)
			}

			var address int32 = 101
			if err := binary.Write(buf, binary.LittleEndian, address); err != nil {
				logrus.Error(err)
			}

			commandTypeToCommand := map[int]int8{
				1: 2,
				2: 2,
				3: 4,
				4: 4,
				5: 5,
				6: 6,
				9: 2,
			}

			command := commandTypeToCommand[commands.TypeCommand]

			if err := binary.Write(buf, binary.LittleEndian, command); err != nil {
				logrus.Error(err)
			}
			logrus.Info("agent start switch")
			switch commands.TypeCommand {
			case 1:
				var commandTemperature entity.CommandTemperature
				err = json.Unmarshal(commands.Data, &commandTemperature)

				zone4Byte := int32(commandTemperature.Zone)
				if err := binary.Write(buf, binary.LittleEndian, zone4Byte); err != nil {
					logrus.Error(err)
				}

				temperature4Byte := commandTemperature.Temperature
				if err := binary.Write(buf, binary.LittleEndian, temperature4Byte); err != nil {
					logrus.Error(err)
				}

				types := uint8(0)
				if err := binary.Write(buf, binary.LittleEndian, types); err != nil {
					logrus.Error(err)
				}
			case 2:
				var commandTemperatureBySensor entity.CommandTemperatureBySensor
				err = json.Unmarshal(commands.Data, &commandTemperatureBySensor)

				var zone4Byte int32 = int32(commandTemperatureBySensor.Zone)
				if err := binary.Write(buf, binary.LittleEndian, zone4Byte); err != nil {
					logrus.Error(err)
				}

				var temperature4Byte float32 = float32(0)

				if err := binary.Write(buf, binary.LittleEndian, temperature4Byte); err != nil {
					logrus.Error(err)
				}

				var types uint8 = uint8(commandTemperatureBySensor.Type)

				if err := binary.Write(buf, binary.LittleEndian, types); err != nil {
					logrus.Error(err)
				}
			case 3:
				var command entity.CommandDataVentModule
				err = json.Unmarshal(commands.Data, &command)

				var (
					zone      int16 = command.Zone
					ventSpeed int16 = command.VentSpeed
				)

				if command.ForAll {
					zone = 0
					ventSpeed = 0
				}

				if err := binary.Write(buf, binary.LittleEndian, zone); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, ventSpeed); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, command.Delta); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, command.TypeRegulation); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, command.IntervalTimeVentilationDampers); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, command.VentilationPeriodAfterCO2ReductionTime); err != nil {
					logrus.Error(err)
				}
			case 4:
				var command entity.CommandDataVentModule
				err = json.Unmarshal(commands.Data, &command)

				var (
					zone      int16 = command.Zone
					ventSpeed int16 = command.VentSpeed
				)

				if command.ForAll {
					zone = 0
					ventSpeed = 0
				}

				if err := binary.Write(buf, binary.LittleEndian, zone); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, ventSpeed); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, command.Delta); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, command.TypeRegulation); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, command.IntervalTimeVentilationDampers); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, command.VentilationPeriodAfterCO2ReductionTime); err != nil {
					logrus.Error(err)
				}
				logrus.Info("agent command 4 success", err)
			case 5:
				var command entity.CommandDataVentModuleForAll
				err = json.Unmarshal(commands.Data, &command)

				var (
					zone      int16
					ventSpeed int16
				)

				if err := binary.Write(buf, binary.LittleEndian, zone); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, ventSpeed); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, command.Delta); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, command.TypeRegulation); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, command.IntervalTimeVentilationDampers); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, command.VentilationPeriodAfterCO2ReductionTime); err != nil {
					logrus.Error(err)
				}
				logrus.Info("agent command 5 success", err)
			case 6:
				var command entity.CommandDataVentByZone
				err = json.Unmarshal(commands.Data, &command)
				var (
					zone int16 = command.Zone
				)

				if err := binary.Write(buf, binary.LittleEndian, zone); err != nil {
					logrus.Error(err)
				}
			case 9:
				var commHysteresis entity.CommandHysteresisOnHumidityModule
				logrus.Info(commands.Data)
				err = json.Unmarshal(commands.Data, &commHysteresis)
				
				if err := binary.Write(buf, binary.LittleEndian, commHysteresis.Zone); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, commHysteresis.Humidity); err != nil {
					logrus.Error(err)
				}

				if err := binary.Write(buf, binary.LittleEndian, commHysteresis.Hysteresis); err != nil {
					logrus.Error(err)
				}
			default:
				logrus.Error("not found command for unmarshaling command")
				continue
			}

			h := uint16(32767)
			if err := binary.Write(buf, binary.LittleEndian, h); err != nil {
				logrus.Error(err)
			}

			_, err = a.port.Write(buf.Bytes())
			if err != nil {
				logrus.WithError(err).Error("failed write to device")
			}
			logrus.Info("success write to device ", buf.Bytes())
		}
	}(a)

	preambles := Preambles{}
	preambles.Init()

	for {
		// команды прилетают с малинки и отправляются на сервер
		buf := make([]byte, 1)

		if !a.read {
			logrus.Info("return")
			return
		}

		_, err := a.port.Read(buf)

		if err != nil {
			logrus.Println("Error reading from serial port: ", err)
			if err != io.EOF {
				logrus.Println("Error reading from serial port: ", err)
			}

		} else {
			preamblesByteBuff := buf[0]

			if preamblesByteBuff == preamOneByte || preambles.One {
				preambles.One = true

				if preamblesByteBuff == preamOneByte && !preambles.Two {
					continue
				}

				if preamblesByteBuff == preamTwoByte || preambles.Two {
					preambles.Two = true

					if preamblesByteBuff == preamTwoByte && !preambles.Three {
						continue
					}

					if preamblesByteBuff == preamOneByte || preambles.Three {
						preambles.Three = true

						if preamblesByteBuff == preamOneByte {
							continue
						}

						if preamblesByteBuff == preamTwoByte || preambles.Four {
							preambles.Four = true

						} else {
							preambles.Reset()
						}
					} else {
						preambles.Reset()
					}
				} else {
					preambles.Reset()
				}

			} else {
				preambles.Reset()
			}

			if preambles.One && preambles.Two && preambles.Three && preambles.Four {
				preambles.Reset()

				lenDataBuf := make([]byte, 2)
				_, err := a.port.Read(lenDataBuf)
				if err != nil {
					if err != io.EOF {
						logrus.Println("Error reading from serial port: ", err)
					}
				}

				var lenDataBytes int16
				buf := bytes.NewReader(lenDataBuf)
				err = binary.Read(buf, binary.LittleEndian, &lenDataBytes)
				if err != nil {
					logrus.Error("binary.Read failed:", err)
				}

				data := make([]byte, 5)
				_, err = a.port.Read(data)
				if err != nil {
					if err != io.EOF {
						logrus.Println("Error reading from serial port: ", err)
					}
				}

				type DataAddresCommand struct {
					Addres  int32 `json:"addres"`
					Command uint8 `json:"command"`
				}

				var dataAddresCommand DataAddresCommand
				bufReader := bytes.NewReader(data)
				err = binary.Read(bufReader, binary.LittleEndian, &dataAddresCommand)
				if err != nil {
					logrus.Error("binary.Read failed:", err)
					logrus.Info("datebase", dataAddresCommand)
				}

				switch dataAddresCommand.Command {
				case 0:
					var (
						tempAir     float32
						humidityAir float32
						tempfloor   float32
						co2         int32
						crc         uint16
					)

					{
						data = a.readOneByte(4)

						buf := bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &tempAir)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}
					}

					{
						data = a.readOneByte(4)

						buf = bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &humidityAir)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}
					}

					{
						data = a.readOneByte(4)
						buf = bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &tempfloor)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}
					}

					{
						data = a.readOneByte(4)

						buf = bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &co2)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}
					}

					{
						data = a.readOneByte(2)
						buf = bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &crc)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}

					}

					dataCommandTemperature := entity.DataCommandTemperature{
						Zone:        dataAddresCommand.Addres,
						TempAir:     tempAir,
						HumidityAir: int(humidityAir),
						Tempfloor:   tempfloor,
						CO2:         float32(co2),
					}

					command, err := entity.NewCommand(dataCommandTemperature, typeCommand)
					if err != nil {
						logrus.WithError(err).Error("failed commnd 0")
						continue
					}

					data, err := json.Marshal(command)

					a.mutex.Lock()
					err = a.connect.WriteMessage(websocket.TextMessage, data)
					logrus.WithField("command 0", dataCommandTemperature).Info("send to messages")
					a.mutex.Unlock()

					if err != nil {
						logrus.WithError(err).Error("connection to lost")
						a.err <- err
						return
					}
				case 1:
					logBuff := make([]byte, 0)

					dataSlice := make([]entity.DataCommandTemperatureBySensor, 0)

					var crc uint16

					for i := 0; i < 6; i++ {
						var (
							zone              int32
							setpointValueTemp float32
							typeRegulation    uint8
						)

						{
							data = a.readOneByte(4)

							buf = bytes.NewReader(data)
							err = binary.Read(buf, binary.LittleEndian, &zone)
							if err != nil {
								logrus.Error("binary.Read failed:", err)
							}

							for _, item := range data {
								logBuff = append(logBuff, item)
							}

						}

						{
							data = a.readOneByte(4)

							buf = bytes.NewReader(data)
							err = binary.Read(buf, binary.LittleEndian, &setpointValueTemp)
							if err != nil {
								logrus.Error("binary.Read failed:", err)
							}

							for _, item := range data {
								logBuff = append(logBuff, item)
							}

						}

						{
							data = a.readOneByte(1)

							buf = bytes.NewReader(data)
							err = binary.Read(buf, binary.LittleEndian, &typeRegulation)
							if err != nil {
								logrus.Error("binary.Read failed:", err)
							}
							for _, item := range data {
								logBuff = append(logBuff, item)
							}

							typeRegulation -= 1
						}

						dataSlice = append(dataSlice, entity.DataCommandTemperatureBySensor{
							Zone:              zone,
							SetpointValueTemp: setpointValueTemp,
							TypeRegulation:    typeRegulation,
						})

					}

					{
						data = a.readOneByte(2)

						buf = bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &crc)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}

						for _, item := range data {
							logBuff = append(logBuff, item)
						}
					}

					a.readOneByte(1)

					command, err := entity.NewCommand(dataSlice, typeCommand)
					if err != nil {
						logrus.WithError(err).Error("failed commnd 1")
						continue
					}

					dataBytes, err := json.Marshal(command)

					a.mutex.Lock()
					err = a.connect.WriteMessage(websocket.TextMessage, dataBytes)
					logrus.WithField("command 1", dataSlice).Info("send to messages")
					a.mutex.Unlock()

					if err != nil {
						logrus.WithError(err).Error("connection to lost")
						a.err <- err
						return
					}
				case 3:
					logrus.Info("start 3 command")
					logBuff := make([]byte, 0)

					dataSlice := make([]entity.DataVent, 0)

					var crc uint16

					for i := 0; i < 9; i++ {

						var (
							zone      int16
							ventSpeed int16
						)

						{
							data = a.readOneByte(2)

							buf = bytes.NewReader(data)
							err = binary.Read(buf, binary.LittleEndian, &zone)
							if err != nil {
								logrus.Error("binary.Read failed:", err)
							}

							for _, item := range data {
								logBuff = append(logBuff, item)
							}

						}

						{
							data = a.readOneByte(2)

							buf = bytes.NewReader(data)
							err = binary.Read(buf, binary.LittleEndian, &ventSpeed)
							if err != nil {
								logrus.Error("binary.Read failed:", err)
							}

							for _, item := range data {
								logBuff = append(logBuff, item)
							}

						}

						dataSlice = append(dataSlice, entity.DataVent{
							Zone:      zone,
							VentSpeed: ventSpeed,
						})
					}

					logrus.Info(dataSlice)

					{
						data = a.readOneByte(2)

						buf = bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &crc)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}

						for _, item := range data {
							logBuff = append(logBuff, item)
						}
					}

					a.readOneByte(1)

					command, err := entity.NewCommand(dataSlice, typeCommand)
					if err != nil {
						logrus.WithError(err).Error("failed commnd 3")
						continue
					}

					dataBytes, err := json.Marshal(command)

					a.mutex.Lock()
					err = a.connect.WriteMessage(websocket.TextMessage, dataBytes)
					logrus.WithField("command 3", dataSlice).Info("send to messages")
					a.mutex.Unlock()

					if err != nil {
						logrus.WithError(err).Error("connection to lost")
						a.err <- err
						return
					}
				case 2:
					logrus.Info("start 2 command")
				case 4:

				case 5:
					logrus.Info("start 5 command")
				case 6:
					logrus.Info("start 6 command")
				case 7:

					var (
						zone                                   int16
						ventSpeed                              int16
						delta                                  uint8
						typeRegulation                         uint8
						intervalTimeVentilationDampers         uint8
						ventilationPeriodAfterCO2ReductionTime uint8
					)

					{
						data = a.readOneByte(2)

						buf := bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &zone)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}
					}

					{
						data = a.readOneByte(2)

						buf = bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &ventSpeed)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}
					}

					{
						data = a.readOneByte(1)

						buf = bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &delta)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}
					}

					{
						data = a.readOneByte(1)

						buf = bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &typeRegulation)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}
					}

					{
						data = a.readOneByte(1)
						buf = bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &intervalTimeVentilationDampers)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}
					}

					{
						data = a.readOneByte(1)
						buf = bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &ventilationPeriodAfterCO2ReductionTime)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}

					}

					dataCommandVentModule := entity.DataVentModule{
						Zone:                                   zone,
						VentSpeed:                              ventSpeed,
						Delta:                                  delta,
						TypeRegulation:                         typeRegulation,
						IntervalTimeVentilationDampers:         intervalTimeVentilationDampers,
						VentilationPeriodAfterCO2ReductionTime: ventilationPeriodAfterCO2ReductionTime,
					}

					command, err := entity.NewCommand(dataCommandVentModule, typeCommand)
					if err != nil {
						logrus.WithError(err).Error("failed commnd 7")
						continue
					}

					data, err := json.Marshal(command)

					a.mutex.Lock()
					err = a.connect.WriteMessage(websocket.TextMessage, data)
					logrus.WithField("command 7", dataCommandVentModule).Info("send to messages")
					a.mutex.Unlock()

					if err != nil {
						logrus.WithError(err).Error("connection to lost")
						a.err <- err
						return
					}
				case 8:
					logBuff := make([]byte, 0)

					dataSlice := make([]entity.DataCommandHumidityModule, 0)

					var crc uint16

					for i := 0; i < 6; i++ {
						var (
							zone       int32
							setpoint   float32
							hysteresis uint8
						)

						{
							data = a.readOneByte(4)

							buf = bytes.NewReader(data)
							err = binary.Read(buf, binary.LittleEndian, &zone)
							if err != nil {
								logrus.Error("binary.Read failed:", err)
							}

							for _, item := range data {
								logBuff = append(logBuff, item)
							}

						}

						{
							data = a.readOneByte(4)

							buf = bytes.NewReader(data)
							err = binary.Read(buf, binary.LittleEndian, &setpoint)
							if err != nil {
								logrus.Error("binary.Read failed:", err)
							}

							for _, item := range data {
								logBuff = append(logBuff, item)
							}

						}

						{
							data = a.readOneByte(1)

							buf = bytes.NewReader(data)
							err = binary.Read(buf, binary.LittleEndian, &hysteresis)
							if err != nil {
								logrus.Error("binary.Read failed:", err)
							}
							for _, item := range data {
								logBuff = append(logBuff, item)
							}

							hysteresis -= 1
						}

						dataSlice = append(dataSlice, entity.DataCommandHumidityModule{
							Zone:       zone,
							Setpoint:   setpoint,
							Hysteresis: hysteresis,
						})

					}

					{
						data = a.readOneByte(2)

						buf = bytes.NewReader(data)
						err = binary.Read(buf, binary.LittleEndian, &crc)
						if err != nil {
							logrus.Error("binary.Read failed:", err)
						}

						for _, item := range data {
							logBuff = append(logBuff, item)
						}
					}

					a.readOneByte(1)

					command, err := entity.NewCommand(dataSlice, typeCommand)
					if err != nil {
						logrus.WithError(err).Error("failed commnd 8")
						continue
					}

					dataBytes, err := json.Marshal(command)

					a.mutex.Lock()
					err = a.connect.WriteMessage(websocket.TextMessage, dataBytes)
					logrus.WithField("command 8", dataSlice).Info("send to messages")
					a.mutex.Unlock()

					if err != nil {
						logrus.WithError(err).Error("connection to lost")
						a.err <- err
						return
					}
				}
			}
		}
	}
}

func (a *Agent) readOneByte(lenBuff int) []byte {
	commonData := make([]byte, 0)

	for i := 0; i < lenBuff; i++ {
		data := make([]byte, 1)

		_, err := a.port.Read(data)
		if err != nil {
			if err != io.EOF {
				logrus.Println("Error reading from serial port: ", err)
			}
		}
		commonData = append(commonData, data[0])
	}

	return commonData
}

func (a *Agent) Close() {
	a.port.Close()
	a.connect.Close()
}

func test(b []byte, l int) {

	//t := [256]uint16{
	//0x0000, 0x1021, 0x2042, 0x3063, 0x4084, 0x50A5, 0x60C6, 0x70E7,
	//0x8108, 0x9129, 0xA14A, 0xB16B, 0xC18C, 0xD1AD, 0xE1CE, 0xF1EF,
	//0x1231, 0x0210, 0x3273, 0x2252, 0x52B5, 0x4294, 0x72F7, 0x62D6,
	//0x9339, 0x8318, 0xB37B, 0xA35A, 0xD3BD, 0xC39C, 0xF3FF, 0xE3DE,
	//0x2462, 0x3443, 0x0420, 0x1401, 0x64E6, 0x74C7, 0x44A4, 0x5485,
	//0xA56A, 0xB54B, 0x8528, 0x9509, 0xE5EE, 0xF5CF, 0xC5AC, 0xD58D,
	//0x3653, 0x2672, 0x1611, 0x0630, 0x76D7, 0x66F6, 0x5695, 0x46B4,
	//0xB75B, 0xA77A, 0x9719, 0x8738, 0xF7DF, 0xE7FE, 0xD79D, 0xC7BC,
	//0x48C4, 0x58E5, 0x6886, 0x78A7, 0x0840, 0x1861, 0x2802, 0x3823,
	//0xC9CC, 0xD9ED, 0xE98E, 0xF9AF, 0x8948, 0x9969, 0xA90A, 0xB92B,
	//0x5AF5, 0x4AD4, 0x7AB7, 0x6A96, 0x1A71, 0x0A50, 0x3A33, 0x2A12,
	//0xDBFD, 0xCBDC, 0xFBBF, 0xEB9E, 0x9B79, 0x8B58, 0xBB3B, 0xAB1A,
	//0x6CA6, 0x7C87, 0x4CE4, 0x5CC5, 0x2C22, 0x3C03, 0x0C60, 0x1C41,
	//0xEDAE, 0xFD8F, 0xCDEC, 0xDDCD, 0xAD2A, 0xBD0B, 0x8D68, 0x9D49,
	//0x7E97, 0x6EB6, 0x5ED5, 0x4EF4, 0x3E13, 0x2E32, 0x1E51, 0x0E70,
	//0xFF9F, 0xEFBE, 0xDFDD, 0xCFFC, 0xBF1B, 0xAF3A, 0x9F59, 0x8F78,
	//0x9188, 0x81A9, 0xB1CA, 0xA1EB, 0xD10C, 0xC12D, 0xF14E, 0xE16F,
	//0x1080, 0x00A1, 0x30C2, 0x20E3, 0x5004, 0x4025, 0x7046, 0x6067,
	//0x83B9, 0x9398, 0xA3FB, 0xB3DA, 0xC33D, 0xD31C, 0xE37F, 0xF35E,
	//0x02B1, 0x1290, 0x22F3, 0x32D2, 0x4235, 0x5214, 0x6277, 0x7256,
	//0xB5EA, 0xA5CB, 0x95A8, 0x8589, 0xF56E, 0xE54F, 0xD52C, 0xC50D,
	//0x34E2, 0x24C3, 0x14A0, 0x0481, 0x7466, 0x6447, 0x5424, 0x4405,
	//0xA7DB, 0xB7FA, 0x8799, 0x97B8, 0xE75F, 0xF77E, 0xC71D, 0xD73C,
	//0x26D3, 0x36F2, 0x0691, 0x16B0, 0x6657, 0x7676, 0x4615, 0x5634,
	//0xD94C, 0xC96D, 0xF90E, 0xE92F, 0x99C8, 0x89E9, 0xB98A, 0xA9AB,
	//0x5844, 0x4865, 0x7806, 0x6827, 0x18C0, 0x08E1, 0x3882, 0x28A3,
	//0xCB7D, 0xDB5C, 0xEB3F, 0xFB1E, 0x8BF9, 0x9BD8, 0xABBB, 0xBB9A,
	//0x4A75, 0x5A54, 0x6A37, 0x7A16, 0x0AF1, 0x1AD0, 0x2AB3, 0x3A92,
	//0xFD2E, 0xED0F, 0xDD6C, 0xCD4D, 0xBDAA, 0xAD8B, 0x9DE8, 0x8DC9,
	//0x7C26, 0x6C07, 0x5C64, 0x4C45, 0x3CA2, 0x2C83, 0x1CE0, 0x0CC1,
	//0xEF1F, 0xFF3E, 0xCF5D, 0xDF7C, 0xAF9B, 0xBFBA, 0x8FD9, 0x9FF8,
	//0x6E17, 0x7E36, 0x4E55, 0x5E74, 0x2E93, 0x3EB2, 0x0ED1, 0x1EF0,
	//}

	//var i uint16 = 0xFFFF
	//var crc16 uint16 = 0xFFFF
	//
	//for i := 0xFFFF; i < l; i++ {
	//	crc16 = (crc16 << 8) ^ t[(crc16 >> 8) ^ &b[0]]
	//}

}
