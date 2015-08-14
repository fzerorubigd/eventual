package eventual

import (
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type mockIEvent struct {
	data int
}

func (m mockIEvent) Data() interface{} {
	return m.data
}

func TestEventual(t *testing.T) {
	Convey("eventual basic functionality, no mandatory, no exclusive", t, func() {
		e := New()

		t, err := e.Register("test.topic", false, false)
		So(err, ShouldBeNil)
		So(t.GetTopic(), ShouldEqual, "test.topic")
		So(t.IsMandatory(), ShouldBeFalse)
		So(t.IsExclusive(), ShouldBeFalse)

		Convey("pub/sub test", func() {
			s1 := t.Sub()
			s2 := t.Sub()
			t.Sub() // This must pass, since its not mandatory
			wg := sync.WaitGroup{}
			wg.Add(2)

			var resp1, resp2 IEvent
			go func() {
				defer wg.Done()
				resp1 = <-s1
			}()

			go func() {
				defer wg.Done()
				resp2 = <-s2
			}()

			// Make sure the readers are ready. I know this is not accurate, but
			// I don't know any beter way
			time.Sleep(time.Second)
			t.Pub(mockIEvent{1})

			wg.Wait()
			So(resp1.Data().(int), ShouldEqual, 1)
			So(resp2.Data().(int), ShouldEqual, 1)
		})

		Convey("pub/sub test on another channel", func() {
			t1, err := e.Register("test.topic", false, false)
			So(err, ShouldBeNil)

			s1 := t.Sub()
			wg := sync.WaitGroup{}
			wg.Add(1)

			var resp IEvent
			go func() {
				defer wg.Done()
				resp = <-s1
			}()
			time.Sleep(time.Second)

			t1.Pub(mockIEvent{2})

			wg.Wait()
			So(resp.Data().(int), ShouldEqual, 2)
		})
	})
}

func TestEventualMandatory(t *testing.T) {
	Convey("eventual basic functionality, mandatory, no exclusive", t, func() {
		e := New()

		tm, err := e.Register("topic.mandatory", true, false)
		So(err, ShouldBeNil)
		So(tm.GetTopic(), ShouldEqual, "topic.mandatory")
		So(tm.IsMandatory(), ShouldBeTrue)
		So(tm.IsExclusive(), ShouldBeFalse)

		s1 := tm.Sub()
		wg := sync.WaitGroup{}

		wg.Add(1)
		var resp IEvent
		go func() {
			defer wg.Done()
			resp = <-s1
		}()

		tm.Pub(mockIEvent{3})
		wg.Wait()
		So(resp.Data().(int), ShouldEqual, 3)
	})
}

func TestEventualExclusive(t *testing.T) {
	Convey("eventual basic functionality,no mandatory, exclusive", t, func() {
		e := New()

		tm, err := e.Register("topic.mandatory", false, true)
		So(err, ShouldBeNil)
		So(tm.GetTopic(), ShouldEqual, "topic.mandatory")
		So(tm.IsMandatory(), ShouldBeFalse)
		So(tm.IsExclusive(), ShouldBeTrue)

		So(func() { tm.Sub() }, ShouldNotPanic)
		So(func() { tm.Sub() }, ShouldPanic)
	})
}

func TestEventualRegister(t *testing.T) {
	Convey("eventual register", t, func() {
		e := New()

		tm, err := e.Register("topic.mandatory", false, true)
		So(err, ShouldBeNil)
		So(tm.GetTopic(), ShouldEqual, "topic.mandatory")
		So(tm.IsMandatory(), ShouldBeFalse)
		So(tm.IsExclusive(), ShouldBeTrue)

		tm2, err := e.Register("topic.mandatory", false, true)
		So(err, ShouldBeNil)
		So(tm2.GetTopic(), ShouldEqual, "topic.mandatory")
		So(tm2.IsMandatory(), ShouldBeFalse)
		So(tm2.IsExclusive(), ShouldBeTrue)

		_, err = e.Register("topic.mandatory", true, false)
		So(err, ShouldNotBeNil)
	})
}
