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

	Describe("E2E-style validation for all field types, operators, and edge cases", func() {
		It("validates all supported action types with valid and invalid values", func() {
			// Amount: valid
			typ := models.RuleFieldAmount
			val := "123.45"
			req := models.UpdateRuleActionRequest{ActionType: &typ, ActionValue: &val}
			Expect(v.ValidateUpdateAction(req)).To(Succeed())

			// Amount: negative
			val = "-10.5"
			req = models.UpdateRuleActionRequest{ActionType: &typ, ActionValue: &val}
			Expect(v.ValidateUpdateAction(req)).To(Succeed())

			// Amount: invalid
			val = "notanumber"
			req = models.UpdateRuleActionRequest{ActionType: &typ, ActionValue: &val}
			Expect(v.ValidateUpdateAction(req)).ToNot(Succeed())

			// Category: valid
			typ = models.RuleFieldCategory
			val = "42"
			req = models.UpdateRuleActionRequest{ActionType: &typ, ActionValue: &val}
			Expect(v.ValidateUpdateAction(req)).To(Succeed())

			// Category: negative
			val = "-1"
			req = models.UpdateRuleActionRequest{ActionType: &typ, ActionValue: &val}
			Expect(v.ValidateUpdateAction(req)).To(Succeed())

			// Category: invalid
			val = "notanint"
			req = models.UpdateRuleActionRequest{ActionType: &typ, ActionValue: &val}
			Expect(v.ValidateUpdateAction(req)).ToNot(Succeed())

			// Name: valid
			typ = models.RuleFieldName
			val = "Some Name"
			req = models.UpdateRuleActionRequest{ActionType: &typ, ActionValue: &val}
			Expect(v.ValidateUpdateAction(req)).To(Succeed())

			// Name: empty
			val = ""
			req = models.UpdateRuleActionRequest{ActionType: &typ, ActionValue: &val}
			Expect(v.ValidateUpdateAction(req)).ToNot(Succeed())

			// Description: valid
			typ = models.RuleFieldDescription
			val = "Some Description"
			req = models.UpdateRuleActionRequest{ActionType: &typ, ActionValue: &val}
			Expect(v.ValidateUpdateAction(req)).To(Succeed())

			// Description: empty
			val = ""
			req = models.UpdateRuleActionRequest{ActionType: &typ, ActionValue: &val}
			Expect(v.ValidateUpdateAction(req)).ToNot(Succeed())

			// Transfer: valid
			typ = models.RuleFieldTransfer
			val = "42"
			req = models.UpdateRuleActionRequest{ActionType: &typ, ActionValue: &val}
			Expect(v.ValidateUpdateAction(req)).To(Succeed())

			// Transfer: invalid (not an integer)
			val = "notanint"
			req = models.UpdateRuleActionRequest{ActionType: &typ, ActionValue: &val}
			Expect(v.ValidateUpdateAction(req)).ToNot(Succeed())
		})

		It("validates all supported condition types and operators", func() {
			// Amount with valid operators
			typ := models.RuleFieldAmount
			val := "100"
			for _, op := range []models.RuleOperator{models.OperatorEquals, models.OperatorGreater, models.OperatorLower} {
				req := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				Expect(v.ValidateUpdateCondition(req)).To(Succeed())
			}
			// Amount with invalid operator
			op := models.OperatorContains
			req := models.UpdateRuleConditionRequest{
				ConditionType:     &typ,
				ConditionValue:    &val,
				ConditionOperator: &op,
			}
			Expect(v.ValidateUpdateCondition(req)).ToNot(Succeed())

			// Name with valid operators
			typ = models.RuleFieldName
			val = "abc"
			for _, op := range []models.RuleOperator{models.OperatorEquals, models.OperatorContains} {
				req := models.UpdateRuleConditionRequest{
					ConditionType:     &typ,
					ConditionValue:    &val,
					ConditionOperator: &op,
				}
				Expect(v.ValidateUpdateCondition(req)).To(Succeed())
			}
			// Name with invalid operator
			op = models.OperatorGreater
			req = models.UpdateRuleConditionRequest{
				ConditionType:     &typ,
				ConditionValue:    &val,
				ConditionOperator: &op,
			}
			Expect(v.ValidateUpdateCondition(req)).ToNot(Succeed())

			// Category with valid operator
			typ = models.RuleFieldCategory
			val = "1"
			op = models.OperatorEquals
			req = models.UpdateRuleConditionRequest{
				ConditionType:     &typ,
				ConditionValue:    &val,
				ConditionOperator: &op,
			}
			Expect(v.ValidateUpdateCondition(req)).To(Succeed())
			// Category with invalid operator
			op = models.OperatorContains
			req = models.UpdateRuleConditionRequest{
				ConditionType:     &typ,
				ConditionValue:    &val,
				ConditionOperator: &op,
			}
			Expect(v.ValidateUpdateCondition(req)).ToNot(Succeed())

			// Transfer with valid operator
			typ = models.RuleFieldTransfer
			val = "1"
			op = models.OperatorEquals
			req = models.UpdateRuleConditionRequest{
				ConditionType:     &typ,
				ConditionValue:    &val,
				ConditionOperator: &op,
			}
			Expect(v.ValidateUpdateCondition(req)).To(Succeed())
			// Transfer with invalid operator
			op = models.OperatorContains
			req = models.UpdateRuleConditionRequest{
				ConditionType:     &typ,
				ConditionValue:    &val,
				ConditionOperator: &op,
			}
			Expect(v.ValidateUpdateCondition(req)).ToNot(Succeed())
		})

		It("validates a rule with multiple actions and conditions of different types", func() {
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
					{
						ActionType:  models.RuleFieldCategory,
						ActionValue: "2",
						RuleId:      1,
					},
					{
						ActionType:  models.RuleFieldName,
						ActionValue: "Some Name",
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
					{
						ConditionType:     models.RuleFieldCategory,
						ConditionValue:    "2",
						ConditionOperator: models.OperatorEquals,
						RuleId:            1,
					},
					{
						ConditionType:     models.RuleFieldName,
						ConditionValue:    "Some Name",
						ConditionOperator: models.OperatorContains,
						RuleId:            1,
					},
				},
			}
			Expect(v.Validate(req)).To(Succeed())
		})

		It("rejects a rule with a mix of valid and invalid actions/conditions", func() {
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
						ActionValue: "notanumber",
						RuleId:      1,
					},
					{
						ActionType:  models.RuleFieldCategory,
						ActionValue: "2",
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
					{
						ConditionType:     models.RuleFieldCategory,
						ConditionValue:    "notanint",
						ConditionOperator: models.OperatorEquals,
						RuleId:            1,
					},
				},
			}
			Expect(v.Validate(req)).ToNot(Succeed())
		})

		It("rejects a rule with missing required fields in nested actions/conditions", func() {
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
						ActionType:  "",
						ActionValue: "",
						RuleId:      1,
					},
				},
				Conditions: []models.CreateRuleConditionRequest{
					{
						ConditionType:     "",
						ConditionValue:    "",
						ConditionOperator: "",
						RuleId:            1,
					},
				},
			}
			Expect(v.Validate(req)).ToNot(Succeed())
		})
	})
})
