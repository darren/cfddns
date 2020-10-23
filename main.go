package main

import (
	"flag"
	"log"
	"os"
	"time"
)

var key = flag.String("key", os.Getenv("CF_API_KEY"), "the cloudflare api key")
var email = flag.String("email", os.Getenv("CF_API_EMAIL"), "the cloudflare api email")
var zone = flag.String("zone", "", "the zone name, like example.com")
var name = flag.String("name", "", "hostname to update, like www without example.com")
var ipv4 = flag.String("ipv4", "", "ipv4 to update, if emtpy guess ipv4 address from system, use no to skip ipv4 update")
var ipv6 = flag.String("ipv6", "", "ipv6 to update, if emtpy guess ipv6 address from system, use no to skip ipv6 update")
var duration = flag.Duration("duration", time.Minute, "interval to update")
var dnsResolver = flag.String("resolver", "1.1.1.1,8.8.8.8,9.9.9.9", "resolver to use to check before update, if empty, use system resolver")

func main() {
	flag.Parse()

	if *key == "" || *email == "" {
		log.Fatalf("key or email is empty")
	}

	if *zone == "" || *name == "" {
		log.Fatalf("zone or name not specified")
	}

	client, err := NewClient(*key, *email, *zone)
	if err != nil {
		log.Fatal(err)
	}

	if *dnsResolver != "" {
		initResolver(*dnsResolver)
		log.Printf("Using %s to check", *dnsResolver)
	} else {
		log.Println("Using system resolver to check")
	}

	for {
		if *ipv4 != "no" {
			ip, err := LocalIP("IPv4")
			if err == nil {
				err = client.UpdateIPv4(*name, ip.String())
				if err != nil {
					log.Println(err)
				} else {
					log.Printf("Update %s ipv4 to %s ok", *name, ip.String())
				}
			}
		}

		if *ipv6 != "no" {
			ip, err := LocalIP("IPv6")
			if err == nil {
				err = client.UpdateIPv6(*name, ip.String())
				if err != nil {
					log.Println(err)
				} else {
					log.Printf("Update %s ipv6 to %s ok", *name, ip.String())
				}
			}
		}

		time.Sleep(*duration)
	}
}
