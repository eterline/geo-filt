# geo-filt

**geo-filt** is a [Traefik](https://traefik.io/) middleware plugin that filters incoming HTTP requests by IP address and geographic location (country, subnet, custom rules). Written in ![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)

## Features

- Extracts client IP from standard headers: `Forwarded`, `X-Forwarded-For`, `X-Real-IP`.
- IPv4 and IPv6 support.
- Option to allow private ranges (RFC1918, RFC4193, loopback).
- Country-based access filtering (ISO codes).
- IP or subnet allow-list.
- Fully compatible with the [Traefik Plugin System](https://doc.traefik.io/traefik/plugins/overview/).

## Installation

Enable the local plugin in your `traefik.yaml` or `config.yaml`:

```yaml
experimental:
  localPlugins:
    traefik-plugin-geo-filt:
      moduleName: github.com/eterline/geo-filt
```

### Configuration

Example middleware configuration (dynamic.yaml):

```yaml
http:
  middlewares:
    geofilter:
        plugin:
            traefik-plugin-geo-filt:
                enabled: true # Turn on filter plugin
                headerBearer: false # Search request IP in headers "Forwarded", "X-Forwarded-For", "X-Real-IP"
                allowPrivate: true # Allows RFC 1918 (IPv4 addresses), RFC 4193 (IPv6 addresses) and loopback IPs
                codeFile: "./plugins/github.com/eterline/geo-filt/dataset/locations.csv"
                geoFile: 
                  - "./plugins/github.com/eterline/geo-filt/dataset/subnets_ipv4.csv"
                  - "./plugins/github.com/eterline/geo-filt/dataset/subnets_ipv6.csv"
                tags: ["ru"] # Country allow code
                defined: ["216.58.211.0/24", "142.250.74.142"] # IP or subnet allow

```

Default parameters usage:
| Option         | Type      | Default | Description                                                           |
| -------------- | --------- | ------- | --------------------------------------------------------------------- |
| `enabled`      | bool      | `false` | Enable or disable the filter                                          |
| `headerBearer` | bool      | `false` | If `true`, try to resolve IP from headers; otherwise use `RemoteAddr` |
| `allowPrivate` | bool      | `false` | Allow private and loopback IPs                                        |
| `codeFile`     | string    | —       | Path to CSV with country codes                                        |
| `geoFile`      | \[]string | —       | Paths to IP subnets CSV                                              |                                      |
| `tags`         | \[]string | —       | Allowed country ISO codes                                             |
| `defined`      | \[]string | —       | Additional allowed IPs or subnets                                     |

Example router usage:
```yaml
http:
  routers:
    my-app:
      rule: "Host(`example.com`)"
      service: my-app-svc
      middlewares:
        - geofilter@file
```

## License

[MIT](https://choosealicense.com/licenses/mit/)