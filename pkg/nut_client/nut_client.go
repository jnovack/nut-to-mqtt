package nutclient

import (
	"fmt"
	"path/filepath"

	secrets "github.com/ijustfool/docker-secrets"
	"github.com/namsral/flag"
	nut "github.com/robbiet480/go.nut"
	"github.com/rs/zerolog/log"
)

// Metric is exported
type Metric struct {
	Topic   string
	Message string
}

var (
	ok               bool
	trackedVariables map[string]string
	// upsStatus defines a map for the ups.status return
	upsStatus = map[string]string{
		"OL":      "Online",
		"OB":      "On Battery",
		"LB":      "Low Battery",
		"HB":      "High Battery",
		"RB":      "Battery Needs Replaced",
		"CHRG":    "Battery Charging",
		"DISCHRG": "Battery Discharging",
		"BYPASS":  "Bypass Active",
		"CAL":     "Runtime Calibration",
		"OFF":     "Offline",
		"OVER":    "Overloaded",
		"TRIM":    "Trimming Voltage",
		"BOOST":   "Boosting Voltage",
		"FSD":     "Forced Shutdown",
	}
)

var (
	client       nut.Client
	hostname     = flag.String("nut_hostname", "127.0.0.1", "nut hostname where nut-server is running on port 3493")
	username     = flag.String("nut_username", "", "username for nut authentication (required)")
	password     = flag.String("nut_password", "", "password for nut authentication (required)")
	passwordFile = flag.String("nut_password_file", "", "path to the 'nut_password' file, which holds the nut_password")
)

func init() {
	log.Logger = log.With().Str("component", "nutclient").Logger()
}

// Initialize hack until I can re-write the NUT library
func Initialize() {
	if *passwordFile != "" {
		dockerSecrets, _ := secrets.NewDockerSecrets(filepath.Dir(*passwordFile))
		secret, _ := dockerSecrets.Get("nut_password")
		*password = secret
	}

	if *username == "" {
		log.Fatal().Msgf("nut_username must be supplied to connect to nut_hostname")
	}
	if *password == "" {
		log.Fatal().Msgf("nut_password must be supplied to connect to nut_hostname")
	}
}

// Connect to the NUT server
func Connect() {
	var err error
	// Validate flags

	client, err = nut.Connect(*hostname)
	if err != nil {
		log.Fatal().Err(err).Str("hostname", *hostname).Msg("Cannot connect to NUT host")
	}
	log.Info().Msgf("Connected to %s", *hostname)

	_, err = client.Authenticate(*username, *password)
	if err != nil {
		log.Fatal().Err(err).Str("username", *username).Msg("unable to Authenticate()")
	}
}

// Collect variables from connection
func Collect(variables map[string]string) []Metric {
	log.Debug().Caller().Msg("Collect()")
	trackedVariables = variables
	upsList, err := client.GetUPSList()
	if err != nil {
		log.Fatal().Err(err).Str("hostname", *hostname).Msg("unable to GetUPSList()")
	}

	var metrics []Metric
	var metric Metric

	for num := range upsList {
		upsVariables, err := upsList[0].GetVariables()
		if err != nil {
			log.Fatal().Err(err).Str("hostname", *hostname).Msg("unable to GetVariables()")
		}
		for _, obj := range upsVariables {
			if _, ok = trackedVariables[obj.Name]; ok {
				metric.Topic = fmt.Sprintf("%s/%s", upsList[num].Name, obj.Name)

				// Hardcoded transforms
				switch name := obj.Name; name {
				case "ups.status":
					metric.Message = upsStatus[obj.Value.(string)]
					metrics = append(metrics, metric)
					log.Debug().Msgf("%s/%s = %s", upsList[num].Name, obj.Name, metric.Message)
				case "battery.runtime":
					metric.Message = formatValue(obj.Value.(int64) / 60)
					metrics = append(metrics, metric)
					log.Debug().Msgf("%s/%s = %s", upsList[num].Name, obj.Name, metric.Message)
				default:
					metric.Message = formatValue(obj.Value)
					metrics = append(metrics, metric)
					log.Debug().Msgf("%s/%s = %s", upsList[num].Name, obj.Name, metric.Message)
				}
			}
		}
	}
	return metrics
}

func formatValue(i interface{}) string {
	var retval string
	var err error
	switch v := i.(type) {
	case int64:
		retval = fmt.Sprintf("%v", v)
	case string:
		retval = fmt.Sprintf("%s", v)
	case float64:
		retval = fmt.Sprintf("%.1f", v)
	default:
		return ""
	}
	if err != nil {
		fmt.Println(err)
	}
	return retval
}
