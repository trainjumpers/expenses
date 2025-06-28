package validator

import (
	"expenses/internal/models"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuleValidator", func() {
	var v *RuleValidator

	BeforeEach(func() {
		v = &RuleValidator{}
	})

	Describe("ValidateUpdate", func() {
		It("accepts valid name, description, and effective_from", func() {
			name := "Valid Name"
			desc := "Valid Description"
			now := time.Now()
			req := models.UpdateRuleRequest{
				Name:          &name,
				Description:   &desc,
				EffectiveFrom: &now,
			}
			Expect(v.ValidateUpdate(req)).To(Succeed())
		})

		It("rejects effective_from in the future", func() {
			name := "Valid"
			future := time.Now().Add(24 * time.Hour)
			req := models.UpdateRuleRequest{Name: &name, EffectiveFrom: &future}
			Expect(v.ValidateUpdate(req)).ToNot(Succeed())
		})

		It("accepts nil fields", func() {
			req := models.UpdateRuleRequest{}
			Expect(v.ValidateUpdate(req)).To(Succeed())
		})
	})

	Describe("ValidateUpdateAction", func() {
		It("accepts valid action type and value", func() {
			typ := models.RuleFieldAmount
			val := "123.45"
			req := models.UpdateRuleActionRequest{
				ActionType:  &typ,
				ActionValue: &val,
			}
			Expect(v.ValidateUpdateAction(req)).To(Succeed())
		})

		It("rejects invalid action type", func() {
			typ := models.RuleFieldType("invalid")
			val := "foo"
			req := models.UpdateRuleActionRequest{
				ActionType:  &typ,
				ActionValue: &val,
			}
			Expect(v.ValidateUpdateAction(req)).ToNot(Succeed())
		})

		It("rejects invalid action value for type", func() {
			typ := models.RuleFieldAmount
			val := "not-a-number"
			req := models.UpdateRuleActionRequest{
				ActionType:  &typ,
				ActionValue: &val,
			}
			Expect(v.ValidateUpdateAction(req)).ToNot(Succeed())
		})

		It("accepts nil fields", func() {
			req := models.UpdateRuleActionRequest{}
			Expect(v.ValidateUpdateAction(req)).To(Succeed())
		})
	})

	Describe("ValidateUpdateCondition", func() {
		It("accepts valid condition type, value, and operator", func() {
			typ := models.RuleFieldAmount
			val := "123.45"
			op := models.OperatorGreater
			req := models.UpdateRuleConditionRequest{
				ConditionType:     &typ,
				ConditionValue:    &val,
				ConditionOperator: &op,
			}
			Expect(v.ValidateUpdateCondition(req)).To(Succeed())
		})

		It("rejects invalid condition type", func() {
			typ := models.RuleFieldType("invalid")
			val := "foo"
			op := models.OperatorEquals
			req := models.UpdateRuleConditionRequest{
				ConditionType:     &typ,
				ConditionValue:    &val,
				ConditionOperator: &op,
			}
			Expect(v.ValidateUpdateCondition(req)).ToNot(Succeed())
		})

		It("rejects invalid condition value for type", func() {
			typ := models.RuleFieldAmount
			val := "not-a-number"
			op := models.OperatorEquals
			req := models.UpdateRuleConditionRequest{
				ConditionType:     &typ,
				ConditionValue:    &val,
				ConditionOperator: &op,
			}
			Expect(v.ValidateUpdateCondition(req)).ToNot(Succeed())
		})

		It("rejects invalid operator for type", func() {
			typ := models.RuleFieldAmount
			val := "123.45"
			op := models.OperatorContains // not valid for amount
			req := models.UpdateRuleConditionRequest{
				ConditionType:     &typ,
				ConditionValue:    &val,
				ConditionOperator: &op,
			}
			Expect(v.ValidateUpdateCondition(req)).ToNot(Succeed())
		})

		It("accepts nil fields", func() {
			req := models.UpdateRuleConditionRequest{}
			Expect(v.ValidateUpdateCondition(req)).To(Succeed())
		})
	})

	Describe("Validate (CreateRuleRequest)", func() {
		It("accepts valid create rule request", func() {
			now := time.Now()
			req := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Test",
					Description:   nil,
					EffectiveFrom: now,
					CreatedBy:     1,
				},
				Actions: []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldAmount,
						ActionValue: "123.45",
						RuleId:      1,
					},
				},
				Conditions: []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "123.45",
						ConditionOperator: models.OperatorEquals,
						RuleId:            1,
					},
				},
			}
			Expect(v.Validate(req)).To(Succeed())
		})

		It("rejects missing actions", func() {
			now := time.Now()
			req := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Test",
					Description:   nil,
					EffectiveFrom: now,
					CreatedBy:     1,
				},
				Actions: []models.CreateRuleActionRequest{},
				Conditions: []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "123.45",
						ConditionOperator: models.OperatorEquals,
						RuleId:            1,
					},
				},
			}
			Expect(v.Validate(req)).ToNot(Succeed())
		})

		It("rejects missing conditions", func() {
			now := time.Now()
			req := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Test",
					Description:   nil,
					EffectiveFrom: now,
					CreatedBy:     1,
				},
				Actions: []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldAmount,
						ActionValue: "123.45",
						RuleId:      1,
					},
				},
				Conditions: []models.CreateRuleConditionRequest{},
			}
			Expect(v.Validate(req)).ToNot(Succeed())
		})

		It("rejects invalid effective date", func() {
			zero := time.Time{}
			req := models.CreateRuleRequest{
				Rule: models.CreateBaseRuleRequest{
					Name:          "Test",
					Description:   nil,
					EffectiveFrom: zero,
					CreatedBy:     1,
				},
				Actions: []models.CreateRuleActionRequest{
					{
						ActionType:  models.RuleFieldAmount,
						ActionValue: "123.45",
						RuleId:      1,
					},
				},
				Conditions: []models.CreateRuleConditionRequest{
					{
						ConditionType:     models.RuleFieldAmount,
						ConditionValue:    "123.45",
						ConditionOperator: models.OperatorEquals,
						RuleId:            1,
					},
				},
			}
			Expect(v.Validate(req)).ToNot(Succeed())
		})
	})
})
