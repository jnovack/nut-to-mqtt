package mqttpub

import (
	"crypto/tls"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	secrets "github.com/ijustfool/docker-secrets"
	goVersion "github.com/jnovack/go-version"
	"github.com/namsral/flag"
	"github.com/rs/zerolog/log"
)

var (
	// Client is an interface to the MQTT client
	Client mqtt.Client
	// Flag variables
	endpoint     = flag.String("mqtt_endpoint", "tcp://mosquitto:1883", "mosquitto message broker endpoint")
	clientid     = flag.String("mqtt_clientid", "random", "mqtt client id")
	username     = flag.String("mqtt_username", "", "username for mqtt authentication")
	password     = flag.String("mqtt_password", "", "password for mqtt authentication")
	passwordFile = flag.String("mqtt_password_file", "", "path to the 'mqtt_password' file, which holds the mqtt_password")
	certFile     = flag.String("mqtt_certfile", "", "certificate (in pem format) for mqtt authentication")
	keyFile      = flag.String("mqtt_keyfile", "", "private key (in pem format) for mqtt authentication")
)

func init() {
	log.Logger = log.With().Str("component", "mqttpub").Logger()
}

// Noop is for testing without re-writing
func Noop() mqtt.Client {
	return mqtt.NewClient(mqtt.NewClientOptions())
}

// Connect does nothing apparently
func Connect() mqtt.Client {
	if *passwordFile != "" {
		dockerSecrets, _ := secrets.NewDockerSecrets(filepath.Dir(*passwordFile))
		secret, _ := dockerSecrets.Get("mqtt_password")
		*password = secret
	}

	opts := mqtt.NewClientOptions()
	opts.SetCleanSession(true)
	opts.AddBroker(*endpoint)

	// set client id if it is not random
	if *clientid != "random" {
		opts.SetClientID(*clientid)
	} else {
		opts.SetClientID(fmt.Sprintf("%s_%v", goVersion.Application, time.Now().Unix()))
	}

	// if you have a username you'll need a password with it
	if *username != "" {
		opts.SetUsername(*username)
		if *password != "" {
			opts.SetPassword(*password)
		} else {
			log.Fatal().Msg("MQTT requires a password when you supply a username")
		}
	}

	// if you have a client certificate you want a key aswell
	if *certFile != "" && *keyFile != "" {
		keyPair, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			log.Err(err).Msg("Failed to load certificate/keypair")
		}
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{keyPair},
			InsecureSkipVerify: true,
			ClientAuth:         tls.NoClientCert,
		}
		opts.SetTLSConfig(tlsConfig)
		if !strings.HasPrefix(*endpoint, "ssl://") &&
			!strings.HasPrefix(*endpoint, "tls://") {
			log.Warn().Msg("Warning: To use TLS the endpoint URL will have to begin with 'ssl://' or 'tls://'")
		}
	} else if (*certFile != "" && *keyFile == "") ||
		(*certFile == "" && *keyFile != "") {
		log.Warn().Msg("Warning: For TLS to work both certificate and private key are needed. Skipping TLS.")
	}

	opts.OnConnect = func(client mqtt.Client) {
		log.Info().Msgf("Connected to %s", *endpoint)

		// subscribe on every (re)connect
		// may be useful later for "poll now"?
		// token := client.Subscribe("$SYS/#", 0, func(_ mqtt.Client, msg mqtt.Message) {
		// 	processUpdate(msg.Topic(), string(msg.Payload()))
		// })
		// if !token.WaitTimeout(10 * time.Second) {
		// 	log.Error().Msg("Error: Timeout subscribing to topic $SYS/#")
		// }
		// if err := token.Error(); err != nil {
		// 	log.Error().Msgf("Failed to subscribe to topic $SYS/#: %s", err)
		// }
	}

	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Warn().Msgf("Warning: Connection to %s lost: %s", *endpoint, err)
	}

	client := mqtt.NewClient(opts)

	// launch the first connection in another thread so it is no blocking
	// and exporter can serve metrics in case of no connection
	go mqttConnect(client)
	return client
}

// try to connect forever with the MQTT broker
func mqttConnect(client mqtt.Client) {
	// try to connect forever
	for {
		token := client.Connect()
		log.Info().Str("endpoint", *endpoint).Msg("Attempting to connect to mosquitto endpoint")
		if token.WaitTimeout(5 * time.Second) {
			if token.Error() == nil {
				break
			}
			log.Error().Err(token.Error()).Str("endpoint", *endpoint).Msg("Failed to connect to mosquitto endpoint")
		} else {
			log.Error().Str("endpoint", *endpoint).Msg("Timeout connecting to mosquitto endpoint")
		}
		time.Sleep(5 * time.Second)
	}
}

// process the messages received in $SYS/
/*
func processUpdate(topic, payload string) {
	//log.Debugf("Got broker update with topic %s and data %s", topic, payload)
	if _, ok := ignoreKeyMetrics[topic]; !ok {
		if _, ok := counterKeyMetrics[topic]; ok {
			log.Debug().Str("topic", topic).Str("payload", payload).Msg("Processing counter metric")
			processCounterMetric(topic, payload)
		} else {
			log.Debug().Str("topic", topic).Str("payload", payload).Msg("Processing gauge metric")
			processGaugeMetric(topic, payload)
		}
		// restartSecondsSinceLastUpdate()
	} else {
		log.Debug().Str("topic", topic).Str("payload", payload).Msg("Ignoring metric")
	}
}
*/

/*
func processCounterMetric(topic, payload string) {
	// if counterMetrics[topic] != nil {
	// 	value := parseValue(payload)
	// 	counterMetrics[topic].  .Set(value)
	// } else {
	// 	// create a mosquitto counter pointer
	// 	mCounter := NewMosquittoCounter(prometheus.NewDesc(
	// 		parseForPrometheus(topic),
	// 		topic,
	// 		[]string{},
	// 		prometheus.Labels{},
	// 	))

	// 	// register the metric
	// 	prometheus.MustRegister(mCounter)
	// 	// add the first value
	// 	value := parseValue(payload)
	// 	counterMetrics[topic].Set(value)
	// }
}

func processGaugeMetric(topic, payload string) {
	if gaugeMetrics[topic] == nil {
		gaugeMetrics[topic] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        parseForPrometheus(topic),
			Help:        topic,
			ConstLabels: prometheus.Labels{"broker": *endpoint},
		})
		// register the metric
		prometheus.MustRegister(gaugeMetrics[topic])
		// 	// add the first value
	}
	value := parseValue(payload)
	gaugeMetrics[topic].Set(value)
}
*/

/*
func parseForPrometheus(incoming string) string {
	outgoing := strings.Replace(incoming, "$SYS", "mqtt", 1)
	outgoing = strings.Replace(outgoing, "/", "_", -1)
	outgoing = strings.Replace(outgoing, " ", "_", -1)
	outgoing = strings.Replace(outgoing, "-", "_", -1)
	outgoing = strings.Replace(outgoing, ".", "_", -1)
	return outgoing
}

func parseValue(payload string) float64 {
	// fmt.Printf("Payload %s \n", payload)
	var validValue = regexp.MustCompile(`-?\d{1,}[.]\d{1,}|\d{1,}`)
	// get the first value of the string
	strArray := validValue.FindAllString(payload, 1)
	if len(strArray) > 0 {
		// parse to float
		value, err := strconv.ParseFloat(strArray[0], 64)
		if err == nil {
			return value
		}
	}
	return 0
}
*/
