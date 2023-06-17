package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tatsushid/go-fastping"
)

var (
	latency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ip_latency",
			Help: "Latency information for IP",
		},
		[]string{
			"ip",
		},
	)
)

func main() {
	ips, err := getIPsFromFile("iplist.txt")
	if err != nil {
		log.Fatalf("Failed to get IPs from file: %s", err)
	}

	r := prometheus.NewRegistry()
	r.MustRegister(latency)

	go runPingMonitor(ips)

	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func getIPsFromFile(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ips []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if net.ParseIP(ip) == nil {
			log.Printf("WARNING: Invalid IP address: %s", ip)
		} else {
			ips = append(ips, ip)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ips, nil
}

func runPingMonitor(ips []string) {
	for {
		p := fastping.NewPinger()
		p.MaxRTT = time.Second

		for _, ip := range ips {
			ra, err := net.ResolveIPAddr("ip4:icmp", ip)
			if err != nil {
				log.Fatalf("Failed to resolve address: %s", err)
			}
			p.AddIPAddr(ra)
		}

		p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
			if rtt > 50*time.Millisecond {
				log.Printf("WARNING: IP %s latency is %s which is greater than 50ms\n", addr.String(), rtt)
			}
			latency.With(prometheus.Labels{"ip": addr.String()}).Set(float64(rtt / time.Millisecond))
		}

		err := p.Run()
		if err != nil {
			log.Printf("ERROR: Request timed out for %s\n", err)
		}

		time.Sleep(500 * time.Millisecond)
	}
}
