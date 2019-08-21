GOPATH	= $(CURDIR)
BINDIR	= $(CURDIR)/bin

PROGRAMS = ipxe_distributor

build:
	env GOPATH=$(GOPATH) go install $(PROGRAMS)

destdirs:
	mkdir -p -m 0755 $(DESTDIR)/usr/bin

strip: build
	strip --strip-all $(BINDIR)/ipxe_distributor

install: strip destdirs install-bin

install-bin:
	install -m 0755 $(BINDIR)/ipxe_distributor $(DESTDIR)/usr/bin

clean:
	/bin/rm -f bin/ipxe_distributor

uninstall:
	/bin/rm -f $(DESTDIR)/usr/bin

all: build strip install

