# What's that?

Clingon is a Go library that helps in creating a command-line
interface (CLI) à la Quake. Can't wait to see how does it look? Watch
this [video](http://www.youtube.com/watch?v=nee3BOtvUCE) and see if it
is worth reading below.

Clingon exploits a fully concurrent and (hopefully) clean
design. Basically, there are two goroutine running in parallel:

* The console goroutine exposes command-line functionalities to client
  code. Client code sends characters and commands to this goroutine in
  order to modify the console state.

* The renderer goroutine receives console instances (forwarded by the
  console goroutine) in order to render a graphical representation of
  the console state.

<pre>
 +---------------+	      +---------------+
 |    console    |	      |   renderer    |
 |   goroutine   |------->|   goroutine   |
 +---------------+	      +---------------+
        ^
        |				 
 +---------------+
 |  client code  |
 +---------------+
</pre>

The fact that clingon runs in parallel makes it an interesting choice
for adding console functionalities to games and graphical
applications. For example, it is used in
[gospeccy](https://github.com/remogatto/gospeccy) allowing for
displaying and using a CLI without blocking the emulation process.

Moreover, the fact that console operations are isolated from the
rendering backend allows for non-blocking graphical effects. For
example, there could be a nice non-blocking scrolling effect happening
on newline events.

Clingon tries to emulate a part of readline-like functionalities for
line editing. At the moment, only a very small subset of the whole
readline commands are available. However, clingon will never offer a
complete readline emulation as it's simply not needed in games. See
the Features section for more details.

Clingon is renderer-agnostic. Currently, only an SDL renderer is
available in the package. But it should not be difficult to add
different backends (e.g. opengl, draw/x11, etc.)

# Features

* Concurrent design
* Graphical backend agnostic (currently an SDL renderer is included)
* Readline-like commands
** left/right cursor movements
** up/down history browsing

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
    examples/shell -auto-fps -bg-image testdata/gopher.jpg testdata/VeraMono.ttf

# TODO

* Improve readline emulation adding more commands 
* Add new rendering backends
* Add graphical effects

# Video

* [Clingon demo](http://www.youtube.com/watch?v=nee3BOtvUCE)

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





