package main

import (
	"fmt"
	"os"
	"time"

	_ "github.com/jnovack/go-version"
	mqttPublisher "github.com/jnovack/nut-to-mqtt/pkg/mqtt_publisher"
	nutClient "github.com/jnovack/nut-to-mqtt/pkg/nut_client"
	"github.com/mattn/go-isatty"
	"github.com/namsral/flag"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	trackedMetrics = map[string]string{
		"battery.charge":  "Battery charge (percent of full)",
		"battery.runtime": "Battery runtime (seconds)",
		"battery.voltage": "Battery voltage (V)",
		"input.voltage":   "Input voltage (V)",
		"ups.load":        "Load on UPS (percent of full)",
		"ups.status":      "UPS status",
	}
)

// This example connects to NUT, authenticates and returns the first UPS listed.
func main() {

	nutClient.Initialize()
	mqttClient := mqttPublisher.Connect()
	nutClient.Connect()

	for {
		if mqttClient.IsConnected() {
			metrics := nutClient.Collect(trackedMetrics)
			for _, obj := range metrics {
				// Add Topic Prefix
				// TODO Permit run-time setting of topic prefix
				obj.Topic = fmt.Sprintf("%s%s", "v1/ups/", obj.Topic)
				log.Debug().Str("topic", obj.Topic).Str("msg", obj.Message).Msg("Found metric")
				token := mqttClient.Publish(obj.Topic, 0, false, obj.Message)
				if !token.WaitTimeout(10 * time.Second) {
					log.Error().Str("topic", obj.Topic).Str("msg", obj.Message).Msg("Error: Timeout sending message")
				}
				if err := token.Error(); err != nil {
					log.Error().Str("topic", obj.Topic).Str("msg", obj.Message).Msgf("Failed to send message")
				}
			}
			// TODO Permit run-time setting of loop delay
			time.Sleep(12 * time.Second)
		}
		time.Sleep(3 * time.Second)
	}

}

func init() {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		// Format using ConsoleWriter if running straight
		zerolog.TimestampFunc = func() time.Time {
			return time.Now().In(time.Local)
		}
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		// Format using JSON if running as a service (or container)
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}

	// TODO Permit run-time setting of LogLevel
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	flag.Parse()
}
