package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	recurly "github.com/recurly/recurly-client-go/v5"
)

// PlanAddOnAPI abstracts the Recurly SDK methods used by plan add-on commands,
// allowing tests to inject mocks without making real API calls.
type PlanAddOnAPI interface {
	ListPlanAddOns(planId string, params *recurly.ListPlanAddOnsParams, opts ...recurly.Option) (recurly.AddOnLister, error)
	GetPlanAddOn(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error)
	CreatePlanAddOn(planId string, body *recurly.AddOnCreate, opts ...recurly.Option) (*recurly.AddOn, error)
	UpdatePlanAddOn(planId string, addOnId string, body *recurly.AddOnUpdate, opts ...recurly.Option) (*recurly.AddOn, error)
	RemovePlanAddOn(planId string, addOnId string, opts ...recurly.Option) (*recurly.AddOn, error)
}

// newPlanAddOnAPI is the factory function used by plan add-on commands to get an API client.
// Tests override this to inject mocks.
var newPlanAddOnAPI = func() (PlanAddOnAPI, error) {
	return client.NewClient()
}
