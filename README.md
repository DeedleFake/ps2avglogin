ps2avglogin
===========

ps2avglogin is a simple tracker that scans login and logout events in [PlanetSide 2][ps2] using the [Daybreak Census API][census]. It keeps track of every player that logs in after it is started, and then calculates a rolling average time spent logged in every time a tracked player logs out. It also has a few related features, such as the ability to filter short play sessions.

Installation
------------

To install, first you will need a working [Go toolchain][go] that is at least version 1.6. Next you will need to set up your [GOPATH][gopath]. Once this is done, simply run

> go get github.com/DeedleFake/ps2avglogin

to install ps2avglogin.

Docker
------

To build a Docker image containing ps2avglogin, first, build a binary by running the following command in the source directory:

> go build -ldflags '-extldflags "-static"'

The binary will also need access to the SSL root certificates, so copy those to the repository. They are usually located at `/etc/ssl/certs/ca-certificates.crt`, so run

> cp /etc/ssl/certs/ca-certificates.crt .

Then run

> docker build -t ps2avglogin .

to build the image, and you're done.

A quick note: The default working directory for the image is in `/data`, so it may be a good idea to mount `/data` to somewhere on the host system for access to the database and for preserving data when updating.

Usage
-----

For usage information, simply run `ps2avglogin -help`. Just running `ps2avglogin` should be good enough for most use cases.

Authors
-------

* DeedleFake

[ps2]: http://www.planetside2.com
[census]: http://census.daybreakgames.com

[go]: https://www.golang.org
[gopath]: https://blog.golang.org/organizing-go-code
