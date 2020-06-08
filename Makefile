GOPATH	= $(CURDIR)
BINDIR	= $(CURDIR)/bin

PROGRAMS = ipxe_distributor

depend:
	env GOPATH=$(GOPATH) go get -u gopkg.in/yaml.v2
	env GOPATH=$(GOPATH) go get -u github.com/sirupsen/logrus
	env GOPATH=$(GOPATH) go get -u github.com/gorilla/mux

build: depend
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

distclean: clean
	/bin/rm -rf src/gopkg.in/
	/bin/rm -rf src/github.com/
	/bin/rm -rf src/golang.org/

uninstall:
	/bin/rm -f $(DESTDIR)/usr/bin

all: build strip install

