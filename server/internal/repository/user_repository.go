package repository

import (
	"errors"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	commonErrors "expenses/internal/errors"
	"expenses/internal/models"
	database "expenses/pkg/database/manager"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type UserRepositoryInterface interface {
	CreateUser(c *gin.Context, newUser models.CreateUserInput) (models.UserResponse, error)
	GetUserByEmailWithPassword(c *gin.Context, email string) (models.UserWithPassword, error)
	GetUserByIdWithPassword(c *gin.Context, userId int64) (models.UserWithPassword, error)
	GetUserById(c *gin.Context, userId int64) (models.UserResponse, error)
	DeleteUser(c *gin.Context, userId int64) error
	UpdateUser(c *gin.Context, userId int64, updatedUser models.UpdateUserInput) (models.UserResponse, error)
	UpdateUserPassword(c *gin.Context, userId int64, password string) (models.UserResponse, error)
}

type UserRepository struct {
	db        database.DatabaseManager
	schema    string
	tableName string
}

func NewUserRepository(db database.DatabaseManager, cfg *config.Config) UserRepositoryInterface {
	return &UserRepository{
		db:        db,
		schema:    cfg.DBSchema,
		tableName: "user",
	}
}

/*
CreateUser creates a new user in the user table
newUser: User object with the details of the new user
returns: User object of the created user
*/
func (u *UserRepository) CreateUser(c *gin.Context, newUser models.CreateUserInput) (models.UserResponse, error) {
	var user models.UserResponse
	query, values, ptrs, err := helper.CreateInsertQuery(&newUser, &user, u.tableName, u.schema)
	if err != nil {
		return models.UserResponse{}, err
	}
	err = u.db.FetchOne(c, query, values...).Scan(ptrs...)
	if err != nil {
		return models.UserResponse{}, err
	}
	return user, nil
}

/*
GetUserByEmailWithPassword returns a user by email
email: Email of the user to be fetched
returns: User object of the fetched user with password
*/
func (u *UserRepository) GetUserByEmailWithPassword(c *gin.Context, email string) (models.UserWithPassword, error) {
	var user models.UserWithPassword
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&user)
	if err != nil {
		return models.UserWithPassword{}, err
	}
	query := fmt.Sprintf(`
		SELECT %s FROM %s.%s
		WHERE email = $1 AND deleted_at IS NULL;`,
		strings.Join(dbFields, ", "), u.schema, u.tableName)
	err = u.db.FetchOne(c, query, email).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserWithPassword{}, commonErrors.NewUserNotFoundError(err)
		}
		return models.UserWithPassword{}, err
	}
	return user, nil
}

/*
GetUserByIdWithPassword returns a user by Id
userId: Id of the user to be fetched
returns: User object of the fetched user with password
*/
func (u *UserRepository) GetUserByIdWithPassword(c *gin.Context, userId int64) (models.UserWithPassword, error) {
	var user models.UserWithPassword
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&user)
	if err != nil {
		return models.UserWithPassword{}, err
	}
	query := fmt.Sprintf(`
		SELECT %s FROM %s.%s WHERE id = $1 AND deleted_at IS NULL;`, strings.Join(dbFields, ", "), u.schema, u.tableName)
	err = u.db.FetchOne(c, query, userId).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserWithPassword{}, commonErrors.NewUserNotFoundError(err)
		}
		return models.UserWithPassword{}, err
	}
	return user, nil
}

/*
GetUserById returns a user by Id
userId: Id of the user to be fetched
returns: User object of the fetched user
*/
func (u *UserRepository) GetUserById(c *gin.Context, userId int64) (models.UserResponse, error) {
	var user models.UserResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&user)
	if err != nil {
		return models.UserResponse{}, err
	}
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s.%s
		WHERE id = $1 AND deleted_at IS NULL;`,
		strings.Join(dbFields, ", "), u.schema, u.tableName)
	err = u.db.FetchOne(c, query, userId).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserResponse{}, commonErrors.NewUserNotFoundError(err)
		}
		return models.UserResponse{}, err
	}
	return user, nil
}

/*
DeleteUser deletes a user by Id
userId: Id of the user to be deleted
returns: nil
*/
func (u *UserRepository) DeleteUser(c *gin.Context, userId int64) error {
	query := fmt.Sprintf(`
	UPDATE %s.%s
	SET deleted_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL;
	`, u.schema, u.tableName)

	rowsAffected, err := u.db.ExecuteQuery(c, query, userId)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return commonErrors.NewUserNotFoundError(errors.New("user not found or already deleted"))
	}

	return nil
}

/*
UpdateUser updates a user by Id
userId: Id of the user to be updated
updatedUser: User object with the updated details
returns: User object of the updated user
*/
func (u *UserRepository) UpdateUser(c *gin.Context, userId int64, updatedUser models.UpdateUserInput) (models.UserResponse, error) {
	fieldsClause, argValues, argIndex, err := helper.CreateUpdateParams(&updatedUser)
	if err != nil {
		return models.UserResponse{}, err
	}
	var user models.UserResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&user)
	if err != nil {
		return models.UserResponse{}, err
	}

	query := fmt.Sprintf(`
		UPDATE %s.%s SET %s
		WHERE id = $%d AND deleted_at IS NULL %s;`,
		u.schema, u.tableName, fieldsClause, argIndex, "RETURNING "+strings.Join(dbFields, ", "))

	err = u.db.FetchOne(c, query, append(argValues, userId)...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserResponse{}, commonErrors.NewUserNotFoundError(err)
		}
		return models.UserResponse{}, err
	}
	return user, nil
}

/*
UpdateUserPassword updates a user's password by Id
userId: Id of the user to be updated
password: Password of the user to be updated
returns: User object of the updated user
*/
func (u *UserRepository) UpdateUserPassword(c *gin.Context, userId int64, password string) (models.UserResponse, error) {
	var user models.UserResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&user)
	if err != nil {
		return models.UserResponse{}, err
	}
	query := fmt.Sprintf(`
		UPDATE %s.%s SET password = $1 WHERE id = $2 AND deleted_at IS NULL RETURNING %s;`, u.schema, u.tableName, strings.Join(dbFields, ", "))
	err = u.db.FetchOne(c, query, password, userId).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserResponse{}, commonErrors.NewUserNotFoundError(err)
		}
		return models.UserResponse{}, err
	}
	return user, nil
}
