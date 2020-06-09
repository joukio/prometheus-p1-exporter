package main

import (
        "bufio"
        "fmt"
        "io"
        "log"
        "net/http"
        "os"
        "strconv"
        "strings"
        "time"

        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
        "github.com/tarm/serial"
)

var (
        reader         *bufio.Reader
        powerDelivered = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "power_delivered_watts",
                Help: "Actual electricity power delivered (+P) in 1 Watt resolution",
        })

        powerReceived = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "power_received_watts",
                Help: "Actual electricity power received (-P) in 1 Watt resolution",
        })

        powerToTariff1 = prometheus.NewCounterFunc(
                prometheus.CounterOpts{
                        Name: "power_to_tariff1_kwh",
                        Help: "Meter Reading electricity delivered to client (Tariff 1) in 0,001 kWh",
                },
                func() float64 {
                        // fmt.Println("reading powerTariff1Meter")
                        return powerToTariff1Meter
                },
        )

        powerToTariff2 = prometheus.NewCounterFunc(
                prometheus.CounterOpts{
                        Name: "power_to_tariff2_kwh",
                        Help: "Meter Reading electricity delivered to client (Tariff 2) in 0,001 kWh",
                },
                func() float64 {
                        // fmt.Println("reading powerTariff2Meter")
                        return powerToTariff2Meter
                },
        )

        powerByTariff1 = prometheus.NewCounterFunc(
                prometheus.CounterOpts{
                        Name: "power_by_tariff1_kwh",
                        Help: "Meter Reading electricity delivered by client (Tariff 1) in 0,001 kWh",
                },
                func() float64 {
                        // fmt.Println("reading powerTariff1Meter")
                        return powerByTariff1Meter
                },
        )

        powerByTariff2 = prometheus.NewCounterFunc(
                prometheus.CounterOpts{
                        Name: "power_by_tariff2_kwh",
                        Help: "Meter Reading electricity delivered by client (Tariff 2) in 0,001 kWh",
                },
                func() float64 {
                        // fmt.Println("reading powerTariff2Meter")
                        return powerByTariff2Meter
                },
        )

        gasMeter = prometheus.NewCounterFunc(
                prometheus.CounterOpts{
                        Name: "gas_meter_m3",
                        Help: "Gas meter reading in m3",
                },
                func() float64 {
                        return gasTotalMeter
                },
        )

        instVoltL1 = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "inst_volt_l1",
                Help: "Instantaneous voltage L1 in V resolution",
        },
        )

        instVoltL2 = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "inst_volt_l2",
                Help: "Instantaneous voltage L2 in V resolution",
        },
        )

        instVoltL3 = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "inst_volt_l3",
                Help: "Instantaneous voltage L3 in V resolution",
        },
        )

        instPlusPowerL1 = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "inst_pluspower_l1",
                Help: "Instantaneous active power L1 (+P)in W resolution",
        },
        )

        instPlusPowerL2 = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "inst_pluspower_l2",
                Help: "Instantaneous active power L2 (+P)in W resolution",
        },
        )

        instPlusPowerL3 = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "inst_pluspower_l3",
                Help: "Instantaneous active power L3 (+P)in W resolution",
        },
        )

        instNegPowerL1 = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "inst_negpower_l1",
                Help: "Instantaneous active power L1 (-P)in W resolution",
        },
        )

        instNegPowerL2 = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "inst_negpower_l2",
                Help: "Instantaneous active power L2 (-P)in W resolution",
        },
        )

        instNegPowerL3 = prometheus.NewGauge(prometheus.GaugeOpts{
                Name: "inst_negpower_l3",
                Help: "Instantaneous active power L3 (-P)in W resolution",
        },
        )

        powerToTariff1Meter float64
        powerToTariff2Meter float64
        powerByTariff1Meter float64
        powerByTariff2Meter float64
        gasTotalMeter       float64
)

func init() {
        // Metrics have to be registered to be exposed:
        prometheus.MustRegister(powerDelivered)
        prometheus.MustRegister(powerReceived)
        prometheus.MustRegister(powerToTariff1)
        prometheus.MustRegister(powerToTariff2)
        prometheus.MustRegister(powerByTariff1)
        prometheus.MustRegister(powerByTariff2)
        prometheus.MustRegister(gasMeter)
        prometheus.MustRegister(instVoltL1)
        prometheus.MustRegister(instVoltL2)
        prometheus.MustRegister(instVoltL3)
        prometheus.MustRegister(instPlusPowerL1)
        prometheus.MustRegister(instPlusPowerL2)
        prometheus.MustRegister(instPlusPowerL3)
        prometheus.MustRegister(instNegPowerL1)
        prometheus.MustRegister(instNegPowerL2)
        prometheus.MustRegister(instNegPowerL3)
}

func main() {
        if os.Getenv("SERIAL_DEVICE") != "" {
                fmt.Println("gonna use serial device")
                config := &serial.Config{Name: os.Getenv("SERIAL_DEVICE"), Baud: 115200}

                usb, err := serial.OpenPort(config)
                if err != nil {
                        fmt.Printf("Could not open serial interface: %s", err)
                        return
                }

                reader = bufio.NewReader(usb)
        } else {
                fmt.Println("gonna use some files")
                file, err := os.Open("examples/fulllist.txt")
                if err != nil {
                        fmt.Println(err)
                        return
                }
                defer file.Close()
                reader = bufio.NewReader(file)
        }

        go listener(reader)

        // sleeping 2 seconds to prevent uninitialized scrapes
        time.Sleep(2 * time.Second)

        fmt.Println("now serving metrics")
        http.Handle("/metrics", promhttp.Handler())
        log.Fatal(http.ListenAndServe(":9222", nil))

}

func listener(source io.Reader) {
        var line string
        for {
                rawLine, err := reader.ReadBytes('\x0a')
                if err != nil {
                        fmt.Println(err)
                        return
                }
                line = string(rawLine[:])
                if strings.HasPrefix(line, "1-0:1.8.1") {
                        tmpVal, err := strconv.ParseFloat(line[10:20], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        powerToTariff1Meter = tmpVal
                } else if strings.HasPrefix(line, "1-0:1.8.2") {
                        tmpVal, err := strconv.ParseFloat(line[10:20], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        powerToTariff2Meter = tmpVal
                } else if strings.HasPrefix(line, "1-0:2.8.1") {
                        tmpVal, err := strconv.ParseFloat(line[10:20], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        powerByTariff1Meter = tmpVal
                } else if strings.HasPrefix(line, "1-0:2.8.2") {
                        tmpVal, err := strconv.ParseFloat(line[10:20], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        powerByTariff2Meter = tmpVal
                } else if strings.HasPrefix(line, "0-1:24.2.1") {
                        tmpVal, err := strconv.ParseFloat(line[26:35], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        gasTotalMeter = tmpVal // m3
                } else if strings.HasPrefix(line, "1-0:1.7.0") {
                        tmpVal, err := strconv.ParseFloat(line[10:16], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        powerDelivered.Set(tmpVal * 1000)
                } else if strings.HasPrefix(line, "1-0:32.7.0") {
                        tmpVal, err := strconv.ParseFloat(line[11:16], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        instVoltL1.Set(tmpVal)
                } else if strings.HasPrefix(line, "1-0:52.7.0") {
                        tmpVal, err := strconv.ParseFloat(line[11:16], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        instVoltL2.Set(tmpVal)
                } else if strings.HasPrefix(line, "1-0:72.7.0") {
                        tmpVal, err := strconv.ParseFloat(line[11:16], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        instVoltL3.Set(tmpVal)
                } else if strings.HasPrefix(line, "1-0:21.7.0") {
                        tmpVal, err := strconv.ParseFloat(line[11:17], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        instPlusPowerL1.Set(tmpVal * 1000)
                } else if strings.HasPrefix(line, "1-0:41.7.0") {
                        tmpVal, err := strconv.ParseFloat(line[11:17], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        instPlusPowerL2.Set(tmpVal * 1000)
                } else if strings.HasPrefix(line, "1-0:61.7.0") {
                        tmpVal, err := strconv.ParseFloat(line[11:17], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        instPlusPowerL3.Set(tmpVal * 1000)
                } else if strings.HasPrefix(line, "1-0:22.7.0") {
                        tmpVal, err := strconv.ParseFloat(line[11:17], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        instNegPowerL1.Set(tmpVal * 1000)
                } else if strings.HasPrefix(line, "1-0:42.7.0") {
                        tmpVal, err := strconv.ParseFloat(line[11:17], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        instNegPowerL2.Set(tmpVal * 1000)
                } else if strings.HasPrefix(line, "1-0:62.7.0") {
                        tmpVal, err := strconv.ParseFloat(line[11:17], 64)
                        if err != nil {
                                fmt.Println(err)
                                continue
                        }
                        instNegPowerL3.Set(tmpVal * 1000)
                }
                if os.Getenv("SERIAL_DEVICE") == "" {
                        time.Sleep(200 * time.Millisecond)
                }
        }
}
