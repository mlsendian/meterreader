SSH_USER=pi
RASPBERRYPI=raspberrypi.local

all:
	env GOOS=linux GOARCH=arm GOARM=6 go build

install:
	sudo install -m 755 -u root -g root meterreader /usr/local/bin/meterreader
	sudo install -m 644 -u root -g root meterreader.service /etc/systemd/system/meterreader.service
	if [ ! -f /etc/default/meterreader ]; then sudo install -m 644 -o root -g root meterreader.env /etc/default/meterreader; fi;
	systemctl enable meterreader
	systemctl start meterreader

ssh_install:
	scp meterreader meterreader.service meterreader.env ${SSH_USER}@${RASPBERRYPI}:
	ssh pi@raspberrypi.local sudo install -m 755 -o root -g root meterreader /usr/local/bin/meterreader
	ssh pi@raspberrypi.local sudo install -m 644 -o root -g root meterreader.service /etc/systemd/system/meterreader.service
	ssh pi@raspberrypi.local 'if [ ! -f /etc/default/meterreader ]; then sudo install -m 644 -o root -g root meterreader.env /etc/default/meterreader; fi;'
	ssh pi@raspberrypi.local sudo systemctl enable meterreader
	ssh pi@raspberrypi.local sudo systemctl restart meterreader