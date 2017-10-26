package controllers

import (
	"github.com/revel/revel"
	"time"
	"github.com/muka/go-bluetooth/api"
)

var adapterId = "hci0"

type DeviceManager struct {
	*revel.Controller
}

type Stuff struct {
	Address string ` json:"address" xml:"address" `
	Name    string ` json:"name" xml:"name" `
}

func (c DeviceManager) ListDevices() revel.Result {
	data := make(map[string]interface{})
	devices, err := fetchDevices()
	data["error"] = err
	data["devices"] = devices
	return c.RenderJSON(data)
}

func fetchDevices() ([]Stuff, error) {
	adapterExists, err := api.AdapterExists(adapterId)
	if err != nil || adapterExists == false {
		return nil, err
	}

	err = api.StartDiscoveryOn(adapterId)
	if err != nil && err.Error() != "Operation already in progress" {
		return nil, err
	}

	time.Sleep(time.Second)
	devices, err := api.GetDevices()
	if err != nil {
		return nil, err
	}

	foundTargets := make([]Stuff, len(devices))

	for i, aDevice := range devices {
		deviceProps, err := aDevice.GetProperties()

		if err != nil {
			return nil, err
		}

		foundTargets[i] = Stuff{
			Address: deviceProps.Address,
			Name:    deviceProps.Name,
		}
	}

	return foundTargets, nil
}
