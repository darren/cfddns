## cfddns: a simple ddns update agent for cloudflare

## Usage

Get CF_API_KEY from Cloudflare: 

https://support.cloudflare.com/hc/en-us/articles/200167836-Managing-API-Tokens-and-Keys

```sh
export CF_API_KEY=APIKEY
export CF_API_EMAIL=user@example.com
cfddns --zone example.com --name www
```

It will run as daemon and check/update ip address every minute

Run `cfddns --help` for more options
