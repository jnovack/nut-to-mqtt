# nut-to-mqtt

**nut-to-mqtt** is a transmogrifier for exporting data from Network UPS Tools
(nut), and streaming to an MQTT broker.

## Quick Start

```
$ nut-to-mqtt \
    -nut_hostname=192.168.9.10 -nut_username=monitor -nut_password=hunter2 \
    -mqtt_hostname=172.23.4.5 -mqtt_username=user24601 -mqtt_password=hunter2
```

## Features

Currently tracks the following metrics from NUT:

* `battery.charge` - Battery charge (percent of full)
* `battery.runtime` - Battery runtime (seconds)
* `battery.voltage` - Battery voltage (V)
* `input.voltage` - Input voltage (V)
* `ups.load` - Load on UPS (percent of full)
* `ups.status` - UPS status

And publishes them to the MQTT broker under the following topic:

* `v1/ups/UPSNAME/METRIC`

Currently, there are some conversions to published data.

* `battery.runtime` - Read as seconds, published as minutes (with 1 decimal place)
* `ups.status` - Read as two-char code, published as full word

*Example*:

* `battery.runtime` is read as `2725` (seconds) and published to MQTT as `45.4`
(minutes)
* `ups.status` is read as `OL` (code) and published to MQTT as `Online` (text).

## Docker Secret Support

`nut_password` and `mqtt_password` permit a `_FILE` suffix (for
`nut_password_file` and `mqtt_password_file`) to not expose password via
environment variables.

You must point to a file called `nut_password` or `mqtt_password` which
should contain the password only.

*Example*:

```
$ nut-to-mqtt \
    -nut_hostname=192.168.9.10 -nut_username=monitor \
    -nut_password_file=/run/secrets/nut_password \
    -mqtt_hostname=172.23.4.5 -mqtt_username=user24601 \
    -mqtt_password_file=/run/secrets/mqtt_password
```

## Caveats

If connecting multiple **nut-to-mqtt** transmogrifiers to a single broker,
each UPS should have a different name within their `ups.conf`.  UPSes are
identified by their name only, and not by their host.

Having a UPS called `myups` on two different hosts will result in metrics
being overwritten.
