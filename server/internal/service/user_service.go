package service

import (
	"expenses/internal/models"
	"expenses/pkg/logger"
	"expenses/pkg/utils"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	db     *pgxpool.Pool
	schema string
}

func NewUserService(db *pgxpool.Pool) *UserService {
	return &UserService{
		db:     db,
		schema: utils.GetPGSchema(),
	}
}

/*
CreateUser creates a new user in the user table
newUser: User object with the details of the new user
returns: User object of the created user
*/
func (u *UserService) CreateUser(c *gin.Context, newUser models.CreateUserInput) (models.UserOutput, error) {
	var user models.UserOutput
	query, values, ptrs, err := utils.CreateInsertQuery(&newUser, &user, "user")
	if err != nil {
		return models.UserOutput{}, err
	}
	logger.Info("Executing query to create a user: ", query)
	err = u.db.QueryRow(c, query, values...).Scan(ptrs...)
	if err != nil {
		return models.UserOutput{}, err
	}
	return user, nil
}

/*
GetUserByEmail returns a user by email
email: Email of the user to be fetched
returns: User object of the fetched user
*/
func (u *UserService) GetUserByEmail(c *gin.Context, email string) (models.UserWithPassword, error) {
	var user models.UserWithPassword
	ptrs, dbFields, err := utils.GetDbFieldsFromObject(&user)
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
func (u *UserService) GetUserById(c *gin.Context, userId int64) (models.UserOutput, error) {
	var user models.UserOutput
	ptrs, dbFields, err := utils.GetDbFieldsFromObject(&user)
	if err != nil {
		return models.UserOutput{}, err
	}
	query := fmt.Sprintf(`
		SELECT %s 
		FROM %s.user 
		WHERE id = $1 AND deleted_at IS NULL;`,
		strings.Join(dbFields, ", "), u.schema)
	logger.Info("Executing query to get a user by Id: ", query)
	err = u.db.QueryRow(c, query, userId).Scan(ptrs...)
	if err != nil {
		return models.UserOutput{}, err
	}
	return user, nil
}

/*
DeleteUser deletes a user by Id
userId: Id of the user to be deleted
returns: nil
*/
func (u *UserService) DeleteUser(c *gin.Context, userId int64) error {
	query := fmt.Sprintf("UPDATE %s.user SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL;", u.schema)
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
func (u *UserService) UpdateUser(c *gin.Context, userId int64, updatedUser models.UpdateUserInput) (models.UserOutput, error) {
	fieldsClause, argValues, argIndex, err := utils.CreateUpdateParams(&updatedUser)
	if err != nil {
		return models.UserOutput{}, err
	}
	var user models.UserOutput
	ptrs, dbFields, err := utils.GetDbFieldsFromObject(&user)
	if err != nil {
		return models.UserOutput{}, err
	}

	query := fmt.Sprintf(`
		UPDATE %[1]s.user SET %[2]s 
		WHERE id = $%d AND deleted_at IS NULL %s;`,
		u.schema, fieldsClause, argIndex, "RETURNING "+strings.Join(dbFields, ", "))

	logger.Info("Executing query to update a user by Id: ", query)
	err = u.db.QueryRow(c, query, append(argValues, userId)...).Scan(ptrs...)
	if err != nil {
		return models.UserOutput{}, err
	}
	return user, nil
}

/*
updateUserPassword updates a user's password by Id
userId: Id of the user to be updated
PasswordDetails: User object with the updated password details
returns: User object of the updated user
*/
func (u *UserService) UpdateUserPassword(c *gin.Context, userId int64, passwordDetails models.UpdateUserPasswordInput) (models.UserOutput, error) {
	// Fetch old password
	var user models.UserOutput
	query := fmt.Sprintf("SELECT password FROM %[1]s.user WHERE id = $1 AND deleted_at IS NULL;", u.schema)
	logger.Info("Executing query to get a user by Id: ", query)
	var oldPassword string
	err := u.db.QueryRow(c, query, userId).Scan(&oldPassword)
	if err != nil {
		return models.UserOutput{}, err
	}
	if !utils.CheckPasswordHash(passwordDetails.OldPassword, oldPassword) {
		return models.UserOutput{}, fmt.Errorf("old password is incorrect")
	}
	hashedPassword, err := utils.HashPassword(passwordDetails.NewPassword)
	if err != nil {
		return models.UserOutput{}, err
	}
	ptrs, dbFields, err := utils.GetDbFieldsFromObject(&user)
	if err != nil {
		return models.UserOutput{}, err
	}
	returningClause := "RETURNING " + strings.Join(dbFields, ", ")
	query = fmt.Sprintf(`
		UPDATE %[1]s.user SET password = $2 
		WHERE id = $1 AND deleted_at IS NULL %s;`, u.schema, returningClause)
	logger.Info("Executing query to update a user by Id: ", query)
	err = u.db.QueryRow(c, query, userId, hashedPassword).Scan(ptrs...)
	if err != nil {
		return models.UserOutput{}, err
	}
	return user, nil
}
