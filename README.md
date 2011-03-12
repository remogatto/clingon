# What's that?

Clingon (Command Line INterface for GO Nuts) is a Go library that
helps in creating command-line interfaces à la Quake. Can't wait to
see how does it look? Watch this
[video](http://www.youtube.com/watch?v=nee3BOtvUCE) (early version) and
see if it is worth reading below.

Clingon exploits a parallel and (hopefully) clean design. Basically,
there are four parts running in parallel:

* The console: it runs (almost) in the same goroutine of the
  client-code and it exposes console functionalities to it. The client
  modifies the console state through a simple API. When a new console
  instance is created, a goroutine is spawned. This goroutine simply
  triggers the "blink cursor" event to the renderer at regular time's
  interval.

* The renderer: it responds to the events triggered by the console in
  order to render a graphical representation of the console state. A
  renderer object should implement the Renderer interface.

* The animation service: it provides a stream of changing values used
  for animations. For example, this service provides the changing Y
  coordinate as a function of time in order to achieve the console
  sliding effect.

* The evaluator: it evaluates the commands sent to the console
  providing back a result. It implements the Evaluator interface.

Because of its design, clingon could be a neat choice for adding
console functionalities to games and graphical applications. For
example, it is used in
[gospeccy](https://github.com/remogatto/gospeccy/tree/clingon)
to provide a non-blocking CLI which runs in parallel with the emulator.

Moreover, the fact that console operations are isolated from the
rendering backend allows for non-blocking graphical effects.

Clingon tries to emulate a part of readline functionalities for line
editing. At the moment, only a very small subset of the whole readline
commands are available. However, clingon will never offer a complete
readline emulation as it's simply not needed in games. See the
Features section for more details.

Clingon is completely renderer-agnostic. Currently, only an SDL
renderer is being shipped with the package but it should not be
difficult to add more backends (e.g. opengl, draw/x11, etc.)

# Features

* Concurrent design
* Graphical backend agnostic (currently an SDL renderer is included)
* Console scrolling
* Readline-like commands
  * left/right cursor movements
  * up/down history browsing

# Installation

In order to use the SDL renderer, you should install the following
packages (assuming a debian-flavored linux distribution):

    sudo apt-get install libsdl1.2-dev libsdl-mixer1.2-dev libsdl-image1.2-dev libsdl-ttf2.0-dev

Clingon is using the GOAM build tool. To install it:

    goinstall github.com/0xe2-0x9a-0x9b/goam
    cd $GOROOT/src/pkg/github.com/0xe2-0x9a-0x9b/goam
    make install

To install the dependencies and build the package:

    git clone https://github.com/remogatto/clingon.git
    cd clingon
    goam make

To install (uninstall) clingon:

    goam install
    (goam uninstall)

The following dependencies are installed automatically:

* [⚛Go-SDL](https://github.com/0xe2-0x9a-0x9b/Go-SDL)
* [prettytest](https://github.com/remogatto/prettytest)

# Quick start

After installing the package try the sample code in <tt>example/</tt>
folder:

    cd clingon
    example/shell -bg-image testdata/gopher.jpg testdata/VeraMono.ttf

# TODO

* Improve readline emulation by adding more commands
* Add new rendering backends
* Add more graphical effects
* Experimenting with clingon + exp/eval + draw2d

# Video

* [Clingon demo (early version)](http://www.youtube.com/watch?v=nee3BOtvUCE)

# Credits

Thanks to the following people for the contribute they give to this
project:

* [⚛](https://github.com/0xe2-0x9a-0x9b)

# License

Copyright (c) 2010 Andrea Fazzi

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.





