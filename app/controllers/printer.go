package controllers

import (
	"github.com/revel/revel"
	"github.com/muka/go-bluetooth/bluez/profile/obex"
	"github.com/muka/go-bluetooth/api"
	"errors"
	"time"
)

var (
	temp       = 1
	obexClient = obex.NewObexClient1()
)

type Printer struct {
	*revel.Controller
}

type PrintJob struct {
}

func (c *Printer) Index() revel.Result {
	revel.AppLog.Error("Hello!", "temp", temp)
	temp++
	return c.Render()
}

func (c *Printer) PrintImage(deviceAddress string) revel.Result {
	filepath := "/home/mr/Pictures/who-is-awesome.jpg"
	log := revel.AppLog
	log.Error("Hello!", "temp", temp)

	go sendFile(deviceAddress, filepath)
	//if err != nil {
	//	return c.RenderError(err)
	//}
	return c.Render()
}

func sendFile(targetAddress string, filePath string) error {
	log := revel.AppLog
	log.Info("sendFile", "targetAddress", targetAddress, "filePath", filePath)
	dev, err := api.GetDeviceByAddress(targetAddress)
	if err != nil {
		panic(err)
	}
	log.Debug("device (dev): ", "dev", dev)

	if dev == nil {
		panic("Device not found")
	}

	props, err := dev.GetProperties()
	if !props.Paired {
		log.Debug("not paired")

		err = dev.Pair()
		if err != nil {
			log.Error(err.Error())
			return err
		}

	} else {
		log.Debug("already paired")
	}

	sessionArgs := map[string]interface{}{}
	sessionArgs["Target"] = "opp"

	tries := 1
	maxRetry := 20
	var sessionPath string
	for tries < maxRetry {
		log.Debug("Create Session...")
		sessionPath, err = obexClient.CreateSession(targetAddress, sessionArgs)
		if err == nil {
			break
		}

		tries++

		if err != nil {
			log.Error(err.Error())
			obexClient.GetClient().Connect()
		}
	}
	if tries >= maxRetry {
		//log.Fatal("Max tries reached")
		return errors.New("Max tries reached")
	}

	log.Debug("Session created: ", "sessionPath", sessionPath)

	obexSession := obex.NewObexSession1(sessionPath)
	sessionProps, err := obexSession.GetProperties()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	log.Debug("Source		: ", "Source		", sessionProps.Source)
	log.Debug("Destination	: ", "Destination	", sessionProps.Destination)
	log.Debug("Channel		: ", "Channel		", sessionProps.Channel)
	log.Debug("Target		: ", "Target		", sessionProps.Target)
	log.Debug("Root			: ", "Root			", sessionProps.Root)

	log.Debug("Init transmission on ", "sessionPath", sessionPath)
	obexObjectPush := obex.NewObjectPush1(sessionPath)
	log.Debug("Send File: ", "filePath", filePath)

	obexObjectPush.SendFile(filePath)
	transPath, transProps, err := obexObjectPush.SendFile(filePath)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	log.Debug("Transmission initiated: ", "Transmission ", transPath)
	log.Debug("Status      : ", "Status      ", transProps.Status)
	log.Debug("Session     : ", "Session     ", transProps.Session)
	log.Debug("Name        : ", "Name        ", transProps.Name)
	log.Debug("Type        : ", "Type        ", transProps.Type)
	log.Debug("Time        : ", "Time        ", transProps.Time)
	log.Debug("Size        : ", "Size        ", transProps.Size)
	log.Debug("Transferred : ", "Transferred ", transProps.Transferred)
	log.Debug("Filename    : ", "Filename    ", transProps.Filename)

	for transProps.Transferred < transProps.Size {
		time.Sleep(1 * time.Second)

		obexTransfer := obex.NewObexTransfer1(transPath)
		transProps, err = obexTransfer.GetProperties()
		if err != nil {
			log.Error(err.Error())
			//return err
		}
		transferedPercent := (100 / float64(transProps.Size)) * float64(transProps.Transferred)

		log.Debug("Progress    : ", "Progress", transferedPercent)
	}

	log.Debug(sessionPath)

	obexClient.RemoveSession(sessionPath)

	return nil
}
