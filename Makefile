include $(GOROOT)/src/Make.inc

TARG=cli
GOFILES=\
	console.go\
	sdlrenderer.go\

include $(GOROOT)/src/Make.pkg

