package cmd

import (
	"github.com/built-fast/recurly-cli/internal/client"
	recurly "github.com/recurly/recurly-client-go/v5"
)

// PlanAPI abstracts the Recurly SDK methods used by plan commands,
// allowing tests to inject mocks without making real API calls.
type PlanAPI interface {
	ListPlans(params *recurly.ListPlansParams, opts ...recurly.Option) (recurly.PlanLister, error)
	GetPlan(planId string, opts ...recurly.Option) (*recurly.Plan, error)
	CreatePlan(body *recurly.PlanCreate, opts ...recurly.Option) (*recurly.Plan, error)
	UpdatePlan(planId string, body *recurly.PlanUpdate, opts ...recurly.Option) (*recurly.Plan, error)
	RemovePlan(planId string, opts ...recurly.Option) (*recurly.Plan, error)
}

// newPlanAPI is the factory function used by plan commands to get an API client.
// Tests override this to inject mocks.
var newPlanAPI = func() (PlanAPI, error) {
	return client.NewClient()
}
