package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jszumigaj/hart"
	"github.com/jszumigaj/hart/status"
	"github.com/jszumigaj/hart/univrsl"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	serialPortName = flag.String("c", "COM1", "Serial port name.")
	listenPort = flag.Int("l", 9090, "Listening port.")
	delay      = flag.Int("d", 10, "Delay between each commands set execution in seconds.")
)

func main() {
	flag.Parse()
	log.Printf("Start reading commands every %d sec", *delay)
	log.Printf("Using serial port %s", *serialPortName)

	serial, err := hart.Open(*serialPortName)
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
				incrementHartCommErrorsCounter(err)
			} else {
				executed <- cmd
			}

		}
		elapsed := time.Now().Sub(start)
		time.Sleep(time.Duration(*delay)*time.Second - elapsed)
	}
}

// This funct increments errors counters by name. In case flags errors each flag will be counted separatelly.
func incrementHartCommErrorsCounter(err error) {
	// if this is Communications error flags they should be count spearate by each flags, not together.
	if e, ok := err.(status.CommunicationsErrorSummaryFlags); ok {
		for i := 0; i < 8; i++ {
			mask := status.CommunicationsErrorSummaryFlags(1 << i)
			if e.HasFlag(mask) {
				hartCommErrorsCounter.WithLabelValues(mask.Error()).Inc()
			}
		}
		return
	}
		
	// this is not comm error flag
	hartCommErrorsCounter.WithLabelValues(err.Error()).Inc()
}

func displayResults(device *univrsl.Device, executed <-chan hart.Command) {
	for command := range executed {
		log.Println("Command status:", command.Status())
		log.Println("Device status:", device.Status())
		deviceStatusCounter.WithLabelValues(device.Status().String()).Inc()
		commandStatusCounter.WithLabelValues(command.Status().String()).Inc()

		switch cmd := command.(type) {
		case *univrsl.Command0:
			log.Printf("Cmd #0: Device %v", device)
			manid := fmt.Sprintf("%x", device.ManufacturerId())
			devtype := fmt.Sprintf("%x", device.MfrsDeviceType())
			devid := fmt.Sprintf("%07d", device.Id())
			deviceInfoGauge.WithLabelValues(manid, devtype, devid).Set(1)

		case *univrsl.Command1:
			log.Printf("Cmd #1: PV = %v [%v]\n", cmd.PV, cmd.Unit)
			pvGauge.WithLabelValues(cmd.Unit.String()).Set(float64(cmd.PV))

		case *univrsl.Command2:
			log.Printf("Cmd #2: Current = %v [mA]\n", cmd.Current)
			log.Printf("Cmd #2: PoR = %v [%%]\n", cmd.PercentOfRange)
			currentGauge.Set(float64(cmd.Current))
			porGauge.Set(float64(cmd.PercentOfRange))

		case *univrsl.Command3:
			log.Printf("Cmd #3: SV = %v [%v]\n", cmd.Sv, cmd.SvUnit)
			svGauge.WithLabelValues(cmd.SvUnit.String()).Set(float64(cmd.Sv))
			log.Printf("Cmd #3: TV = %v [%v]\n", cmd.Tv, cmd.TvUnit)
			tvGauge.WithLabelValues(cmd.TvUnit.String()).Set(float64(cmd.Tv))
			log.Printf("Cmd #3: FV = %v [%v]\n", cmd.Fv, cmd.FvUnit)
			fvGauge.WithLabelValues(cmd.FvUnit.String()).Set(float64(cmd.Tv))

		case *univrsl.Command13:
			log.Printf("Cmd #13: Tag: %v", cmd.Tag)
			log.Printf("Cmd #13: Descriptor: %v", cmd.Descriptor)
			log.Printf("Cmd #13: Date: %v", cmd.Date.Format("2006-01-02"))
			deviceInfo13Gauge.WithLabelValues(cmd.Tag, cmd.Descriptor, cmd.Date.Format("2006-01-02")).Set(1)

		case *univrsl.Command15:
			unit := cmd.UpperAndLowerRangeValuesUnit.String()
			log.Printf("Cmd #15: %v", cmd)
			log.Printf("Cmd #15: Range = %v ... %v [%v]\n", cmd.LowerRangeValue, cmd.UpperRangeValue, unit)
			log.Printf("Cmd #15: Damping = %v [s]\n", cmd.Damping)
			lrvGauge.WithLabelValues(unit).Set(float64(cmd.LowerRangeValue))
			urvGauge.WithLabelValues(unit).Set(float64(cmd.UpperRangeValue))
			dampingGauge.Set(float64(cmd.Damping))
		}
	}
}

// Handle http request by serializing data
func hartHandler(data ...interface{}) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		json.NewEncoder(w).Encode(data)
	})
}
