package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	recurly "github.com/recurly/recurly-client-go/v5"
)

// SubscriptionAPI abstracts the Recurly SDK methods used by subscription commands,
// allowing tests to inject mocks without making real API calls.
type SubscriptionAPI interface {
	ListSubscriptions(params *recurly.ListSubscriptionsParams, opts ...recurly.Option) (recurly.SubscriptionLister, error)
	GetSubscription(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error)
	CreateSubscription(body *recurly.SubscriptionCreate, opts ...recurly.Option) (*recurly.Subscription, error)
	UpdateSubscription(subscriptionId string, body *recurly.SubscriptionUpdate, opts ...recurly.Option) (*recurly.Subscription, error)
	CancelSubscription(subscriptionId string, params *recurly.CancelSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error)
	ReactivateSubscription(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error)
	PauseSubscription(subscriptionId string, body *recurly.SubscriptionPause, opts ...recurly.Option) (*recurly.Subscription, error)
	ResumeSubscription(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error)
	TerminateSubscription(subscriptionId string, params *recurly.TerminateSubscriptionParams, opts ...recurly.Option) (*recurly.Subscription, error)
	ConvertTrial(subscriptionId string, opts ...recurly.Option) (*recurly.Subscription, error)
}

// newSubscriptionAPI is the factory function used by subscription commands to get an API client.
// Tests override this to inject mocks.
var newSubscriptionAPI = func() (SubscriptionAPI, error) {
	return client.NewClient()
}
