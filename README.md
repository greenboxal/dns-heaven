# dns-heaven

dns-heaven fixes macOS DNS stack by enabling the usage of the native DNS stack through /etc/resolv.conf.

# Overview

Some programs like dig, nslookup and anything compiled with Go doesn't use macOS native name resolution stack. This makes some features like split DNS to not work with those programs.

This occurs because macOS native name resolution uses a set of rules that aren't compatible with resolv.conf. This includes:

* Per interface DNS settings (scoped)
* Per domain settings

In order to support programs that uses resolv.conf, macOS writes a file with only the primary name server and search domains that were configured either through DHCP or manually.

# Installation

Just run:

    curl -L https://git.io/fix-my-dns-plz | sudo bash

This script downloads the latest version and installs a LaunchAgent making sure that dns-heaven is always running.

If you want to do this manually, just download the latest release or compile dns-heaven yourself, and make sure it's always running.

# How it works

dns-heaven exposes a DNS server that acts as a proxy mimicking native macOS behaviour. This is accomplished by periodically reading the output of `scutil --dns` and updating upstream rules and nameservers.

It also keeps /etc/resolv.conf pointing to 127.0.0.1 as the system will rewrite this file whenever your network settings changes (e.g.: changing wifi network).

# Alternatives

## dnsmasq
This is one of the best options but it has some drawbacks. In order to use dnsmasq you need to manually specify it on network settings and manually configure the upstream forwarders. This is bad because sometimes you want to use the servers announced on DHCP instead of something static like 8.8.8.8 and 8.8.4.4.

# License

[MIT](LICENSE).
