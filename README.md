# dns-split

This is a simple dns proxy service, capable of routing DNS queries to two different DNS servers based on configuration.

## why?

While `dns-split` could be used in itself as a local DNS server, the main intent of it is to use it with a split-tunneling vpn setup, for example global protect  `<include-split-tunneling-domain>`.

## install

```
git clone https://git.okki.hu/garric/dns-split.git
cd dns-split
sudo make clean install
```

This will install the following:

* binary (`/usr/local/bin`)
* default configuration (`/etc/dns-split/config.json`)
* systemd service (`/etc/systemd/system/dns-split.service`)

## upgrade

To upgrade from an older version, use the `upgrade` make target. This will only update the binary and leave your config file and systemd service definition intact:

```
sudo make upgrade
```

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

The following command line args can be used:

| Command Line Arg | Description                        | Default Value              |
| ---------------- | ---------------------------------- | -------------------------- |
| `--config`       | location of the `config.json` file | /etc/dns-split/config.json |
| `--verbose`      | enable verbose logging             | false                      |

## configuration

An example configuration is included in `etc/config` and is copied to `/etc/dns-split/config.json` during install.

```json
{
  "addr": "127.0.0.59",
  "dns": "1.1.1.1",
  "split-tunnel": {
    "domains": [
      "*.example.org.",
      "*.example.com.",
      "*.example.net."
    ],
    "dns": "9.9.9.9",
    "netlink": {
      "enabled": true,
      "dev": "tun0"
    }
  }
}
```

_Note: the trailing "`.`" characters in the domain names is not a mistake, make sure you end your domains with "`.`" as well, otherwise they will not match!_

The configuration file is a simple json document, and the following options can be used in it:

| Attribute      | Description                                     | Type   |
| -------------- | ----------------------------------------------- | ------ |
| `addr`         | the address to start `dns-split` server on      | string |
| `dns`          | the default DNS server to proxy requests to     | string |
| `split-tunnel` | an object containing split-tunnel configuration | object |

The `split-tunnel` object contains the following fields:

| Attribute | Description                                                                                          | Type             |
| --------- | ---------------------------------------------------------------------------------------------------- | ---------------- |
| `domains` | a list of domain patterns                                                                            | array of strings |
| `dns`     | the DNS server to proxy to if the host name matches any of the patterns from the domains array above | string           |
| `netlink` | an optional netlink configuration object                                                             | object           |


The `netlink` object contains the following fields:

| Attribute | Description                                       | Type    |
| --------- | ------------------------------------------------- | ------- |
| `enabled` | a value indicating if the netlink hook is enabled | boolean |
| `dev`     | the network device name to set ip links on        | string  |



## configuring for vpn split-dns usage

A working DNS split tunnel setup will revolve around the following:

* dns-split is configured as your default local DNS server
* configuration:
    - the default DNS server is set to a real remote DNS server
    - the split-tunnel DNS server is set to the vpn DNS server
    - a set of domains that need to be routed through the vpn
    - (optionally) enabled netlink and vpn network device specified


The workflow will look like this:

1. a dns query arrives
2. the query is routed to dns-split
3. dns-split cross-references the hostname with its domain patterns
4. if no match:
   1. the query is forwarded to the default DNS server
   2. the response is proxied back to the caller
5. if there is a match:
   1. the query is forwarded to the split-tunnel DNS server
   2. when it responds, the netlink hook is invoked (if enabled), setting up ip routes on the kernel to the VPN network device using the resolved ip address
   3. the response is proxied back to the caller


### about the netlink hook

This hook is especially useful for global protect `<include-split-tunneling-domain>` when using `openconnect`. In this scenario the VPN server will send a list of domains (including wildcards) that should be routed through the VPN network interface.

If the domain list actually contains wildcards, `openconnect`'s `vpnc` script will not be sufficient to properly configure the VPN, since the name of the hosts is not known in advance (foo.corporate.com, bar.corporate.com, etc.)

If your VPN is not relying on domain wildcards, the netlink option can safely be disabled.


### a note on systemd-resolved

Currently `dns-split` is not confirmed to work with systemd-resolve, and most likely you would face issues because of systemd stub resolver's cache and your split-dns netlink hooks not being triggered.

To disable systemd-resolved:

```
sudo systemctl disable --now systemd-resolved.service
```

If systemd-resolved is running in stub mode, remove the symlink as well:

```
sudo rm -f /etc/resolv.conf
```

After this, proceed with the standard setup as below.

### using glibc resolver

Make sure, that:

   * `systemd-resolved` is __not__ enabled/running
   * `/etc/resolv.conf` is __not__ a symlink to a `systemd-resolved` stub


Edit `/etc/resolv.conf`, and set the default nameserver to the address of dns-split.
Note, that the port __must__ be set to `53`, and need to be defined in `/etc/resolv.conf` using this approach.

```
nameserver 127.0.0.59
```

### preventing openconnect/vpnc-script to change default DNS

The default vpnc script will temporarily change your default DNS server to route DNS queries through the VPN DNS server.

To prevent this, you can use the included `vpnc-script-no-dns.sh`. The script is installed into the `/etc/dns-split` directory by default.

```
#!/bin/sh
unset INTERNAL_IP4_DNS

VPNC_SCRIPT="/etc/vpnc/vpnc-script"
. $VPNC_SCRIPT
```

Use the above script with `openconnect`:

```
openconnect --script "/etc/dns-split/vpnc-script-no-dns.sh"
```

### configure split dns

Edit the `/etc/dns-split/config.json` file:

   * specify the default dns server
   * in the split-tunnel section:
     * specify the vpn dns server
     * specify the domains that need to be routed through the vpn:
       * use trailing `.` characters in the domains (see example above)
       * the usual `*.` wildcard can be used inside domains
     * enable the netlink hook if needed, and specify the vpn network interface


### a note on caching

It's easy to see, that if dns-split is put behind another DNS proxy that handles caching, the netlink hook might not get properly invoked (especially if the cache is persisted between restarts). This is one reason why integration with systemd-resolved is currently not supported.
