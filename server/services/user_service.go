package services

import (
	"expenses/entities"
	logger "expenses/logger"
	"expenses/models"
	"expenses/utils"
	"fmt"
	"strconv"
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
func (u *UserService) CreateUser(c *gin.Context, newUser models.User) (models.User, error) {
	fmt.Println(u.schema)
	query := fmt.Sprintf("INSERT INTO %s.user (name, email, password) VALUES ($1, $2, $3) "+
		"RETURNING id, name, email;", u.schema)
	insert := u.db.QueryRow(c, query, newUser.Name, newUser.Email, newUser.Password)
	logger.Info("Executing query to create a user: ", query)
	var createdUser models.User

	err := insert.Scan(&createdUser.ID, &createdUser.Name, &createdUser.Email)
	if err != nil {
		return models.User{}, err
	}

	return createdUser, nil
}

/*
GetUserByEmail returns a user by email

email: Email of the user to be fetched

returns: User object of the fetched user
*/
func (u *UserService) GetUserByEmail(c *gin.Context, email string) (models.User, error) {
	var user models.User
	fmt.Println(u.schema)

	query := fmt.Sprintf("SELECT id, email, password FROM %s.user WHERE email = $1 AND deleted_at IS NULL;", u.schema)
	result := u.db.QueryRow(c, query, email)

	err := result.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

/*
GetUserByID returns a user by ID

userID: ID of the user to be fetched

returns: User object of the fetched user
*/
func (u *UserService) GetUserByID(c *gin.Context, userID int64) (models.User, error) {
	var user models.User
	query := fmt.Sprintf("SELECT id, name, email FROM %s.user WHERE id = $1 AND deleted_at IS NULL;", u.schema)
	logger.Info("Executing query to get a user by ID: ", query)
	result := u.db.QueryRow(c, query, userID)

	err := result.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

/*
DeleteUser deletes a user by ID

userID: ID of the user to be deleted

returns: nil
*/
func (u *UserService) DeleteUser(c *gin.Context, userID int64) error {
	query := fmt.Sprintf("DELETE FROM %s.user WHERE id = $1 AND deleted_at IS NULL;", u.schema)

	logger.Info("Executing query to delete a user by ID: ", query)
	_, err := u.db.Exec(c, query, userID)
	if err != nil {
		return err
	}

	return nil
}

/*
UpdateUser updates a user by ID

userID: ID of the user to be updated

updatedUser: User object with the updated details

returns: User object of the updated user
*/
func (u *UserService) UpdateUser(c *gin.Context, userID int64, updatedUser entities.UpdateUserInput) (models.User, error) {
	fields := map[string]interface{}{
		"name": updatedUser.Name,
	}

	fieldsClause := ""
	argIndex := 1
	argValues := make([]interface{}, 0)
	for k, v := range fields {
		if v == "" {
			continue
		}

		fieldsClause += k + " = $" + strconv.FormatInt(int64(argIndex), 10) + ", "
		argIndex++
		argValues = append(argValues, v)
	}
	fieldsClause = strings.TrimSuffix(fieldsClause, ", ")

	query := fmt.Sprintf("UPDATE %[1]s.user SET %[2]s WHERE id = $%d AND deleted_at IS NULL "+
		"RETURNING id, name, email;", u.schema, fieldsClause, argIndex)

	logger.Info("Executing query to update a user by ID: ", query)
	result := u.db.QueryRow(c, query, append(argValues, userID)...)

	var user models.User
	err := result.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}
