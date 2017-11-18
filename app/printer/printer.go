package printer

import (
	"container/list"
	"github.com/revel/revel"
	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/bluez/profile/obex"
	"time"
	"errors"
)

func init() {
	go printer()
}

var (
	images     = make(chan string)
	obexClient *obex.ObexClient1
	targetDevice = "98:4E:97:00:3F:3C"

)

func AddImage(path string) {
	revel.AppLog.Error("Called AddImage")
	images <- path
}

func printer() {
	transfer := list.New()
	log := revel.AppLog
	obexClient = obex.NewObexClient1()

	log.Error("Printer routine started!")

	for {
		log.Error("Waiting...")

		select {
		case image := <-images:
			log.Error("Received File to send!")

			temp(targetDevice, image)
			transfer.PushBack(image)
		}
	}
}

func connectToDevice(targetAddress string) (string, error) {
	log := revel.AppLog
	log.Infof("Connect to Device", "Target", targetAddress)
	sessionArgs := map[string]interface{}{}
	sessionArgs["Target"] = "opp"

	tries := 1
	maxRetry := 20
	var sessionPath string
	var err error
	for tries < maxRetry {
		log.Debug("Create Session...")
		sessionPath, err = obexClient.CreateSession(targetAddress, sessionArgs)
		if err == nil {
			return sessionPath, nil
		}

		tries++
		log.Error(err.Error())
	}

	//log.Fatal("Max tries reached")
	return "", errors.New("Max tries reached")
}

func temp(targetAddress string, filePath string) error {
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

	sessionPath, err := connectToDevice(targetAddress)
	//defer obexClient.RemoveSession(sessionPath)

	if err != nil {
		return err
	}

	log.Debug("Session created: ", "sessionPath", sessionPath)

	printSession(sessionPath)
	sendFile(sessionPath, filePath)

	return nil
}

func printSession(sessionPath string) error {
	log := revel.AppLog

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

	log.Debug(sessionPath)

	return nil
}

func sendFile(sessionPath string, filePath string) error {
	log := revel.AppLog

	obexObjectPush := obex.NewObjectPush1(sessionPath)
	log.Debug("Send File: ", "filePath", filePath)

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

	return nil
}
