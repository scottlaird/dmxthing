# dmxthing
Tool for controlling DMX video lighting using a Loupedeck as a UI controller.

My desktop video conferencing setup include several video lights which
are all controlled via DMX, so I can control brightness and color
temperature.  I have lights above my monitor on the left side,
centered above the camera, and on the right side, plus an overhead
light and a background light.

Originally, I was using a cheap hardware DMX controller for this with
a bunch of sliders, but I ended up writing this instead.  I have a
Raspberry Pi with a cheap USB DMX controller, and I'm using a
Loupedeck (I've used both a Loupedeck Live and a Loupedeck CT) for the
UI.

This repository contains the code that makes the whole thing work.  It
knows about my lights and gives me a simple interface for changing
brightness and color, plus turning the lights all off or on to a
specific default.

This is (currently) very specific to my needs, but could be used as an
example for how to use my Loupedeck library, and it'd be easy enough
to adapt if anyone wants to build anything similar for themselves.



