package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	deviceInfoGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "device_info",
		Help: "HART device info",
	}, []string{"ManufacturerId", "DeviceType", "DeviceId"})

	deviceInfo13Gauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "device_info_cmd13",
		Help: "HART command 13 output",
	}, []string{"Tag", "Descriptor", "Date"})

	deviceStatusCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "device_status",
		Help: "HART field device status counter",
	}, []string{"Status"})

	commandStatusCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "command_status",
		Help: "HART command status counter",
	}, []string{"Status"})

	hartCommErrorsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "hart_communication_errors_count",
		Help: "HART communications errors total counter",
	}, []string{"Error"})

	pvGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "device_pv_value",
		Help: "A device PV float32 value.",
	}, []string{"unit"})

	svGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "device_sv_value",
		Help: "A device SV float32 value.",
	}, []string{"unit"})

	tvGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "device_tv_value",
		Help: "A device TV float32 value.",
	}, []string{"unit"})

	fvGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "device_fv_value",
		Help: "A device FV float32 value.",
	}, []string{"unit"})

	currentGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "device_current_value",
		Help: "A device current output float32 value in mA.",
	})

	porGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "device_percent_of_range_value",
		Help: "A device Percent of Range float32 value in %.",
	})

	lrvGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "device_lower_range_value",
		Help: "A device Lower Range Value (LRV) float32 value.",
	}, []string{"unit"})

	urvGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "device_upper_range_value",
		Help: "A device Upper Range Value (URV) float32 value.",
	}, []string{"unit"})

	dampingGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "device_damping_value",
		Help: "A device Damping float32 value.",
	})
)

func init() {
	
	var appInfoGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "app_info",
		Help: "Application info",
	}, []string{"version"})

	appInfoGauge.WithLabelValues("0.3.0").Set(1)
}
