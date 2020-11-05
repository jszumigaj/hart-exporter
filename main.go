package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jszumigaj/hart"
	"github.com/jszumigaj/hart/univrsl"
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

	master := hart.NewMaster(serial)
	device := &univrsl.Device{}
	commands := []hart.Command{
		&univrsl.Command1{},
		&univrsl.Command2{},
		&univrsl.Command3{},
	}
	executed := make(chan hart.Command)
	go executeCommands(master, device, commands, executed)
	go displayResults(device, executed)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/hart", hartHandler(device))

	log.Printf("Listening on %d", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", listenPort), nil))

}

func executeCommands(master *hart.Master, device *univrsl.Device, commands []hart.Command, executed chan<- hart.Command) error {

	// identification
	command0 := device.Command0()
	_, err := master.Execute(command0, device)
	if err != nil {
		log.Println(err)
		close(executed)
		return nil
	}

	executed <- command0

	for {
		for _, c := range commands {
			_, err = master.Execute(c, device)
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

func displayResults(device *univrsl.Device, executed <-chan hart.Command) {
	for command := range executed {
		log.Println("Command status:", command.Status())
		log.Println("Device status:", device.Status())

		if _, ok := command.(*univrsl.Command0); ok {
			log.Printf("Device %v", device)
			manid := fmt.Sprintf("%x", device.ManufacturerId())
			devtype := fmt.Sprintf("%x", device.MfrsDeviceType())
			devid := fmt.Sprintf("%07d", device.Id())
			deviceInfoGauge.WithLabelValues(manid, devtype, devid).Set(1)
		} else if cmd, ok := command.(*univrsl.Command1); ok {
			v, u := cmd.PV()
			log.Printf("PV: %v [%v]\n", v, u)
			pvGauge.WithLabelValues(fmt.Sprint(u)).Set(float64(v))
		} else if cmd, ok := command.(*univrsl.Command2); ok {
			log.Printf("Current: %v [mA]\n", cmd.Current())
			log.Printf("PoR: %v [%%]\n", cmd.PercentOfRange())
			currentGauge.Set(float64(cmd.Current()))
			porGauge.Set(float64(cmd.PercentOfRange()))
		} else if cmd, ok := command.(*univrsl.Command3); ok {
			v, u := cmd.SV()
			log.Printf("SV: %v [%v]\n", v, u)
			svGauge.WithLabelValues(fmt.Sprint(u)).Set(float64(v))
			v, u = cmd.TV()
			log.Printf("TV: %v [%v]\n", v, u)
			tvGauge.WithLabelValues(fmt.Sprint(u)).Set(float64(v))
			v, u = cmd.FV()
			log.Printf("FV: %v [%v]\n", v, u)
			fvGauge.WithLabelValues(fmt.Sprint(u)).Set(float64(v))
		}
	}
}

func hartHandler(device *univrsl.Device) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		json.NewEncoder(w).Encode(device)
	})
}
