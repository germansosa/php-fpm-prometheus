package main

import (
	"flag"
	"github.com/tomasen/fcgi_client"
	"gopkg.in/tylerb/graceful.v1"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	statusLineRegexp = regexp.MustCompile(`(?m)^(.*):\s+(.*)$`)
	fpmStatusHost    = ""
	fpmStatusPort    = ""
)

func main() {
	urlhost := flag.String("status-host", "", "PHP-FPM status host")
	urlport := flag.String("status-port", "9000", "PHP-FPM status port")
	addr := flag.String("addr", "0.0.0.0:8080", "IP/port for the HTTP server")
	flag.Parse()

	if *urlhost == "" {
		log.Fatal("The status-host flags is required.")
	} else {
		fpmStatusHost = *urlhost
	}
	fpmStatusPort = *urlport

	scrapeFailures := 0

	// TODO: Migrate to go 1.8 with built in graceful shutdown
	server := &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:        *addr,
			ReadTimeout: time.Duration(5) * time.Second,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				env := make(map[string]string)
				env["SCRIPT_NAME"] = "/status"
				env["SCRIPT_FILENAME"] = "/status"

				fcgi, err := fcgiclient.Dial(fpmStatusHost, fpmStatusPort)
				defer fcgi.Close();

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

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
