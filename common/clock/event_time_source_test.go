package clock_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/server/common/clock"
)

// event is a helper to verify how many times a callback was triggered. Because callbacks are triggered synchronously
// with calls to EventTimeSource.Advance, we don't need any further synchronization.
type event struct {
	t     *testing.T
	count int
}

// Fire is the callback to be triggered.
func (e *event) Fire() {
	e.count++
}

// AssertFiredOnce asserts that the callback was triggered exactly once.
func (e *event) AssertFiredOnce(msg string) {
	e.t.Helper()
	assert.Equal(e.t, 1, e.count, msg)
}

// AssertFired asserts that the callback was triggered a certain number of times.
func (e *event) AssertFired(n int, msg string) {
	e.t.Helper()
	assert.Equal(e.t, n, e.count, msg)
}

// AssertNotFired asserts that the callback was not triggered.
func (e *event) AssertNotFired(msg string) {
	e.t.Helper()
	assert.Zero(e.t, e.count, msg)
}

func ExampleEventTimeSource() {
	// Create a new fake timeSource.
	source := clock.NewEventTimeSource()

	// Create a timer which fires after 1 second.
	source.AfterFunc(time.Second, func() {
		fmt.Println("timer fired")
	})

	// Advance the time source by 1 second.
	fmt.Println("advancing time source by 1 second")
	source.Advance(time.Second)
	fmt.Println("time source advanced")

	// Output:
	// advancing time source by 1 second
	// timer fired
	// time source advanced
}

func TestEventTimeSource_Since(t *testing.T) {
	t.Parallel()

	// Create a new fake time source.
	source := clock.NewEventTimeSource()

	// No delta expected yet
	start := source.Now()
	assert.Equal(t, time.Duration(0), source.Since(start))

	// Advance by one
	source.Advance(1 * time.Second)
	assert.Equal(t, 1*time.Second, source.Since(start))

	// Advance back to start
	source.Advance(-1 * time.Second)
	assert.Equal(t, time.Duration(0), source.Since(start))
}

func TestEventTimeSource_AfterFunc(t *testing.T) {
	t.Parallel()

	// Create a new fake time source and an event to fire.
	source := clock.NewEventTimeSource()
	ev := event{t: t}

	// Create a timer which fires after 2ns.
	source.AfterFunc(2, ev.Fire)

	// Advance the time source by 1ns.
	source.Advance(1)
	ev.AssertNotFired(
		"Advancing the time source should not fire the timer if its deadline still has not been reached",
	)

	// Advance the time source by 1ns more.
	source.Advance(1)
	ev.AssertFiredOnce("Advancing a time source past a timer's deadline should fire the timer")
}

func TestEventTimeSource_AfterFunc_Reset(t *testing.T) {
	t.Parallel()

	// Create a new fake time source and two events to fire.
	source := clock.NewEventTimeSource()
	ev1 := event{t: t}
	ev2 := event{t: t}

	// Create a timer for each event which fires after 2ns.
	timer := source.AfterFunc(2, ev1.Fire)
	source.AfterFunc(2, ev2.Fire)

	// Advance the time source by 1ns and verify that neither timer has fired.
	source.Advance(1)
	ev1.AssertNotFired("Timer should not fire before deadline")
	ev2.AssertNotFired("Timer should not fire before deadline")

	// Reset the first timer to fire after an additional 2ns.
	assert.True(t, timer.Reset(2), "`Reset` should return true if the timer was not already stopped")

	// Advance the time source by 1ns and verify that the first timer has not fired but the second timer has.
	source.Advance(1)
	ev1.AssertNotFired("Timer which was reset should not fire after original deadline but before new deadline")
	ev2.AssertFiredOnce("Timer which was not reset should fire after deadline")

	// Advance the time source by 1ns more and verify that the reset timer has fired.
	source.Advance(1)
	ev1.AssertFiredOnce("The reset timer should fire after its new deadline")

	// Reset the first timer and advance the time source past the new deadline to verify that the timer fires again.
	assert.False(t, timer.Reset(1), "`Reset` should return false if the timer was already stopped")
	source.Advance(1)
	ev1.AssertFired(2, "The timer should fire again")
}

func TestEventTimeSource_AfterFunc_Stop(t *testing.T) {
	t.Parallel()

	// Create a new fake time source and two events to fire.
	source := clock.NewEventTimeSource()
	ev1 := event{t: t}
	ev2 := event{t: t}

	// Create a timer for each event which fires after 1ns.
	timer := source.AfterFunc(1, ev1.Fire)
	source.AfterFunc(1, ev2.Fire)

	// Stop the first timer.
	assert.True(t, timer.Stop(), "`Stop` should return true if the timer was not already stopped")

	// Advance the time source by 1ns and verify that the first timer has not fired and the second timer has.
	source.Advance(1)
	ev1.AssertNotFired("A timer should not fire if it was already stopped")
	ev2.AssertFiredOnce("A timer which was not stopped should fire after its deadline")

	// Verify that subsequent calls to `Stop` return false.
	assert.False(t, timer.Stop(), "`Stop` return false if the timer was already stopped")
}

func TestEventTimeSource_AfterFunc_NegativeDelay(t *testing.T) {
	t.Parallel()

	// Create a new fake time source and one event to fire.
	source := clock.NewEventTimeSource()
	ev1 := event{t: t}

	// Create a timer which fires after -1ns. This should fire immediately.
	timer := source.AfterFunc(-1, ev1.Fire)

	// Verify that the timer has fired.
	ev1.AssertFiredOnce("A timer with a negative delay should fire immediately")

	// Verify that the timer is stopped.
	assert.False(t, timer.Stop(), "`Stop` should return false if the timer was already stopped")
}

func TestEventTimeSource_Update(t *testing.T) {
	t.Parallel()

	// Create a new fake time source and two events to fire.
	source := clock.NewEventTimeSource()
	ev1 := event{t: t}
	ev2 := event{t: t}

	// Create a timer for each event which fires after 1ns.
	source.AfterFunc(1, ev1.Fire)
	source.AfterFunc(1, ev2.Fire)

	// Verify that the time source starts at Unix epoch.
	assert.Equal(
		t, time.Unix(0, 0), source.Now(), "The fake time source should start at the unix epoch",
	)

	// Update to move the time source forward by 1ns.
	source.Update(time.Unix(0, 1))
	assert.Equal(t, time.Unix(0, 1), source.Now())
	ev1.AssertFiredOnce("Timer should fire after deadline")
	ev2.AssertFiredOnce("Timer should fire after deadline")
}

func TestEventTimeSource_NewTimerWithChannelAndReset(t *testing.T) {
	t.Parallel()

	source := clock.NewEventTimeSource()

	ch, timer := source.NewTimer(time.Second)
	expectedFireTime := source.Now().Add(time.Second)

	select {
	case <-ch:
		t.Error("shouldn't fire yet")
	default:
	}

	source.Advance(2 * time.Second)

	// Since the timer duration was 1s, it should send the time at which the timer fired (which was 1s ago) on the channel
	select {
	case result := <-ch:
		assert.Equal(t, expectedFireTime, result)
	default:
		t.Error("should have fired")
	}

	// Reset the timer so that it fires in 1 second
	timer.Reset(time.Second)
	expectedFireTime = source.Now().Add(time.Second)

	source.Advance(2 * time.Second)

	// Check that the timer sends the time at which it fired on the channel
	select {
	case result := <-ch:
		assert.Equal(t, expectedFireTime, result)
	default:
		t.Error("should have fired")
	}
}
