PROGRAM_NAME := sanepack
GOCOMPILER := go build
GOFLAGS	+= -ldflags "-X main.Version $(shell git describe --dirty=+)"


.PHONY: all install clean distclean

all: $(PROGRAM_NAME)

$(PROGRAM_NAME):
	$(GOCOMPILER) $(GOFLAGS)

clean:
	$(RM) $(PROGRAM_NAME)
