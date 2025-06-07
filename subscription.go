package qilin

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"
)

var (
	// ErrAlreadySubscribed occurs when a subscription already exists
	ErrAlreadySubscribed = errors.New("already subscribed")
	// ErrResourceModificationSubscriptionNotFound occurs when a resource list subscription is not found
	ErrResourceModificationSubscriptionNotFound = errors.New(
		"resource modification subscription not found",
	)
	// ErrResourceListChangeSubscriptionNotFound occurs when a resource list subscription is not found
	ErrResourceListChangeSubscriptionNotFound = errors.New("resource list subscription not found")
)

// Subscription represents a subscription
type Subscription interface {
	// SignalAlive signals that the subscription is still alive.
	SignalAlive()
	// Unsubscribed returns the channel that was closed when this subscription was canceled.
	Unsubscribed() <-chan struct{}
	// LastAliveTime returns the time when the subscription was last checked for health.
	LastAliveTime() time.Time
}

// ResourceListChangeSubscriptionStore perpetuates resource list subscriptions
type ResourceListChangeSubscriptionStore interface {
	// Get retrieves a subscription by session ID
	Get(ctx context.Context, sessionID string) (Subscription, error)
	// Issue issues a new subscription by session ID
	Issue(ctx context.Context, sessionID string) (Subscription, error)
	// Delete removes a subscription by session ID
	Delete(ctx context.Context, sessionID string) error
}

// ResourceModificationSubscriptionStore perpetuates resource modification subscriptions
type ResourceModificationSubscriptionStore interface {
	// Get retrieves a subscription by session ID and URI
	Get(ctx context.Context, sessionID string, uri *url.URL) (Subscription, error)
	// Issue issues a new subscription by session ID and URI
	Issue(ctx context.Context, sessionID string, uri *url.URL) (Subscription, error)
	// Delete removes a subscription
	Delete(ctx context.Context, sessionID string, uri *url.URL) error
	// RetrieveBySessionID retrieves all subscriptions by session ID
	RetrieveBySessionID(ctx context.Context, sessionID string) ([]Subscription, error)
	// RetrieveUnhealthyURIBySessionID retrieves all unhealthy subscriptions uri by session ID
	RetrieveUnhealthyURIBySessionID(ctx context.Context, sessionID string) ([]*url.URL, error)
	// DeleteBySessionID deletes all subscriptions by session ID
	DeleteBySessionID(ctx context.Context, sessionID string) error
}

// compatibility check
var _ Subscription = (*InMemorySubscription)(nil)

// InMemorySubscription is an in-memory implementation of Subscription
//
// NOTE: It will only work properly if Qilin is running on a single server.
type InMemorySubscription struct {
	_                 struct{}
	mu                sync.Mutex
	ctx               context.Context
	cancel            context.CancelFunc
	lastHealthChecked time.Time
}

// SignalAlive See: Subscription#SignalAlive
func (s *InMemorySubscription) SignalAlive() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastHealthChecked = time.Now()
}

// Unsubscribed See: Subscription#Unsubscribed
func (s *InMemorySubscription) Unsubscribed() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ctx.Done()
}

// LastAliveTime See: Subscription#LastAliveTime
func (s *InMemorySubscription) LastAliveTime() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastHealthChecked
}

// compatibility check
var _ ResourceListChangeSubscriptionStore = (*InMemoryResourceListChangeSubscriptionStore)(nil)

// InMemoryResourceListChangeSubscriptionStore is an in-memory implementation of ResourceListChangeSubscriptionStore
//
// NOTE: It will only work properly if Qilin is running on a single server.
type InMemoryResourceListChangeSubscriptionStore struct {
	_                          struct{}
	subscriptions              sync.Map
	nowFunc                    func() time.Time
	subscriptionHealthInterval time.Duration
}

func (s *InMemoryResourceListChangeSubscriptionStore) Issue(
	_ context.Context,
	sessionID string,
) (Subscription, error) {
	ctx, cancel := context.WithCancel(context.Background())
	subscription := &InMemorySubscription{
		ctx:               ctx,
		cancel:            cancel,
		lastHealthChecked: s.nowFunc(),
	}
	s.subscriptions.Store(sessionID, subscription)
	return subscription, nil
}

func (s *InMemoryResourceListChangeSubscriptionStore) Get(
	_ context.Context,
	sessionID string,
) (Subscription, error) {
	v, loaded := s.subscriptions.Load(sessionID)
	if !loaded {
		return nil, ErrResourceListChangeSubscriptionNotFound
	}
	return v.(*InMemorySubscription), nil
}

func (s *InMemoryResourceListChangeSubscriptionStore) Delete(
	_ context.Context,
	sessionID string,
) error {
	value, loaded := s.subscriptions.LoadAndDelete(sessionID)
	if !loaded {
		return nil
	}
	subscription := value.(*InMemorySubscription)
	subscription.cancel()
	return nil
}

// compatibility check
var _ ResourceModificationSubscriptionStore = (*InMemoryResourceModificationSubscriptionStore)(nil)

// InMemoryResourceModificationSubscriptionStore is an in-memory implementation of ResourceModificationSubscriptionStore
//
// NOTE: It will only work properly if Qilin is running on a single server.
type InMemoryResourceModificationSubscriptionStore struct {
	_                          struct{}
	subscriptions              sync.Map
	nowFunc                    func() time.Time
	subscriptionHealthInterval time.Duration
}

func (s *InMemoryResourceModificationSubscriptionStore) Get(
	_ context.Context,
	sessionID string,
	uri *url.URL,
) (Subscription, error) {
	untypedSessionSubscriptions, ok := s.subscriptions.Load(sessionID)
	if !ok {
		return nil, ErrResourceModificationSubscriptionNotFound
	}
	sessionSubscriptions := untypedSessionSubscriptions.(*sync.Map)

	untypedSubscription, ok := sessionSubscriptions.Load(uri.String())
	if !ok {
		return nil, ErrResourceModificationSubscriptionNotFound
	}
	subscription := untypedSubscription.(*InMemorySubscription)

	select {
	case <-subscription.ctx.Done():
		return nil, ErrResourceModificationSubscriptionNotFound
	default:
		subscription.SignalAlive()
	}
	return subscription, nil
}

func (s *InMemoryResourceModificationSubscriptionStore) Issue(
	_ context.Context,
	sessionID string,
	uri *url.URL,
) (Subscription, error) {
	ctx, cancel := context.WithCancel(context.Background())
	subscription := &InMemorySubscription{
		ctx:               ctx,
		cancel:            cancel,
		lastHealthChecked: s.nowFunc(),
	}
	untypedSessionSubscriptions, _ := s.subscriptions.LoadOrStore(sessionID, &sync.Map{})
	sessionSubscriptions := untypedSessionSubscriptions.(*sync.Map)
	sessionSubscriptions.LoadOrStore(uri.String(), subscription)
	return subscription, nil
}

func (s *InMemoryResourceModificationSubscriptionStore) Delete(
	_ context.Context,
	sessionID string,
	uri *url.URL,
) error {
	untypedSessionSubscriptions, ok := s.subscriptions.Load(sessionID)
	if !ok {
		return nil
	}
	sessionSubscriptions := untypedSessionSubscriptions.(*sync.Map)

	untypedSubscription, ok := sessionSubscriptions.LoadAndDelete(uri.String())
	if !ok {
		return nil
	}
	subscription := untypedSubscription.(*InMemorySubscription)
	subscription.cancel()
	return nil
}

func (s *InMemoryResourceModificationSubscriptionStore) RetrieveBySessionID(
	_ context.Context,
	sessionID string,
) ([]Subscription, error) {
	untypedSessionSubscriptions, ok := s.subscriptions.Load(sessionID)
	if !ok {
		return nil, fmt.Errorf("session subscription '%s' not found", sessionID)
	}
	sessionSubscriptions := untypedSessionSubscriptions.(*sync.Map)

	subscriptions := make([]Subscription, 0)
	sessionSubscriptions.Range(func(k, v interface{}) bool {
		subscription := v.(*InMemorySubscription)
		select {
		case <-subscription.ctx.Done():
			return true
		default:
			subscriptions = append(subscriptions, subscription)
		}
		return true
	})
	return subscriptions, nil
}

func (s *InMemoryResourceModificationSubscriptionStore) RetrieveUnhealthyURIBySessionID(
	_ context.Context,
	sessionID string,
) ([]*url.URL, error) {
	untypedSessionSubscriptions, ok := s.subscriptions.Load(sessionID)
	if !ok {
		return nil, fmt.Errorf("session subscription '%s' not found", sessionID)
	}
	sessionSubscriptions := untypedSessionSubscriptions.(*sync.Map)

	uris := make([]*url.URL, 0)
	sessionSubscriptions.Range(func(k, v interface{}) bool {
		subscription := v.(*InMemorySubscription)
		if s.nowFunc().Sub(subscription.lastHealthChecked) > s.subscriptionHealthInterval {
			resourceURI, err := url.Parse(k.(string))
			if err != nil {
				return true
			}
			uris = append(uris, resourceURI)
		}
		return true
	})
	return uris, nil
}

func (s *InMemoryResourceModificationSubscriptionStore) DeleteBySessionID(
	_ context.Context,
	sessionID string,
) error {
	untypedSessionSubscriptions, ok := s.subscriptions.LoadAndDelete(sessionID)
	if !ok {
		return fmt.Errorf("session subscription '%s' not found", sessionID)
	}
	sessionSubscriptions := untypedSessionSubscriptions.(*sync.Map)

	sessionSubscriptions.Range(func(k, v interface{}) bool {
		subscription := v.(*InMemorySubscription)
		subscription.cancel()
		return true
	})
	return nil
}

// ResourcesSubscriptionManager manages the subscription status of resources
type ResourcesSubscriptionManager interface {
	// SubscribeToResourceModification records that a resource modification subscription has been started
	// and returns a Subscription.
	// if the subscription already exists, it returns the existing subscription.
	//
	//   ## params
	//
	//   - ctx: context
	//   - sessionID: identifier of the session
	//   - resourceURI: the URI of the resource to subscribe to
	//
	//   ## returns
	//
	//   - subscription: a subscription that can be used to check the health of the subscription
	//   - err: an error if the subscription failed
	SubscribeToResourceModification(
		ctx context.Context,
		sessionID string,
		resourceURI *url.URL,
	) (Subscription, error)
	// UnsubscribeToResourceModification records that the resource modification subscription has been ended
	//
	//   ## params
	//
	//   - ctx: context
	//   - sessionID: identifier of the session
	//   - resourceURI: the URI of the resource to unsubscribe to
	//
	//   ## returns
	//
	//   - err: an error if the unsubscription failed
	UnsubscribeToResourceModification(
		ctx context.Context,
		sessionID string,
		resourceURI *url.URL,
	) error
	// UnhealthSubscriptions returns a list of unhealth subscriptions urls
	UnhealthSubscriptions(ctx context.Context, sessionID string) ([]*url.URL, error)
}

// compatibility check
var _ ResourcesSubscriptionManager = (*resourcesSubscribeManager)(nil)

// resourcesSubscribeManager is the simplest implementation of ResourcesSubscriptionManager
type resourcesSubscribeManager struct {
	_                          struct{}
	nowFunc                    func() time.Time
	subscriptionHealthInterval time.Duration

	store ResourceModificationSubscriptionStore
}

func (m *resourcesSubscribeManager) SubscribeToResourceModification(
	ctx context.Context,
	sessionID string,
	resourceURI *url.URL,
) (Subscription, error) {
	subscription, _ := m.store.Get(ctx, sessionID, resourceURI)
	if subscription != nil {
		return subscription, nil
	}
	return m.store.Issue(ctx, sessionID, resourceURI)
}

func (m *resourcesSubscribeManager) UnsubscribeToResourceModification(
	ctx context.Context,
	sessionID string,
	resourceURI *url.URL,
) error {
	if err := m.store.Delete(ctx, sessionID, resourceURI); err != nil {
		return err
	}
	return nil
}

func (m *resourcesSubscribeManager) UnhealthSubscriptions(
	ctx context.Context,
	sessionID string,
) ([]*url.URL, error) {
	return m.store.RetrieveUnhealthyURIBySessionID(ctx, sessionID)
}

// ResourceListChangeSubscriptionManager manages the subscription status of resource lists
type ResourceListChangeSubscriptionManager interface {
	// SubscribeToResourceListChanges records that a resource list subscription has been started
	// and returns a Subscription.
	// if the subscription already exists, it returns the existing subscription.
	//
	//   ## params
	//
	//   - ctx: context
	//   - sessionID: identifier of the session
	//
	//   ## returns
	//
	//   - subscription: a subscription that can be used to check the health of the subscription
	//   - err: an error if the subscription failed
	SubscribeToResourceListChanges(
		ctx context.Context,
		sessionID string,
	) (subscription Subscription, err error)
	// UnsubscribeToResourceListChanges unsubscribes from the resource list changes
	//
	//   ## params
	//
	//   - ctx: context
	//   - sessionID: identifier of the session
	//
	//   ## returns
	//
	//   - err: an error if the unsubscription failed
	UnsubscribeToResourceListChanges(
		ctx context.Context,
		sessionID string,
	) error
	// Health returns true if the resource list subscription is healthy, otherwise false.
	//
	//   ## params
	//
	//   - ctx: context
	//   - sessionID: identifier of the session
	//
	//   ## returns
	//
	//   - healthy: true if the resource list subscription is healthy, otherwise false
	//   - err: an error if the health check failed
	Health(
		ctx context.Context,
		sessionID string,
	) (healthy bool, err error)
}

// compatibility check
var _ ResourceListChangeSubscriptionManager = (*resourceListChangeSubscriptionManager)(nil)

// resourceListChangeSubscriptionManager is the simplest implementation of ResourceListChangeSubscriptionManager
type resourceListChangeSubscriptionManager struct {
	_                          struct{}
	nowFunc                    func() time.Time
	subscriptionHealthInterval time.Duration

	store ResourceListChangeSubscriptionStore
}

func (m *resourceListChangeSubscriptionManager) SubscribeToResourceListChanges(
	ctx context.Context,
	sessionID string,
) (Subscription, error) {
	subscription, err := m.store.Get(ctx, sessionID)
	if err == nil {
		return subscription, nil
	}
	return m.store.Issue(ctx, sessionID)
}

func (m *resourceListChangeSubscriptionManager) UnsubscribeToResourceListChanges(
	ctx context.Context,
	sessionID string,
) error {
	if err := m.store.Delete(ctx, sessionID); err != nil {
		return err
	}
	return nil
}

func (m *resourceListChangeSubscriptionManager) Health(
	ctx context.Context,
	sessionID string,
) (bool, error) {
	subscription, _ := m.store.Get(ctx, sessionID)
	if subscription != nil {
		return false, ErrResourceListChangeSubscriptionNotFound
	}
	return m.nowFunc().Sub(subscription.LastAliveTime()) > m.subscriptionHealthInterval, nil
}
