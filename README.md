ps2avglogin
===========

ps2avglogin is a simple tracker that scans login and logout events in [PlanetSide 2][ps2] using the [Daybreak Census API][census]. It keeps track of every player that logs in after it is started, and then calculates a rolling average time spent logged in every time a tracked player logs out. It also has a few related features, such as the ability to filter short play sessions.

Installation
------------

To install, first you will need a working [Go toolchain][go] of at least version 1.6. Next you will need to set up your [GOPATH][gopath]. Once this is done, simply run

> go get github.com/DeedleFake/ps2avglogin

to install ps2avglogin.

Usage
-----

For usage information, simple run `ps2avglogin -help`. Just running `ps2avglogin` should be good enough for most use cases.

Authors
-------

* DeedleFake

[ps2]: http://www.planetside2.com
[census]: http://census.daybreakgames.com

[go]: https://www.golang.org
[gopath]: https://blog.golang.org/organizing-go-code
