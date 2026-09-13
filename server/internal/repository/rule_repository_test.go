package repository

import (
	"context"
	"fmt"
	"time"

	"expenses/internal/config"
	"expenses/internal/models"
	databasemanager "expenses/pkg/database/manager"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuleRepository", func() {
	var (
		ctx    context.Context
		cfg    *config.Config
		db     databasemanager.DatabaseManager
		repo   RuleRepositoryInterface
		userID int64
	)

	BeforeEach(func() {
		var err error
		cfg, err = config.NewConfig()
		Expect(err).NotTo(HaveOccurred())

		db, err = databasemanager.NewDatabaseManager(cfg)
		Expect(err).NotTo(HaveOccurred())

		repo = NewRuleRepository(db, cfg)
		ctx = context.Background()
		userID = createRepoTestUser(ctx, db, cfg.DBSchema)
	})

	AfterEach(func() {
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.transaction WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.account WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.rule WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.user WHERE id = $1", cfg.DBSchema), userID)
		Expect(db.Close()).To(Succeed())
	})

	createRule := func(name string, description *string) models.RuleResponse {
		rule, err := repo.CreateRule(ctx, models.CreateBaseRuleRequest{
			Name:          name,
			Description:   description,
			EffectiveFrom: time.Now().UTC().Truncate(time.Second),
			CreatedBy:     userID,
		})
		Expect(err).NotTo(HaveOccurred())
		return rule
	}

	createAction := func(ruleID int64) models.RuleActionResponse {
		actions, err := repo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
			{ActionType: models.RuleFieldCategory, ActionValue: "Food", RuleId: ruleID},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(actions).To(HaveLen(1))
		return actions[0]
	}

	createCondition := func(ruleID int64) models.RuleConditionResponse {
		conditions, err := repo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
			{ConditionType: models.RuleFieldName, ConditionValue: "swiggy", ConditionOperator: models.OperatorContains, RuleId: ruleID},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(conditions).To(HaveLen(1))
		return conditions[0]
	}

	Describe("CreateRule and GetRule", func() {
		It("persists and returns a rule", func() {
			description := "salary credits"
			rule := createRule("Round trip", &description)

			Expect(rule.Id).NotTo(BeZero())
			Expect(rule.Name).To(Equal("Round trip"))
			Expect(rule.CreatedBy).To(Equal(userID))
			Expect(rule.ConditionLogic).To(Equal(models.ConditionLogicAnd))

			fetched, err := repo.GetRule(ctx, rule.Id, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetched.Id).To(Equal(rule.Id))
			Expect(fetched.Name).To(Equal("Round trip"))
			Expect(fetched.Description).NotTo(BeNil())
			Expect(*fetched.Description).To(Equal(description))
		})

		It("returns not found for an unknown id", func() {
			_, err := repo.GetRule(ctx, 999999, userID)
			expectAuthErrorType(err, "RuleNotFound")
		})

		It("hides rules owned by another user", func() {
			rule := createRule("Private rule", nil)
			otherUserID := createRepoTestUser(ctx, db, cfg.DBSchema)
			defer func() {
				_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.user WHERE id = $1", cfg.DBSchema), otherUserID)
			}()

			_, err := repo.GetRule(ctx, rule.Id, otherUserID)
			expectAuthErrorType(err, "RuleNotFound")
		})

		It("wraps insert failures", func() {
			_, err := repo.CreateRule(ctx, models.CreateBaseRuleRequest{
				Name:          "Orphan rule",
				EffectiveFrom: time.Now(),
				CreatedBy:     999999999,
			})
			expectAuthErrorType(err, "ruleRepository")
		})
	})

	Describe("CreateRuleActions and CreateRuleConditions", func() {
		It("inserts and lists rule actions", func() {
			rule := createRule("With actions", nil)

			actions, err := repo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
				{ActionType: models.RuleFieldCategory, ActionValue: "Food", RuleId: rule.Id},
				{ActionType: models.RuleFieldTransfer, ActionValue: "true", RuleId: rule.Id},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(actions).To(HaveLen(2))

			listed, err := repo.ListRuleActionsByRuleId(ctx, rule.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(listed).To(HaveLen(2))
			for _, action := range listed {
				Expect(action.RuleId).To(Equal(rule.Id))
			}
		})

		It("returns an insert error when the rule does not exist", func() {
			_, err := repo.CreateRuleActions(ctx, []models.CreateRuleActionRequest{
				{ActionType: models.RuleFieldCategory, ActionValue: "Food", RuleId: 999999},
			})
			expectAuthErrorType(err, "RuleActionInsert")
		})

		It("inserts and lists rule conditions", func() {
			rule := createRule("With conditions", nil)

			conditions, err := repo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
				{ConditionType: models.RuleFieldName, ConditionValue: "swiggy", ConditionOperator: models.OperatorContains, RuleId: rule.Id},
				{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorGreater, RuleId: rule.Id},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(conditions).To(HaveLen(2))

			listed, err := repo.ListRuleConditionsByRuleId(ctx, rule.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(listed).To(HaveLen(2))
			for _, condition := range listed {
				Expect(condition.RuleId).To(Equal(rule.Id))
			}
		})

		It("returns an insert error when the rule does not exist", func() {
			_, err := repo.CreateRuleConditions(ctx, []models.CreateRuleConditionRequest{
				{ConditionType: models.RuleFieldName, ConditionValue: "swiggy", ConditionOperator: models.OperatorContains, RuleId: 999999},
			})
			expectAuthErrorType(err, "RuleConditionInsert")
		})
	})

	Describe("ListRules", func() {
		It("paginates and counts rules", func() {
			createRule("Rule A", nil)
			createRule("Rule B", nil)
			createRule("Rule C", nil)

			pageOne, err := repo.ListRules(ctx, userID, models.RuleListQuery{Page: 1, PageSize: 2})
			Expect(err).NotTo(HaveOccurred())
			Expect(pageOne.Total).To(Equal(3))
			Expect(pageOne.Rules).To(HaveLen(2))
			Expect(pageOne.Page).To(Equal(1))
			Expect(pageOne.PageSize).To(Equal(2))

			pageTwo, err := repo.ListRules(ctx, userID, models.RuleListQuery{Page: 2, PageSize: 2})
			Expect(err).NotTo(HaveOccurred())
			Expect(pageTwo.Rules).To(HaveLen(1))
		})

		It("returns every rule when page size is zero", func() {
			createRule("Rule A", nil)
			createRule("Rule B", nil)

			all, err := repo.ListRules(ctx, userID, models.RuleListQuery{Page: 1, PageSize: 0})
			Expect(err).NotTo(HaveOccurred())
			Expect(all.Total).To(Equal(2))
			Expect(all.Rules).To(HaveLen(2))
		})

		It("filters by name and description", func() {
			description := "auto categorise salary credits"
			createRule("Salary income", &description)
			createRule("Groceries", nil)

			nameSearch := "salary"
			byName, err := repo.ListRules(ctx, userID, models.RuleListQuery{Page: 1, PageSize: 10, Search: &nameSearch})
			Expect(err).NotTo(HaveOccurred())
			Expect(byName.Total).To(Equal(1))
			Expect(byName.Rules[0].Name).To(Equal("Salary income"))

			descriptionSearch := "salary credits"
			byDescription, err := repo.ListRules(ctx, userID, models.RuleListQuery{Page: 1, PageSize: 10, Search: &descriptionSearch})
			Expect(err).NotTo(HaveOccurred())
			Expect(byDescription.Total).To(Equal(1))
			Expect(byDescription.Rules[0].Name).To(Equal("Salary income"))
		})

		It("ignores an empty search term", func() {
			createRule("Rule A", nil)
			emptySearch := ""

			result, err := repo.ListRules(ctx, userID, models.RuleListQuery{Page: 1, PageSize: 10, Search: &emptySearch})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(1))
		})
	})

	Describe("UpdateRule", func() {
		It("updates name, description and condition logic", func() {
			rule := createRule("Before", nil)
			name := "After"
			description := "updated description"
			logic := models.ConditionLogicOr

			updated, err := repo.UpdateRule(ctx, rule.Id, userID, models.UpdateRuleRequest{
				Name:           &name,
				Description:    &description,
				ConditionLogic: &logic,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.Name).To(Equal(name))
			Expect(updated.Description).NotTo(BeNil())
			Expect(*updated.Description).To(Equal(description))
			Expect(updated.ConditionLogic).To(Equal(models.ConditionLogicOr))
		})

		It("returns not found for an unknown rule", func() {
			name := "Missing"
			_, err := repo.UpdateRule(ctx, 999999, userID, models.UpdateRuleRequest{Name: &name})
			expectAuthErrorType(err, "RuleNotFound")
		})

		It("rejects an empty update", func() {
			rule := createRule("Empty update", nil)
			_, err := repo.UpdateRule(ctx, rule.Id, userID, models.UpdateRuleRequest{})
			expectAuthErrorType(err, "NoFieldsToUpdate")
		})
	})

	Describe("UpdateRuleAction", func() {
		It("updates an action", func() {
			rule := createRule("Action update", nil)
			action := createAction(rule.Id)
			newValue := "Travel"

			updated, err := repo.UpdateRuleAction(ctx, action.Id, rule.Id, models.UpdateRuleActionRequest{ActionValue: &newValue})
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.ActionValue).To(Equal(newValue))
		})

		It("returns not found for an unknown action", func() {
			rule := createRule("Action missing", nil)
			newValue := "Travel"

			_, err := repo.UpdateRuleAction(ctx, 999999, rule.Id, models.UpdateRuleActionRequest{ActionValue: &newValue})
			expectAuthErrorType(err, "RuleActionNotFound")
		})

		It("rejects an empty update", func() {
			rule := createRule("Action empty", nil)
			action := createAction(rule.Id)

			_, err := repo.UpdateRuleAction(ctx, action.Id, rule.Id, models.UpdateRuleActionRequest{})
			expectAuthErrorType(err, "NoFieldsToUpdate")
		})
	})

	Describe("UpdateRuleCondition", func() {
		It("updates a condition", func() {
			rule := createRule("Condition update", nil)
			condition := createCondition(rule.Id)
			newValue := "zomato"

			updated, err := repo.UpdateRuleCondition(ctx, condition.Id, rule.Id, models.UpdateRuleConditionRequest{ConditionValue: &newValue})
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.ConditionValue).To(Equal(newValue))
		})

		It("returns not found for an unknown condition", func() {
			rule := createRule("Condition missing", nil)
			newValue := "zomato"

			_, err := repo.UpdateRuleCondition(ctx, 999999, rule.Id, models.UpdateRuleConditionRequest{ConditionValue: &newValue})
			expectAuthErrorType(err, "RuleConditionNotFound")
		})

		It("rejects an empty update", func() {
			rule := createRule("Condition empty", nil)
			condition := createCondition(rule.Id)

			_, err := repo.UpdateRuleCondition(ctx, condition.Id, rule.Id, models.UpdateRuleConditionRequest{})
			expectAuthErrorType(err, "NoFieldsToUpdate")
		})
	})

	Describe("DeleteRuleActionsByRuleId and DeleteRuleConditionsByRuleId", func() {
		It("deletes actions and reports a missing rule afterwards", func() {
			rule := createRule("Delete actions", nil)
			createAction(rule.Id)

			Expect(repo.DeleteRuleActionsByRuleId(ctx, rule.Id)).To(Succeed())

			listed, err := repo.ListRuleActionsByRuleId(ctx, rule.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(listed).To(BeEmpty())

			err = repo.DeleteRuleActionsByRuleId(ctx, rule.Id)
			expectAuthErrorType(err, "RuleNotFound")
		})

		It("deletes conditions and reports a missing rule afterwards", func() {
			rule := createRule("Delete conditions", nil)
			createCondition(rule.Id)

			Expect(repo.DeleteRuleConditionsByRuleId(ctx, rule.Id)).To(Succeed())

			listed, err := repo.ListRuleConditionsByRuleId(ctx, rule.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(listed).To(BeEmpty())

			err = repo.DeleteRuleConditionsByRuleId(ctx, rule.Id)
			expectAuthErrorType(err, "RuleNotFound")
		})
	})

	Describe("DeleteRule", func() {
		It("deletes a rule and cascades its children", func() {
			rule := createRule("Delete me", nil)
			createAction(rule.Id)
			createCondition(rule.Id)

			Expect(repo.DeleteRule(ctx, rule.Id, userID)).To(Succeed())

			_, err := repo.GetRule(ctx, rule.Id, userID)
			expectAuthErrorType(err, "RuleNotFound")

			err = repo.DeleteRule(ctx, rule.Id, userID)
			expectAuthErrorType(err, "RuleNotFound")
		})

		It("refuses to delete another user's rule", func() {
			rule := createRule("Not yours", nil)
			otherUserID := createRepoTestUser(ctx, db, cfg.DBSchema)
			defer func() {
				_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.user WHERE id = $1", cfg.DBSchema), otherUserID)
			}()

			err := repo.DeleteRule(ctx, rule.Id, otherUserID)
			expectAuthErrorType(err, "RuleNotFound")
		})
	})

	Describe("PutRuleActions and PutRuleConditions", func() {
		It("replaces existing actions", func() {
			rule := createRule("Put actions", nil)

			first, err := repo.PutRuleActions(ctx, rule.Id, []models.CreateRuleActionRequest{
				{ActionType: models.RuleFieldCategory, ActionValue: "Food"},
				{ActionType: models.RuleFieldTransfer, ActionValue: "true"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(first).To(HaveLen(2))

			second, err := repo.PutRuleActions(ctx, rule.Id, []models.CreateRuleActionRequest{
				{ActionType: models.RuleFieldCategory, ActionValue: "Travel"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(second).To(HaveLen(1))

			listed, err := repo.ListRuleActionsByRuleId(ctx, rule.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(listed).To(HaveLen(1))
			Expect(listed[0].ActionValue).To(Equal("Travel"))
		})

		It("fails to put actions for a missing rule", func() {
			_, err := repo.PutRuleActions(ctx, 999999, []models.CreateRuleActionRequest{
				{ActionType: models.RuleFieldCategory, ActionValue: "Food"},
			})
			expectAuthErrorType(err, "RuleActionInsert")
		})

		It("replaces existing conditions", func() {
			rule := createRule("Put conditions", nil)

			first, err := repo.PutRuleConditions(ctx, rule.Id, []models.CreateRuleConditionRequest{
				{ConditionType: models.RuleFieldName, ConditionValue: "swiggy", ConditionOperator: models.OperatorContains},
				{ConditionType: models.RuleFieldAmount, ConditionValue: "100", ConditionOperator: models.OperatorGreater},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(first).To(HaveLen(2))

			second, err := repo.PutRuleConditions(ctx, rule.Id, []models.CreateRuleConditionRequest{
				{ConditionType: models.RuleFieldName, ConditionValue: "zomato", ConditionOperator: models.OperatorEquals},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(second).To(HaveLen(1))

			listed, err := repo.ListRuleConditionsByRuleId(ctx, rule.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(listed).To(HaveLen(1))
			Expect(listed[0].ConditionValue).To(Equal("zomato"))
		})

		It("fails to put conditions for a missing rule", func() {
			_, err := repo.PutRuleConditions(ctx, 999999, []models.CreateRuleConditionRequest{
				{ConditionType: models.RuleFieldName, ConditionValue: "swiggy", ConditionOperator: models.OperatorContains},
			})
			expectAuthErrorType(err, "RuleConditionInsert")
		})
	})

	Describe("CreateRuleTransactionMapping", func() {
		newTransaction := func() int64 {
			accountID := createRepoTestAccount(ctx, db, cfg.DBSchema, userID)
			var transactionID int64
			Expect(db.FetchOne(ctx,
				fmt.Sprintf("INSERT INTO %s.transaction (name, amount, date, account_id, created_by) VALUES ($1, $2, $3, $4, $5) RETURNING id", cfg.DBSchema),
				"Mapping transaction", 12.5, time.Now().UTC(), accountID, userID,
			).Scan(&transactionID)).To(Succeed())
			return transactionID
		}

		It("links a rule to a transaction once", func() {
			rule := createRule("Mapping rule", nil)
			transactionID := newTransaction()

			Expect(repo.CreateRuleTransactionMapping(ctx, rule.Id, transactionID)).To(Succeed())
			Expect(repo.CreateRuleTransactionMapping(ctx, rule.Id, transactionID)).To(Succeed())

			var count int
			Expect(db.FetchOne(ctx,
				fmt.Sprintf("SELECT COUNT(*) FROM %s.rule_transaction_mapping WHERE rule_id = $1 AND transaction_id = $2", cfg.DBSchema),
				rule.Id, transactionID,
			).Scan(&count)).To(Succeed())
			Expect(count).To(Equal(1))
		})

		It("returns an error when the rule does not exist", func() {
			transactionID := newTransaction()
			err := repo.CreateRuleTransactionMapping(ctx, 999999, transactionID)
			expectAuthErrorType(err, "ruleRepository")
		})
	})
})
