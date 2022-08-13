# dns-split

This is a simple dns proxy service, capable of routing DNS queries to different DNS servers based on configuration.
It is also able to call a custom defined script hook after DNS lookups, which is especially useful for split-dns
tunneling setup with `openvpn` (for example: global protect protocol's `<include-split-tunneling-domain>`).

## install

```
git clone https://git.okki.hu/garric/dns-split.git
cd dns-split
sudo make clean install
```

This will install the following:

* binary (`/usr/local/bin`)
* default configuration (`/etc/dns-split/config.json`)
* example scripts (`/etc/dns-split/scripts`)
* systemd service (`/etc/systemd/system/dns-split.service`)

## uninstall

Execute the `uninstall` target from the sources dir:

```
sudo make uninstall
```

## start dns-split

Enable and start the included systemd service:

```
sudo systemctl enable dns-split.service
sudo systemctl start dns-split.service
```

To check the status and logs:

```
systemctl status dns-split.service
```

## command line args

The DNS proxy will start on the default address `127.0.0.59:53` and use the default config `/etc/dns-split/config.json`. 
The default options can be overridden with the following args:

| Command Line Arg | Description               | Default Value                     |
|------------------|---------------------------|-----------------------------------|
| `--addr`         | the address to start on   | 127.0.0.59:53                     |
| `--config`       | location of `config.json` | /etc/dns-split/config.json |
| `--cache`        | enable caching            | true                              |

## set as default dns

### using resolved (recommended)

In `/etc/resolv.conf` the default nameserver should be set to the address of dns-split.
With this approach it is not possible to change the port of the DNS proxy.

```
nameserver 127.0.0.59
```

Make sure, that:

   * `systemd-resolved` is __not__ enabled/running
   * `/etc/resolv.conf` is __not__ a symlink to a `systemd-resolved` stub

### using systemd-resolve

Add the folowing to `/etc/systemd/resolved.conf.d/dns_servers.conf`:

```
[Resolve]
DNS=127.0.0.59
Domains=~.
```

Start `systemd-resolved` in stub mode:

```
ln -rsf /run/systemd/resolve/stub-resolv.conf /etc/resolv.conf
systemctl enable --now systemd-resolved.service
```

Ensure the config is picked up:

```
resolvectl status
```


## configure split dns

Edit the `config.json` file to add dns servers to proxy to.

Restrictions:

* there must be __exactly__ one server with `default` set to `true`
* a domain pattern should only appear __exactly__ once across all server nodes (multiple subdomain patterns for the same
  domain are OK)
* for each server specifying the address is mandatory

Schema:

* config (`[]Server`)

* server
    - addr (`string`)
    - default (`boolean`)
    - script (`string`)
    - domains (`[]string`)

Example:

```json
{
  "servers": [
    {
      "addr": "1.1.1.1",
      "default": true
    },
    {
      "addr": "8.8.8.8",
      "script": "/etc/dns-split/scripts/example.sh",
      "domains": [
        "~example.com",
        "*.foo.example.org",
        "*.bar.example.org"
      ]
    }
  ]
}
```

## domain wildcards

The following wildcards can be used inside domain strings:

| Wildcard    | Description                                               |
| ----------- | --------------------------------------------------------- |
| `*.foo.com` | matches all subdomains of foo.com, but not foo.com itself |
| `~foo.com`  | matches foo.com and all subdomains                        |

## implementing dns split tunnel for openconnect vpn

DNS split tunnel will revolve around the following setup:

* dns-split is configured as your default local DNS
* server configuration:
    - default (real) DNS server
    - vpn DNS with the required domain patterns and a script hook
    - a special script that sets ip routes

The workflow will look like this:

__VPN domain:__

1. dns request comes in
2. query is routed to VPN DNS server
3. VPN DNS server responds with IP address
4. script hook is invoked with IP
5. script adds kernel ip route

__Normal domain:__

1. dns request comes in
2. query is routed to default DNS server
3. DNS server responds with IP address

The config for the VPN domains should look similar to this:

```json
{
  "servers": [
    {    },
    {
      "addr": "<ip of vpn dns>",
      "script": "/etc/dns-split/scripts/add-route.sh",
      "domains": [
        "~example.com"
      ]
    }
  ]
}
```

The script is invoked right after the ip address is resolved, thus can hook into DNS resolution to set kernel ip routes:

```sh
$(ip route add $IP via 0.0.0.0 dev tun0)
```

...where `tun0` is the VPN interface name.

The script will get the required parameters as environment variables:

| Environment Variable | Description             |
| -------------------- | ----------------------- |
| $HOST_NAME           | the host queried        |
| $IP                  | the resolved ip address |

### note on caching

If caching is enabled and a query hits the cache the script hook is _not_ invoked.
This is working as intended: since the ip route has already been added and the IP did not change (cache hit) there is no need to call the hook either.
