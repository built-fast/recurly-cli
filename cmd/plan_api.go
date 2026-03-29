package cmd

import (
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
