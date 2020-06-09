# prometheus p1 exporter 

The prometheus p1 exporter is a simple (by design and purpose) binary that can read data from a smart meter through a serial port and expose these metrics to be scraped by prometheus.

## configuration

The only things that can be configured with an environment var is:
 - **SERIAL_DEVICE**: the device that needs to be read (usually something like /dev/ttyUSB0)


## limitations

Lines are processed as they come in, no checksums are handled.

For now, only below fields are exported:

- Actual electricity power delivered (+P) in 1 Watt resolution
- Actual electricity power received (-P) in 1 Watt resolution
- Meter Reading electricity delivered to client (Tariff 1) in 0,001 kWh
- Meter Reading electricity delivered to client (Tariff 2) in 0,001 kWh
- Meter Reading electricity delivered by client (Tariff 1) in 0,001 kWh
- Meter Reading electricity delivered by client (Tariff 2) in 0,001 kWh
- Gas meter reading in m3
- Instantaneous voltage L1 in V resolution
- Instantaneous voltage L2 in V resolution
- Instantaneous voltage L3 in V resolution
- Instantaneous active power L1 (+P)in W resolution
- Instantaneous active power L2 (+P)in W resolution
- Instantaneous active power L3 (+P)in W resolution
- Instantaneous active power L1 (-P)in W resolution
- Instantaneous active power L2 (-P)in W resolution
- Instantaneous active power L3 (-P)in W resolution

For the OID's of these values, please refer to the ESMR 5.0 document.

## sources and acknowledgements

This repo for providing me with the first version: https://github.com/gnur/prometheus-p1-exporter

This document for ESMR 5.0 format: https://www.netbeheernederland.nl/_upload/Files/Slimme_meter_15_a727fce1f1.pdf
