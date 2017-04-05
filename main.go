package main

import (
	"flag"
	"github.com/tomasen/fcgi_client"
	"gopkg.in/tylerb/graceful.v1"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

var (
	statusLineRegexp = regexp.MustCompile(`(?m)^(.*):\s+(.*)$`)
	fpmScheme        = ""
	fpmPath          = ""
)

func main() {
	fpmAddress := flag.String("fpm-address", "", "PHP-FPM address or unix path. Ej: tcp://127.0.0.1:9000 or unix:/path/to/unix.sock")
	statusPath := flag.String("status-path", "/status", "PHP-FPM status path")
	addr := flag.String("addr", "0.0.0.0:9237", "IP/port for the HTTP server")
	flag.Parse()

	if *fpmAddress == "" {
		log.Fatal("The fpm-address flags is required.")
	}

	parsedURI, err := url.ParseRequestURI(*fpmAddress)
	if err != nil {
		log.Fatalf("Cant parse fpm-address: %s", err)
	}

	fpmScheme = parsedURI.Scheme

	switch fpmScheme {
	case "unix":
		fpmPath = parsedURI.RequestURI()
	case "tcp":
		fpmPath = parsedURI.Hostname() + ":" + parsedURI.Port()
	default:
		log.Fatalf("Unsupported protocol: %s", fpmScheme)
	}

	scrapeFailures := 0

	// TODO: Migrate to go 1.8 with built in graceful shutdown
	server := &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:        *addr,
			ReadTimeout: time.Duration(5) * time.Second,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				env := make(map[string]string)
				env["SCRIPT_NAME"] = *statusPath
				env["SCRIPT_FILENAME"] = *statusPath

				fcgi, err := fcgiclient.Dial(fpmScheme, fpmPath)
				defer fcgi.Close()

				if err != nil {
					log.Println(err)
					scrapeFailures = scrapeFailures + 1
					x := strconv.Itoa(scrapeFailures)
					NewMetricsFromMatches([][]string{{"scrape failure:", "scrape failure", x}}).WriteTo(w)
					return
				}

				resp, err := fcgi.Get(env)

				if err != nil {
					log.Println(err)
					scrapeFailures = scrapeFailures + 1
					x := strconv.Itoa(scrapeFailures)
					NewMetricsFromMatches([][]string{{"scrape failure:", "scrape failure", x}}).WriteTo(w)
					return
				}

				//if resp.StatusCode != http.StatusOK {
				//	log.Println("php-fpm status code is not OK.")
				//	scrapeFailures = scrapeFailures + 1
				//	x := strconv.Itoa(scrapeFailures)
				//	NewMetricsFromMatches([][]string{{"scrape failure:", "scrape failure", x}}).WriteTo(w)
				//	return
				//}

				body, err := ioutil.ReadAll(resp.Body)

				if err != nil {
					log.Println("3")
					log.Println(err)
					scrapeFailures = scrapeFailures + 1
					x := strconv.Itoa(scrapeFailures)
					NewMetricsFromMatches([][]string{{"scrape failure:", "scrape failure", x}}).WriteTo(w)
					return
				}

				resp.Body.Close()

				x := strconv.Itoa(scrapeFailures)

				matches := statusLineRegexp.FindAllStringSubmatch(string(body), -1)
				matches = append(matches, []string{"scrape failure:", "scrape failure", x})

				NewMetricsFromMatches(matches).WriteTo(w)
			}),
		},
	}

	log.Printf("Server started on %s", *addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
