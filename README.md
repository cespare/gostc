# gostc

gostc is a Go [StatsD](https://github.com/etsy/statsd/) client. It is specifically designed to work with
[gost](https://github.com/cespare/gost): it doesn't support delta gauge values and it includes first-class
support for [counter forwarding](https://github.com/cespare/gost#counter-forwarding).

## Installation

    go get github.com/cespare/gostc

## Usage

Quick example:

``` go
client, err := gostc.Dial("localhost:8125")
if err != nil {
  panic(err)
}

// Users will typically ignore the return errors of gostc methods as statsd
// is a best-effort service in most software.
client.Count("foobar", 1, 1)
client.Inc("foobar") // Same as above
t := time.Now()
time.Sleep(time.Second)
client.Time("blah", time.Since(t))
```

See full package documentation on [godoc.org](http://godoc.org/github.com/cespare/gostc).
