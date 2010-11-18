include $(GOROOT)/src/Make.inc

TARG=console
GOFILES=\
	console.go\
	sdlrenderer.go\

include $(GOROOT)/src/Make.pkg

