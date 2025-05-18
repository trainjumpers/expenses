package repository

import (
	"expenses/internal/database/helper"
	"expenses/internal/models"
	"expenses/pkg/logger"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db     *pgxpool.Pool
	schema string
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db:     db,
		schema: helper.GetPGSchema(),
	}
}

/*
CreateUser creates a new user in the user table
newUser: User object with the details of the new user
returns: User object of the created user
*/
func (u *UserRepository) CreateUser(c *gin.Context, newUser models.CreateUserInput) (models.UserResponse, error) {
	var user models.UserResponse
	query, values, ptrs, err := helper.CreateInsertQuery(&newUser, &user, "user")
	if err != nil {
		return models.UserResponse{}, err
	}
	logger.Info("Executing query to create a user: ", query)
	err = u.db.QueryRow(c, query, values...).Scan(ptrs...)
	if err != nil {
		return models.UserResponse{}, err
	}
	return user, nil
}

/*
GetUserByEmail returns a user by email
email: Email of the user to be fetched
returns: User object of the fetched user
*/
func (u *UserRepository) GetUserByEmail(c *gin.Context, email string) (models.UserWithPassword, error) {
	var user models.UserWithPassword
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&user)
	if err != nil {
		return models.UserWithPassword{}, err
	}
	query := fmt.Sprintf(`
		SELECT %s FROM %s.user 
		WHERE email = $1 AND deleted_at IS NULL;`,
		strings.Join(dbFields, ", "), u.schema)
	logger.Info("Executing query to get a user by email: ", query)
	err = u.db.QueryRow(c, query, email).Scan(ptrs...)
	if err != nil {
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
		FROM %s.user 
		WHERE id = $1 AND deleted_at IS NULL;`,
		strings.Join(dbFields, ", "), u.schema)
	logger.Info("Executing query to get a user by Id: ", query)
	err = u.db.QueryRow(c, query, userId).Scan(ptrs...)
	if err != nil {
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
	UPDATE %s.user 
	SET deleted_at = NOW() 
	WHERE id = $1 AND deleted_at IS NULL;
	`, u.schema)
	logger.Info("Executing query to delete a user by Id: ", query)
	_, err := u.db.Exec(c, query, userId)
	if err != nil {
		return err
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
		UPDATE %[1]s.user SET %[2]s 
		WHERE id = $%d AND deleted_at IS NULL %s;`,
		u.schema, fieldsClause, argIndex, "RETURNING "+strings.Join(dbFields, ", "))

	logger.Info("Executing query to update a user by Id: ", query)
	err = u.db.QueryRow(c, query, append(argValues, userId)...).Scan(ptrs...)
	if err != nil {
		return models.UserResponse{}, err
	}
	return user, nil
}
