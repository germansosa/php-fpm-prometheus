# php-fpm-prometheus

Simple [PHP-FPM](http://php.net/manual/en/install.fpm.php) status exporter for [Prometheus](https://prometheus.io/) over fastcgi.

## Installation

If you are using Go 1.6+ (or 1.5 with the `GO15VENDOREXPERIMENT=1` environment variable), you can install `php-fpm-prometheus` with the following command:

```bash
$ go get -u github.com/gsosa/php-fpm-prometheus
```

## Usage

```bash
$ ./php-fpm_prometheus --help
Usage of ./php-fpm_exporter:
  -addr string
    	IP/port for the HTTP server (default "0.0.0.0:9237")
  -fpm-address string
    	PHP-FPM address or unix path. Ej: tcp://127.0.0.1:9000 or unix:/path/to/unix.sock
  -status-path string
    	PHP-FPM status path (default "/status")

$ ./php-fpm_prometheus --fpm-address "tcp://127.0.0.1:9000" -status-path "/status" -addr "127.0.0.1:8080"
```

Finally, point Prometheus to `http://127.0.0.1:8080/metrics`.

## Contributing

All contributions are welcome, but if you are considering significant changes, please open an issue beforehand and discuss it with us.

## License

MIT. See the `LICENSE` file for more information.
