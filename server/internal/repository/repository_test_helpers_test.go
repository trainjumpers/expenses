package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	customErrors "expenses/internal/errors"
	databasemanager "expenses/pkg/database/manager"

	. "github.com/onsi/gomega"
)

func createRepoTestUser(ctx context.Context, db databasemanager.DatabaseManager, schema string) int64 {
	var id int64
	Expect(db.FetchOne(ctx,
		fmt.Sprintf("INSERT INTO %s.user (name, email, password) VALUES ($1, $2, $3) RETURNING id", schema),
		"Repository Test User",
		fmt.Sprintf("repository-test-%d@example.com", time.Now().UnixNano()),
		"password",
	).Scan(&id)).To(Succeed())
	return id
}

func createRepoTestAccount(ctx context.Context, db databasemanager.DatabaseManager, schema string, userID int64) int64 {
	var id int64
	Expect(db.FetchOne(ctx,
		fmt.Sprintf("INSERT INTO %s.account (name, bank_type, currency, created_by) VALUES ($1, $2, $3, $4) RETURNING id", schema),
		"Repository Test Account", "others", "inr", userID,
	).Scan(&id)).To(Succeed())
	return id
}

func expectAuthErrorType(err error, errorType string) {
	Expect(err).To(HaveOccurred())
	var authErr *customErrors.AuthError
	Expect(errors.As(err, &authErr)).To(BeTrue(), "expected AuthError, got %T: %v", err, err)
	Expect(authErr.ErrorType).To(Equal(errorType))
}
