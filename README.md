# eventual

[![Build Status](https://travis-ci.org/fzerorubigd/eventual.svg)](https://travis-ci.org/fzerorubigd/eventual)
[![Coverage Status](https://coveralls.io/repos/fzerorubigd/eventual/badge.svg?branch=master&service=github)](https://coveralls.io/github/fzerorubigd/eventual?branch=master)
[![GoDoc](https://godoc.org/github.com/fzerorubigd/eventual?status.svg)](https://godoc.org/github.com/fzerorubigd/eventual)

--
    import "github.com/fzerorubigd/eventual"

Package eventual is a simple event dispatcher for golang base on channels. The
goal is to pass data around, without depending to any package except for this
one.

There is some definition :

- Mandatory means publish is not finished unless all subscribers receive the message, in this case, the receiver must always read the channel. TODO : a way to unsub
- Exclusive means there is only and only one receiver for a topic is available

## Sample code

    package main

    import "github.com/fzerorubigd/eventual"

    func main() {
        e := &eventual.Eventual{}
        // Create a no mandatory, no exclusive topic
        t ,_ := e.Register("the.topic.name", false, false)
        c1 := t.Sub()

        // In any other part of code, even in another package you can create
        // an exactly same topic with exactly same parameter
        t ,_ := e.Register("the.topic.name", false, false)
        t.Pub(SomeIEventStructure{}) // No there is a data in c1, but if there is no reader the data is lost
    }

## TODO

- Support for buffer size (aka pre fetch)

- Support for reply back

## Usage

#### type Event

```go
type Event interface {
	Publisher
	Subscriber

	// GetTopic return the current topic
	GetTopic() string
	// IsMandatory return if this event is mandatory or not , mandatory means
	IsMandatory() bool
	// IsExclusive return if this event is exclusive and there is only one subscriber
	// is allowed
	IsExclusive() bool
}
```

Event is the actual event, must register to get one of this.

#### type Eventual

```go
type Eventual interface {
	Register(topic string, mandatory, exclusive bool) (Event, error)
}
```

Eventual is an event bus

#### func  New

```go
func New() Eventual
```
New return an eventual structure

#### type IEvent

```go
type IEvent interface {
	// Data is the event data that publisherdecide the type of it
	Data() interface{}
}
```

IEvent is an interface to handle one instance of an event in system. each IEvent
must contain the topic and data that passed to this event

#### type Publisher

```go
type Publisher interface {
	// Pub is for publish an event. it's panic prof
	Pub(IEvent)
}
```

Publisher is an interface for publishing event in a system

#### type Subscriber

```go
type Subscriber interface {
	// Sub is for getting an channel to read the events from it.
	// if the event is exclusive, and there is a subscriber, then
	// it panic.
	Sub() <-chan IEvent
}
```

Subscriber is an interface to handle subscribers
