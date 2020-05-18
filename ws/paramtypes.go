package ws

import (
	"encoding/json"
	"errors"
	"net"

	"github.com/bob620/baka-rpc-go/parameters"
)

type IPNetParam struct {
	Name       string
	Default    []net.IPNet
	IsRequired bool
	data       json.RawMessage
}

func (param *IPNetParam) Clone(data json.RawMessage) (parameters.Param, error) {
	clone := IPNetParam{param.Name, param.Default, param.IsRequired, param.data}
	if data != nil {
		err := clone.SetData(data)
		if err != nil {
			return nil, err
		}
	}

	if data == nil && param.IsRequired {
		return nil, errors.New("value needed")
	}
	return &clone, nil
}

func (param *IPNetParam) SetName(newName string) {
	param.Name = newName
}

func (param *IPNetParam) GetName() string {
	return param.Name
}

func (param *IPNetParam) SetData(message json.RawMessage) (err error) {
	param.data = message
	_, err = param.GetIPNet()
	return
}

func (param *IPNetParam) GetData() json.RawMessage {
	if param.data == nil {
		data, _ := json.Marshal(param.Default)
		return data
	}
	return param.data
}

func (param *IPNetParam) GetIPNet() (value []net.IPNet, err error) {
	err = json.Unmarshal(param.GetData(), &value)
	return
}

func (param *IPNetParam) MarshalJSON() ([]byte, error) {
	if param.data == nil {
		data, err := json.Marshal(param.Default)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return param.data, nil
}

func (param *IPNetParam) UnmarshalJSON(jsonData []byte) (err error) {
	data := param.Default
	err = json.Unmarshal(jsonData, data)
	return
}
