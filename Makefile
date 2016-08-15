all:
	$(MAKE) -C hostapd

clean:
	$(MAKE) -C hostapd clean

install:
	$(MAKE) -C hostapd install
