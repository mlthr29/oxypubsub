package broker_test

import (
	"testing"
	"time"

	broker "oxypubsub/src"
)

func TestCreateBroker(t *testing.T) {
	b := broker.CreateBroker()
	if b == nil {
		t.Fatal("Expected: Broker. Actual: nil.")
	}
}

func TestSubscribe(t *testing.T) {
	b := broker.CreateBroker()
	ch := b.Subscribe()

	if ch == nil {
		t.Fatal("Expected: channel. Actual: nil.")
	}
}

func TestUnsubscribe(t *testing.T) {
	b := broker.CreateBroker()
	ch := b.Subscribe()

	b.Unsubscribe(ch)

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("Expected: channel close. Actual: channel open.")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected: channel close. Actual: channel open.")
	}
}

func TestRegisterPublisher(t *testing.T) {
	b := broker.CreateBroker()
	ch := b.RegisterPublisher()

	if ch == nil {
		t.Fatal("Expected: a channel. Actual: nil.")
	}
}

func TestDeregisterPublisher(t *testing.T) {
	b := broker.CreateBroker()
	ch := b.RegisterPublisher()

	b.DeregisterPublisher(ch)

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("Expected: closed channel. Actual: open channel.")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected: closed channel. Actual: open channel.")
	}
}

func TestPublish(t *testing.T) {
	b := broker.CreateBroker()
	subCh := b.Subscribe()

	message := "test message"
	b.Publish(message)

	select {
	case received := <-subCh:
		if received != message {
			t.Fatalf("Expected: message %q. Actual: %q", message, received)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected: message received. Actual: message not received.")
	}
}

func TestPublisherNotifications(t *testing.T) {
	b := broker.CreateBroker()
	pubCh := b.RegisterPublisher()

	b.Subscribe()

	select {
	case notification := <-pubCh:
		if notification != "A NEW SUBSCRIBER JOINED!" {
			t.Fatalf("Expected: new subscriber notification. Actual: %q", notification)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected: notification on new subscriber joining. Actual: none")
	}
}

func TestPublisherNotificationNoSubscribers(t *testing.T) {
	b := broker.CreateBroker()
	pubCh := b.RegisterPublisher()

	subCh := b.Subscribe()

	select {
	case <-pubCh:
	case <-time.After(100 * time.Millisecond):
	}

	b.Unsubscribe(subCh)

	select {
	case notification := <-pubCh:
		if notification != "NO SUBSCRIBERS AVAILABLE!" {
			t.Fatalf("Expected: notification on no subscribers left. Actual: %q", notification)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected: notification for no subscribers connected. Actual: none.")
	}
}
