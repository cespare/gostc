package gostc

import (
	"errors"
	"math/rand"
	"net"
	"strconv"
	"time"
)

// A Client is a StatsD/gost client which has a UDP connection.
type Client struct {
	c *net.UDPConn
}

// Dial creates a new Client with the given UDP address.
func Dial(addr string) (*Client, error) {
	u, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	c, err := net.DialUDP("udp", nil, u)
	if err != nil {
		return nil, err
	}

	client := &Client{
		c: c,
	}
	return client, nil
}

func (c *Client) send(b []byte) error {
	_, err := c.c.Write(b)
	return err
}

// ErrSamplingRate is returned by client.Count (or variants) when a bad sampling rate value is provided.
var ErrSamplingRate = errors.New("Sampling rate must be in (0, 1]")

// Count submits a statsd count message with the given key, value, and sampling rate.
func (c *Client) Count(key string, delta, samplingRate float64) error {
	msg := []byte(key)
	msg = append(msg, ':')
	msg = strconv.AppendFloat(msg, delta, 'f', -1, 64)
	msg = append(msg, []byte("|c")...)
	switch {
	case samplingRate > 1 || samplingRate <= 0:
		return ErrSamplingRate
	case samplingRate == 1:
	default:
		msg = append(msg, '@')
		msg = strconv.AppendFloat(msg, samplingRate, 'f', -1, 64)
	}
	return c.send(msg)
}

// Inc submits a count with delta and sampling rate equal to 1.
func (c *Client) Inc(key string) error { return c.Count(key, 1, 1) }

var randFloat = rand.Float64

// CountProb counts (key, delta) with probability p in (0, 1].
func (c *Client) CountProb(key string, delta, p float64) error {
	if p > 1 || p <= 0 {
		return ErrSamplingRate
	}
	if randFloat() >= p {
		return nil
	}
	return c.Count(key, delta, p)
}

// IncProb increments key with probability p in (0, 1].
func (c *Client) IncProb(key string, p float64) error {
	if p > 1 || p <= 0 {
		return ErrSamplingRate
	}
	if randFloat() >= p {
		return nil
	}
	return c.Count(key, 1, p)
}

// Time submits a statsd timer message.
func (c *Client) Time(key string, duration time.Duration) error {
	msg := []byte(key)
	msg = append(msg, ':')
	msg = strconv.AppendFloat(msg, duration.Seconds()*1000, 'f', -1, 64)
	msg = append(msg, []byte("|ms")...)
	return c.send(msg)
}

// Gauge submits a statsd gauge message.
func (c *Client) Gauge(key string, value float64) error {
	msg := []byte(key)
	msg = append(msg, ':')
	msg = strconv.AppendFloat(msg, value, 'f', -1, 64)
	msg = append(msg, []byte("|g")...)
	return c.send(msg)
}

// Set submits a statsd set message.
func (c *Client) Set(key string, element []byte) error {
	msg := make([]byte, len(key), len(key)+1+len(element)+2)
	copy(msg, key)
	msg = append(msg, ':')
	msg = append(msg, element...)
	msg = append(msg, []byte("|s")...)
	return c.send(msg)
}
