DESTDIR =
PREFIX  =/usr/local

all:
clean:
install:
update:
## -- AUTO-GO --
GO_PROGRAMS += bin/imgstore$(EXE) 
.PHONY all-go: $(GO_PROGRAMS)
all:     all-go
install: install-go
clean:   clean-go
deps:
bin/imgstore$(EXE): deps 
	go build -o $@ $(IMGSTORE_FLAGS) $(GO_CONF) ./cmd/imgstore
install-go:
	install -d $(DESTDIR)$(PREFIX)/bin
	cp bin/imgstore$(EXE) $(DESTDIR)$(PREFIX)/bin
clean-go:
	rm -f $(GO_PROGRAMS)
## -- AUTO-GO --
## -- AUTO-SERVICE --

## -- AUTO-SERVICE --
