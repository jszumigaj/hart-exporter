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
	serialPort = flag.String("c", "COM1", "Serial port name.")
	listenPort = flag.Int("l", 9090, "Listening port.")
	delay      = flag.Int("d", 10, "Delay between each commands set execution in seconds.")
)

func main() {
	flag.Parse()
	log.Printf("Start reading commands every %d sec", *delay)
	log.Printf("Using serial port %s", *serialPort)

	serial, err := hart.Open(*serialPort)
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
	defer serial.Close()

	master := hart.NewMaster(serial)
	device := &univrsl.Device{}
	commands := []hart.Command{
		device.Command0(),
		&univrsl.Command1{},
		&univrsl.Command2{},
		&univrsl.Command3{},
		&univrsl.Command13{},
		&univrsl.Command15{},
	}
	executed := make(chan hart.Command)
	go executeCommands(master, device, commands, executed)
	go displayResults(device, executed)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/hart", hartHandler(commands))

	log.Printf("Listening on %d", *listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *listenPort), nil))
}

func executeCommands(master *hart.Master, device *univrsl.Device, commands []hart.Command, executed chan<- hart.Command) error {
	for {
		start := time.Now()
		for _, cmd := range commands {

			if _, err := master.Execute(cmd, device); err != nil {
				log.Println(cmd.Description(), "error:", err)
			} else {
				executed <- cmd
			}

		}
		elapsed := time.Now().Sub(start)
		time.Sleep(time.Duration(*delay) * time.Second - elapsed)
	}
}

func displayResults(device *univrsl.Device, executed <-chan hart.Command) {
	for command := range executed {
		log.Println("Command status:", command.Status())
		log.Println("Device status:", device.Status())

		if _, ok := command.(*univrsl.Command0); ok {
			log.Printf("Cmd #0: Device %v", device)
			manid := fmt.Sprintf("%x", device.ManufacturerId())
			devtype := fmt.Sprintf("%x", device.MfrsDeviceType())
			devid := fmt.Sprintf("%07d", device.Id())
			deviceInfoGauge.WithLabelValues(manid, devtype, devid).Set(1)

		} else if cmd, ok := command.(*univrsl.Command1); ok {
			log.Printf("Cmd #1: PV = %v [%v]\n", cmd.PV, cmd.Unit)
			pvGauge.WithLabelValues(fmt.Sprint(cmd.Unit)).Set(float64(cmd.PV))

		} else if cmd, ok := command.(*univrsl.Command2); ok {
			log.Printf("Cmd #2: Current = %v [mA]\n", cmd.Current)
			log.Printf("Cmd #2: PoR = %v [%%]\n", cmd.PercentOfRange)
			currentGauge.Set(float64(cmd.Current))
			porGauge.Set(float64(cmd.PercentOfRange))

		} else if cmd, ok := command.(*univrsl.Command3); ok {
			log.Printf("Cmd #3: SV = %v [%v]\n", cmd.Sv, cmd.SvUnit)
			svGauge.WithLabelValues(fmt.Sprint(cmd.SvUnit)).Set(float64(cmd.Sv))
			log.Printf("Cmd #3: TV = %v [%v]\n", cmd.Tv, cmd.TvUnit)
			tvGauge.WithLabelValues(fmt.Sprint(cmd.TvUnit)).Set(float64(cmd.Tv))
			log.Printf("Cmd #3: FV = %v [%v]\n", cmd.Fv, cmd.FvUnit)
			fvGauge.WithLabelValues(fmt.Sprint(cmd.FvUnit)).Set(float64(cmd.Tv))

		} else if cmd, ok := command.(*univrsl.Command13); ok {
			log.Printf("Cmd #13: Tag: %v", cmd.Tag)
			log.Printf("Cmd #13: Descriptor: %v", cmd.Descriptor)
			log.Printf("Cmd #13: Date: %v", cmd.Date.Format("2006-01-02"))
			deviceInfo13Gauge.WithLabelValues(cmd.Tag, cmd.Descriptor, cmd.Date.Format("2006-01-02")).Set(1)

		} else if cmd, ok := command.(*univrsl.Command15); ok {
			log.Printf("Cmd #15: %v", cmd)
			log.Printf("Cmd #15: LRV = %v [%v]\n", cmd.LowerRangeValue, cmd.UpperAndLowerRangeValuesUnit)
			log.Printf("Cmd #15: URV = %v [%v]\n", cmd.UpperRangeValue, cmd.UpperAndLowerRangeValuesUnit)

		}
	}
}

// Handle http request by serializing data
func hartHandler(data ...interface{}) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		json.NewEncoder(w).Encode(data)
	})
}
