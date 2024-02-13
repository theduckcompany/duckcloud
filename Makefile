
DATADIR ?= /usr/share/duckcloud
CONFIGDIR ?= /etc/duckcloud
USER = duckcloud
PASSWORDFILE=$(CONFIGDIR)/password.cred
BINARYFILE=/usr/bin/duckcloud
LICENSEFILE=/usr/share/licenses/duckcloud/LICENSE
SYSTEMDFILE=/usr/lib/systemd/system/duckcloud.service
TMPFILE=/tmp/password.txt


build: duckcloud

duckcloud:
	go build ./cmd/duckcloud/

install: build user-create masterkey-create
	cp -f duckcloud $(BINARYFILE)
	systemctl enable duckcloud.service

user-create:
	useradd -d $(DATADIR) $(USER)
	install --owner=$(USER)  -Dm700 -d $(DATADIR)
	install --backup -Dm644 ./docs/duckcloud.service.example $(SYSTEMDFILE)
	install -Dm644 LICENSE $(LICENSEFILE)

user-delete:
	userdel --remove $(USER)

start:
	systemctl start duckcloud.service

restart:
	systemctl restart duckcloud.service

status:
	systemctl status duckcloud.service

stop:
	systemctl stop duckcloud.service

masterkey-create: configdir-create
	$(info Generate the password file at $(PASSWORDFILE))
	openssl rand -hex 32 | head -n1 > $(TMPFILE)
	sudo systemd-creds encrypt --name=password $(TMPFILE) $(PASSWORDFILE)
	shred -u $(TMPFILE)

configdir-create:
	install -Dm700 -d $(CONFIGDIR)
	install --owner=$(USER) -Dm644 ./docs/var_file $(CONFIGDIR)/var_file

configdir-delete:
	[ ! -d $(CONFIDIR) ] || rm -rf $(CONFIGDIR)
	[ ! -f $(CONFIDIR) ] || rm -rf $(CONFIGDIR)

uninstall: stop user-delete configdir-delete
	rm -f $(SYSTEMDFILE)
	rm -f $(LICENSEFILE)
	rm -f $(BINARYFILE)

clean:
	rm -f duckcloud
