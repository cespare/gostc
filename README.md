# gostc

gostc is a Go [StatsD](https://github.com/etsy/statsd/) client. It is specifically designed to work with
[gost](https://github.com/cespare/gost): it doesn't support delta gauge values and it includes first-class
support for [counter forwarding](https://github.com/cespare/gost#counter-forwarding).

## Other implementations

* [github.com/peterbourgon/g2s](https://github.com/peterbourgon/g2s)
* [github.com/cactus/go-statsd-client](https://github.com/cactus/go-statsd-client)

Compared with these, gostc tries to:

* Support [gost](https://github.com/cespare/gost) features
* Support most StatsD features, including sending multiple stats in a single UDP packet
* Add minimal overhead to your program by being very fast
* Avoid over-abstraction or solving any problems that haven't presented themselves
