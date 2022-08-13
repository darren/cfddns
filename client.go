package cfddns

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudflare/cloudflare-go"
)

// Client gateway through cloudflare
type Client struct {
	api  *cloudflare.API
	name string // cloudflare dns name, like example.com
	zid  string // zone id
}

// NewClient create the Client
func NewClient(key, token, email, name string) (*Client, error) {
	if key == "" && token == "" {
		return nil, fmt.Errorf("key or token is empty")
	}

	var api *cloudflare.API
	var err error

	if key != "" {
		api, err = cloudflare.New(key, email)
		if err != nil {
			return nil, err
		}
	} else {
		api, err = cloudflare.NewWithAPIToken(token)
		if err != nil {
			return nil, err
		}
	}

	zid, err := api.ZoneIDByName(name)
	if err != nil {
		return nil, err
	}

	return &Client{
		api:  api,
		zid:  zid,
		name: name,
	}, nil
}

// UpdateIPv4 update ipv4 record for name
func (c *Client) UpdateIPv4(ctx context.Context, name, address string) error {
	return c.update(ctx, name, address, "A")
}

// UpdateIPv6 update ipv6 record for name
func (c *Client) UpdateIPv6(ctx context.Context, name, address string) error {
	return c.update(ctx, name, address, "AAAA")
}

func (c *Client) fqdn(name string) string {
	return fmt.Sprintf("%s.%s", name, c.name)
}

// check whether ip is up to date
func (c *Client) check(name, address, rtype string) bool {
	fqdn := c.fqdn(name)
	ip, err := Resolve(fqdn, rtype)
	if err != nil {
		log.Println(err)
	}
	if ip == address {
		return true
	}
	return false
}

func (c *Client) update(ctx context.Context, name, address, rtype string) error {
	fqdn := c.fqdn(name)
	if c.check(name, address, rtype) {
		log.Printf("%s %s is up to date", fqdn, address)
		return nil
	}

	rr, err := c.api.DNSRecords(
		ctx,
		c.zid,
		cloudflare.DNSRecord{
			Name: fqdn,
			Type: rtype,
		},
	)
	if err != nil {
		return err
	}

	var rid string
	var ok bool
	var oaddress string

	for _, r := range rr {
		if r.Name == fqdn && r.Type == rtype {
			ok = true
			rid = r.ID
			oaddress = r.Content
		}
		log.Printf("%v in current cloudflare zone: %v %v\n", r.Name, r.Content, r.Type)
	}

	if ok {
		if oaddress == address {
			return fmt.Errorf("%s not changed: %s", c.fqdn(name), address)
		}

		err = c.api.UpdateDNSRecord(
			ctx,
			c.zid,
			rid,
			cloudflare.DNSRecord{
				Name:    name,
				Type:    rtype,
				Content: address,
			},
		)
		if err != nil {
			return fmt.Errorf("update record for %s %s failed %w", name, rid, err)
		}
	} else {
		_, err = c.api.CreateDNSRecord(
			ctx,
			c.zid,
			cloudflare.DNSRecord{
				Name:    name,
				Type:    rtype,
				Content: address,
			},
		)
		if err != nil {
			return fmt.Errorf("create record for %s failed %w", name, err)
		}
	}
	return nil
}
