package repository

import (
	"context"
	"fmt"
	"time"

	"expenses/internal/config"
	databasemanager "expenses/pkg/database/manager"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SessionRepository", func() {
	var (
		ctx    context.Context
		cfg    *config.Config
		db     databasemanager.DatabaseManager
		repo   SessionRepositoryInterface
		userID int64
	)

	BeforeEach(func() {
		var err error
		cfg, err = config.NewConfig()
		Expect(err).NotTo(HaveOccurred())

		db, err = databasemanager.NewDatabaseManager(cfg)
		Expect(err).NotTo(HaveOccurred())

		repo = NewSessionRepository(db, cfg)
		ctx = context.Background()
		userID = insertSessionTestUser(db, ctx, cfg.DBSchema)
	})

	AfterEach(func() {
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.session WHERE user_id = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.user WHERE id = $1", cfg.DBSchema), userID)
		Expect(db.Close()).To(Succeed())
	})

	It("stores a session and returns it while active", func() {
		Expect(repo.Create(ctx, userID, "hash-active", time.Now().Add(time.Hour))).To(Succeed())

		session, err := repo.GetActiveByHash(ctx, "hash-active")
		Expect(err).NotTo(HaveOccurred())
		Expect(session.UserId).To(Equal(userID))
		Expect(session.TokenHash).To(Equal("hash-active"))
		Expect(session.RevokedAt).To(BeNil())
	})

	It("returns no session for an unknown hash", func() {
		_, err := repo.GetActiveByHash(ctx, "hash-missing")
		Expect(err).To(HaveOccurred())
	})

	It("excludes revoked sessions from active lookups", func() {
		Expect(repo.Create(ctx, userID, "hash-revoke", time.Now().Add(time.Hour))).To(Succeed())

		affected, err := repo.RevokeByHash(ctx, "hash-revoke")
		Expect(err).NotTo(HaveOccurred())
		Expect(affected).To(Equal(int64(1)))

		_, err = repo.GetActiveByHash(ctx, "hash-revoke")
		Expect(err).To(HaveOccurred())

		affected, err = repo.RevokeByHash(ctx, "hash-revoke")
		Expect(err).NotTo(HaveOccurred())
		Expect(affected).To(Equal(int64(0)))
	})

	It("revokes every active session for a user", func() {
		Expect(repo.Create(ctx, userID, "hash-one", time.Now().Add(time.Hour))).To(Succeed())
		Expect(repo.Create(ctx, userID, "hash-two", time.Now().Add(time.Hour))).To(Succeed())

		Expect(repo.RevokeAllForUser(ctx, userID)).To(Succeed())

		for _, hash := range []string{"hash-one", "hash-two"} {
			_, err := repo.GetActiveByHash(ctx, hash)
			Expect(err).To(HaveOccurred())
		}
	})

	It("rotates a session, revoking the old token and activating the new one", func() {
		Expect(repo.Create(ctx, userID, "hash-old", time.Now().Add(time.Hour))).To(Succeed())

		Expect(repo.Rotate(ctx, userID, "hash-old", "hash-new", time.Now().Add(time.Hour))).To(Succeed())

		_, err := repo.GetActiveByHash(ctx, "hash-old")
		Expect(err).To(HaveOccurred())

		session, err := repo.GetActiveByHash(ctx, "hash-new")
		Expect(err).NotTo(HaveOccurred())
		Expect(session.TokenHash).To(Equal("hash-new"))
	})

	It("expires a session and prunes it", func() {
		Expect(repo.Create(ctx, userID, "hash-expire", time.Now().Add(time.Hour))).To(Succeed())

		affected, err := repo.ExpireByHash(ctx, "hash-expire")
		Expect(err).NotTo(HaveOccurred())
		Expect(affected).To(Equal(int64(1)))

		_, err = repo.GetActiveByHash(ctx, "hash-expire")
		Expect(err).To(HaveOccurred())

		deleted, err := repo.DeleteExpired(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(deleted).To(BeNumerically(">=", int64(1)))

		var count int64
		Expect(db.FetchOne(ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM %s.session WHERE token_hash = $1", cfg.DBSchema),
			"hash-expire",
		).Scan(&count)).To(Succeed())
		Expect(count).To(Equal(int64(0)))
	})

	It("reports no rows changed when revoking an unknown hash", func() {
		affected, err := repo.RevokeByHash(ctx, "hash-missing")
		Expect(err).NotTo(HaveOccurred())
		Expect(affected).To(Equal(int64(0)))
	})
})

func insertSessionTestUser(db databasemanager.DatabaseManager, ctx context.Context, schema string) int64 {
	var id int64
	Expect(db.FetchOne(ctx,
		fmt.Sprintf("INSERT INTO %s.user (name, email, password) VALUES ($1,$2,$3) RETURNING id", schema),
		"Session Repository", fmt.Sprintf("session-repo-%d@example.com", time.Now().UnixNano()), "x",
	).Scan(&id)).To(Succeed())
	return id
}
