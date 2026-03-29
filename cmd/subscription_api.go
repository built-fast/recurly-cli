package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	"github.com/built-fast/recurly-cli/internal/output"
	recurly "github.com/recurly/recurly-client-go/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
var newSubscriptionAPI = func(cmd *cobra.Command) (SubscriptionAPI, error) {
	cfg := output.FromContext(cmd.Context())
	return client.NewClient(client.ClientConfig{
		APIKey: viper.GetString("api_key"),
		Region: viper.GetString("region"),
		IsJSON: func() bool { return isJSONFormat(cfg.Format) },
	})
}
