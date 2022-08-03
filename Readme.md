# cfddns: a simple ddns update agent for cloudflare

## Build

```
go build -o cfddns ./cmd/cfddns
```

## Usage

Get CF_API_KEY or CF_API_TOKEN from Cloudflare: 

https://support.cloudflare.com/hc/en-us/articles/200167836-Managing-API-Tokens-and-Keys

### with api key and email

```sh
export CF_API_KEY=APIKEY
export CF_API_EMAIL=user@example.com
cfddns --zone example.com --name www
```

### or use token (email can be omitted)
```
export CF_API_TOKEN=API_TOKEN_FROM_CF
cfddns --zone example.com --name www
```

It will run as daemon and check/update ip address every minute

Run `cfddns --help` for more options
