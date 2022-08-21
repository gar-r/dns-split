build:
	go build -o dns-split -buildvcs=false

clean:
	rm -f dns-split

install: build
	mv dns-split /usr/local/bin
	mkdir -p /etc/dns-split
	cp etc/config/config.json /etc/dns-split
	cp etc/scripts/vpnc-script-no-dns.sh /etc/dns-split
	cp etc/systemd/dns-split.service /etc/systemd/system
	systemctl daemon-reload
	systemctl enable --now dns-split.service

upgrade: build
	mv dns-split /usr/local/bin
	systemctl daemon-reload
	systemctl restart dns-split.service

uninstall:
	systemctl disable --now dns-split.service
	rm -f /etc/systemd/system/dns-split.service
	rm -f /usr/local/bin/dns-split
	rm -rf /etc/dns-split