# dns-heaven

This project fixes macOS DNS stack by enabling the usage of the native DNS stack through /etc/resolv.conf.

# Why

Some programs like dig, nslookup and anything compiled with Go, doesn't make usage of macOS native name resolution stack. This makes some features like Split DNS not work with those programs.

# Installing

Just run:

    curl https://git.io/fix-my-dns-plz | sudo bash

This script downloads that latest version and install a LaunchAgent so dns-heaven is always running.

## G1mme the detailz

Just download or build the binary and make sure it's running. Everything is automatic.

# How

dns-heaven is simple DNS proxy that mimics macOS name resolution stack by reading its configuration and proxyfying requests to the correct nameservers.

## I want more details!

From time to time, dns-heaven parses the output of `scutil --dns`, and tries to mimic macOS behaviour using the output as config.

Also, it forces `/etc/resolv.conf` to point to 127.0.0.1 as the native network manager will try to change it to the DNS servers configured on Network Settings or on DHCP.

# License

MIT.

