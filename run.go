package cfddns

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	key         string
	token       string
	email       string
	zone        string
	name        string
	ipv4        string
	ipv6        string
	duration    time.Duration
	dnsResolver string
)

func Run() {
	flag.StringVar(&key, "key", os.Getenv("CF_API_KEY"), "the cloudflare api key")
	flag.StringVar(&token, "token", os.Getenv("CF_API_TOKEN"), "the cloudflare api token")
	flag.StringVar(&email, "email", os.Getenv("CF_API_EMAIL"), "the cloudflare api email")
	flag.StringVar(&zone, "zone", "", "the zone name, like example.com")
	flag.StringVar(&name, "name", "", "hostname to update, like www without example.com")
	flag.StringVar(&ipv4, "ipv4", "", "ipv4 to update, if emtpy guess ipv4 address from system, use no to skip ipv4 update")
	flag.StringVar(&ipv6, "ipv6", "", "ipv6 to update, if emtpy guess ipv6 address from system, use no to skip ipv6 update")
	flag.DurationVar(&duration, "duration", time.Minute, "interval to check and update")
	flag.StringVar(&dnsResolver, "resolver", "", "resolver to use to check before update, if empty, use system resolver")

	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if key == "" && token == "" {
		log.Fatalf("key or token is empty")
	}

	if key != "" && email == "" {
		log.Fatalf("email can not be empty while key is used")
	}

	if zone == "" || name == "" {
		log.Fatalf("zone or name not specified")
	}

	if duration < 30*time.Second {
		log.Fatalf("check interval can not be less thant 30 seconds")
	}

	client, err := NewClient(key, token, email, zone)
	if err != nil {
		log.Fatal(err)
	}

	if dnsResolver != "" {
		InitResolver(dnsResolver)
		log.Printf("Using %s to check", dnsResolver)
	} else {
		log.Println("Using system resolver to check")
	}

	err = run(ctx, client)
	if err != nil {
		log.Printf("initial run failed: %v", err)
	}

	backoff := 1 * time.Minute
	for {
		select {
		case <-ctx.Done():
			log.Printf("Quiting...")
			return
		case <-time.After(duration):
			err = run(ctx, client)
			if err != nil {
				log.Printf("%s, retry after: %v", err, backoff)
				sleepWithContext(ctx, backoff)
				backoff *= 2
			}
		}
	}
}

func sleepWithContext(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(d):
		return
	}
}

func run(ctx context.Context, client *Client) error {
	if ipv4 != "no" {
		ip, err := LocalIP("IPv4")
		if err == nil {
			err = client.UpdateIPv4(ctx, name, ip.String())
			if err != nil {
				return err
			}
		}
	}

	if ipv6 != "no" {
		ip, err := LocalIP("IPv6")
		if err == nil {
			err = client.UpdateIPv6(ctx, name, ip.String())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
