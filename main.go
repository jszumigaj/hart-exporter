package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jszumigaj/hart"
	"github.com/jszumigaj/hart/device"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	serialPort = *flag.String("serial", "COM1", "Serial port name ex. COM1")
	listenPort = *flag.Int("listen", 3333, "Listening port ex. 8000.")
)

func main() {
	flag.Parse()
	log.Printf("Using serial port %s", serialPort)

	serial, err := hart.Open(serialPort)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	defer serial.Close()

	master := hart.NewCommandExecutor(serial)
	var device device.UniversalDevice
	commands := []hart.Command{
		device.Command1(),
		device.Command2(),
		device.Command3(),
	}
	executed := make(chan hart.Command)
	go executeCommands(master, &device, commands, executed)
	go displayResults(&device, executed)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/hart", hartHandler(&device))

	log.Printf("Listening on %d", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", listenPort), nil))

}

func executeCommands(master hart.CommandExecutor, device *device.UniversalDevice, commands []hart.Command, executed chan<- hart.Command) error {

	// identification
	command0 := device.Command0()
	_, err := master.Execute(command0)
	if err != nil {
		log.Println(err)
		close(executed)
		return nil
	}

	executed <- command0

	for {
		for _, c := range commands {
			_, err = master.Execute(c)
			if err == nil {
				executed <- c
			} else {
				log.Println(err)
				panic(err)
			}

			time.Sleep(2 * time.Second)
		}
	}
}

func displayResults(device *device.UniversalDevice, executed <-chan hart.Command) {
	for cmd := range executed {
		log.Println("Command status:", cmd.Status())
		log.Println("Device status:", device.Status())

		switch cmd.No() {
		case 0:
			log.Printf("Device %v", device)
			manid := fmt.Sprintf("%x", device.ManufacturerId())
			devtype := fmt.Sprintf("%x", device.MfrsDeviceType())
			devid := fmt.Sprintf("%07d", device.Id())
			deviceInfoGauge.WithLabelValues(manid, devtype, devid).Set(1)
		case 1:
			v, u := device.PV()
			log.Printf("PV: %v [%v]\n", v, u)
			pvGauge.WithLabelValues(fmt.Sprint(u)).Set(float64(v))
		case 2:
			log.Printf("Current: %v [mA]\n", device.Current())
			log.Printf("PoR: %v [%%]\n", device.PercentOfRange())
			currentGauge.Set(float64(device.Current()))
			porGauge.Set(float64(device.PercentOfRange()))
		case 3:
			v, u := device.SV()
			log.Printf("SV: %v [%v]\n", v, u)
			svGauge.WithLabelValues(fmt.Sprint(u)).Set(float64(v))
			v, u = device.TV()
			log.Printf("TV: %v [%v]\n", v, u)
			tvGauge.WithLabelValues(fmt.Sprint(u)).Set(float64(v))
			v, u = device.FV()
			log.Printf("FV: %v [%v]\n", v, u)
			fvGauge.WithLabelValues(fmt.Sprint(u)).Set(float64(v))
		}
	}
}

func hartHandler(device *device.UniversalDevice) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		json.NewEncoder(w).Encode(device)
	})
}
