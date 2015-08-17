/*Package eventual is a simple event dispatcher for golang base on channels.
The goal is to pass data arond, without depending to any package except for this
one.

There is some definition :

- Mandatory means publish is not finished unless all subscribers receive the message,
in this case, the receiver must always read the channel. TODO : a way to unsub
- Exclusive means there is only and only one receiver for a topic is available

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

*/
package eventual

import (
	"fmt"
	"sync"
)

// IEvent is an interface to handle one instance of an event in system.
// each IEvent must contain the topic and data that passed to this event
type IEvent interface {
	// Data is the event data that publisherdecide the type of it
	Data() interface{}
}

// Publisher is an interface for publishing event in a system
type Publisher interface {
	// Pub is for publishing an event. it's panic prof
	Pub(IEvent)
}

// Subscriber is an interface to handle subscribers
type Subscriber interface {
	// Sub is for getting an channel to read the events from it.
	// if the event is exclusive, and there is a subscriber, then
	// it panic.
	Sub() <-chan IEvent
}

// Event is the actual event, must register to get one of this.
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
	// // GetTimeout return the timeout for this even. if the timeout is equal or less
	// // than zero, there is no wait, if not the event wait this duration for the
	// // subscriber to pick.
	// GetTimeout() time.Duration
}

// Eventual is an event bus
type Eventual interface {
	Register(topic string, mandatory, exclusive bool) (Event, error)
}

type eventInstance struct {
	topic     string
	mandatory bool
	exclusive bool
	subs      []chan IEvent

	lock *sync.RWMutex
}

type eventualInstance struct {
	list map[string]Event
	lock *sync.Mutex
}

func (e *eventInstance) Pub(ei IEvent) {
	if e.mandatory {
		e.pubMandatory(ei)
	} else {
		e.pubNormal(ei)
	}
}

func (e *eventInstance) pubMandatory(ei IEvent) {
	wg := sync.WaitGroup{}

	wg.Add(len(e.subs))
	for i := range e.subs {
		go func(c chan IEvent) {
			e.lock.RLock()
			defer e.lock.RUnlock()
			defer wg.Done()
			c <- ei
		}(e.subs[i])
	}
	// In mandatory, its blocked until all data are recieved
	wg.Wait()
}

func (e *eventInstance) pubNormal(ei IEvent) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	for i := range e.subs {
		select {
		case e.subs[i] <- ei:
		default:
		}
	}
}

// Sub implementation of the Subscriber interface
func (e *eventInstance) Sub() <-chan IEvent {
	e.lock.Lock()
	defer e.lock.Unlock()

	if e.IsExclusive() && len(e.subs) > 0 {
		panic(fmt.Errorf("this is a exclusive event, and there is already a subscriber available"))
	}
	res := make(chan IEvent)
	e.subs = append(e.subs, res)

	return res
}

func (e *eventInstance) IsMandatory() bool {
	return e.mandatory
}

func (e *eventInstance) IsExclusive() bool {
	return e.exclusive
}

func (e *eventInstance) GetTopic() string {
	return e.topic
}

// Register and event in system, the topic is the event topic, mandatory means
// if the event is hanged until the subscriber gets it. exclusive means the event
// can only and only one subscriber.
// The parameters must be exactly same when calling this several times for a topic
// unless yo get an error.
func (e *eventualInstance) Register(topic string, mandatory, exclusive bool) (Event, error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	event, ok := e.list[topic]
	if ok {
		if event.IsMandatory() != mandatory || event.IsExclusive() != exclusive {
			err := fmt.Errorf(
				"the topic mandatory is %t exclusive is %t, requested topic is respectively %t and %t",
				event.IsMandatory(),
				event.IsExclusive(),
				mandatory,
				exclusive,
			)
			return nil, err
		}
	} else {
		event = &eventInstance{topic, mandatory, exclusive, nil, &sync.RWMutex{}}
		e.list[topic] = event
	}

	return event, nil
}

// New return an eventual structure
func New() Eventual {
	return &eventualInstance{make(map[string]Event), &sync.Mutex{}}
}
