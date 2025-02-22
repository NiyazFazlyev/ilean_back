package entity

import (
	"encoding/json"
	"errors"
)

const (
	CommandTemperatureIn = iota
	CommandTemperatureBySensorIn
	DataVentIn
	DataModuleVentIn
)

type CommandTemperature struct {
	SerialNumber int     `json:"serial_number" validate:"required"`
	Temperature  float32 `json:"temperature" validate:"required"`
	Zone         int     `json:"zone" validate:"required"`
}

type CommandTemperatureBySensor struct {
	SerialNumber int `json:"serial_number" validate:"required"`
	Type         int `json:"type" validate:"required"`
	Zone         int `json:"zone" validate:"required"`
}

type CommandDataVentModule struct {
	SerialNumber                           int   `json:"serial_number" validate:"required"`
	Zone                                   int16 `json:"zone"  validate:"required"`
	VentSpeed                              int16 `json:"vent_speed" validate:"required"`
	Delta                                  uint8 `json:"delta"`
	TypeRegulation                         uint8 `json:"type_regulation"`
	IntervalTimeVentilationDampers         uint8 `json:"interval_time_ventilation_dampers"`
	VentilationPeriodAfterCO2ReductionTime uint8 `json:"ventilation_period_after_co_2_reduction_time"`
	ForAll                                 bool  `json:"for_all"`
}

type CommandDataVentModuleForAll struct {
	SerialNumber                           int   `json:"serial_number" validate:"required"`
	Delta                                  uint8 `json:"delta" validate:"required"`
	TypeRegulation                         uint8 `json:"type_regulation" validate:"required"`
	IntervalTimeVentilationDampers         uint8 `json:"interval_time_ventilation_dampers" validate:"required"`
	VentilationPeriodAfterCO2ReductionTime uint8 `json:"ventilation_period_after_co_2_reduction_time" validate:"required"`
}

type CommandDataVentByZone struct {
	SerialNumber int   `json:"serial_number" validate:"required"`
	Zone         int16 `json:"zone"  validate:"required"`
}

type CommandHysteresisOnHumidityModule struct {
	SerialNumber int     `json:"serial_number" validate:"required"`
	Zone         int32   `json:"zone"`
	Humidity     float32 `json:"humidity"`
	Hysteresis   uint8   `json:"hysteresis"`
}

//func (c *CommandDataVentModule) Converter() DataVentModule {
//
//	return DataVentModule{
//		Zone: c.Zone,
//		VentSpeed: c.VentSpeed,
//		Delta: c.Delta,
//		TypeRegulation: c.TypeRegulation,
//		IntervalTimeVentilationDampers: c.IntervalTimeVentilationDampers,
//		VentilationPeriodAfterCO2ReductionTime: c.VentilationPeriodAfterCO2ReductionTime,
//
//	}
//}

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Token   string      `json:"token,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type DataCommandTemperature struct {
	Zone        int32   `json:"zone"`
	TempAir     float32 `json:"temp_air"`
	HumidityAir int     `json:"humidity_air"`
	Tempfloor   float32 `json:"tempfloor"`
	CO2         float32 `json:"co_2"`
}

func (c *DataCommandTemperature) GetType() int {
	return CommandTemperatureIn
}

type DataCommandHumidityModule struct {
	Zone       int32   `json:"zone"`
	Setpoint   float32 `json:"setpoint"`
	Hysteresis uint8   `json:"hysteresis"`
}

type DataCommandTemperatureBySensor struct {
	Zone              int32   `json:"zone"`
	SetpointValueTemp float32 `json:"setpoint_value_temp"`
	TypeRegulation    uint8   `json:"type_regulation"`
}

func (c *DataCommandTemperatureBySensor) GetType() int {
	return CommandTemperatureBySensorIn
}

type DataVent struct {
	Zone      int16 `json:"zone"`
	VentSpeed int16 `json:"vent_speed"`
}

func (c *DataVent) GetType() int {
	return DataVentIn
}

type DataVentModule struct {
	Zone                                   int16 `json:"zone"`
	VentSpeed                              int16 `json:"vent_speed"`
	Delta                                  uint8 `json:"delta"`
	TypeRegulation                         uint8 `json:"type_regulation"`
	IntervalTimeVentilationDampers         uint8 `json:"interval_time_ventilation_dampers"`
	VentilationPeriodAfterCO2ReductionTime uint8 `json:"ventilation_period_after_co_2_reduction_time"`
}

func (c *DataVentModule) GetType() int {
	return DataModuleVentIn
}

type Command struct {
	TypeCommand int             `json:"command,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
}

func NewCommand(msg interface{}, types int) (*Command, error) {

	var typeCommand int

	switch types {
	case 1: //исходящие от устройства команды
		switch msg.(type) {
		case DataCommandTemperature:
			typeCommand = 1 // 1
		case []DataCommandTemperatureBySensor:
			typeCommand = 2 // 0
		case DataVent:
			typeCommand = 3 // 3
		case []DataVent:
			typeCommand = 3 // 3
		case DataVentModule:
			typeCommand = 7
		case DataCommandHumidityModule:
			typeCommand = 8
		default:
			return nil, errors.New("not found command")
		}
	case 2: //входящие в устройство команды
		switch msg.(type) {
		case CommandTemperature:
			typeCommand = 1 //2.1
		case CommandTemperatureBySensor:
			typeCommand = 2 // 2.2
		case CommandDataVentModule:
			typeCommand = 3 // 4
		case CommandDataVentByZone:
			typeCommand = 6
		case CommandDataVentModuleForAll:
			typeCommand = 5
		case CommandHysteresisOnHumidityModule:
			typeCommand = 9
		default:
			return nil, errors.New("not found command")
		}
	default:
		return nil, errors.New("not found types fro command")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return &Command{
		typeCommand,
		data}, nil
}

//func (c *Command) MarshalJSON() ([]byte, error)  {
//	return json.Marshal(*c)
//}
//

//func (c *Command) UnmarshalJSON(data []byte) error  {
//	msg := struct {
//		TypeCommand int             `json:"command,omitempty"`
//		Data        json.RawMessage `json:"data,omitempty"`
//	}{}
//
//	err := json.Unmarshal(data, &msg)
//
//	if err != nil {
//		return err
//	}
//
//	c.
//
//	return json.Unmarshal(data, c)
//}
