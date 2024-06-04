package services

import (
	"expenses/entities"
	"expenses/models"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	logger "github.com/sirupsen/logrus"
)

type UserService struct {
	db     *pgxpool.Pool
	schema string
}

func NewUserService(db *pgxpool.Pool) *UserService {
	return &UserService{
		db:     db,
		schema: os.Getenv("PGSCHEMA"),
	}
}

/*
CreateUser creates a new user in the user table

newUser: User object with the details of the new user

returns: User object of the created user
*/
func (u *UserService) CreateUser(c *gin.Context, newUser models.User) models.User {
	fmt.Println(u.schema)
	query := fmt.Sprintf("INSERT INTO %s.user (first_name, last_name, email, dob, phone, password) VALUES ($1, $2, $3, $4, $5, $6) "+
		"RETURNING id, first_name, last_name, email, dob, phone;", u.schema)
	insert := u.db.QueryRow(c, query, newUser.FirstName, newUser.LastName, newUser.Email, newUser.DOB, newUser.Phone, newUser.Password)
	var createdUser models.User

	err := insert.Scan(&createdUser.ID, &createdUser.FirstName, &createdUser.LastName, &createdUser.Email, &createdUser.DOB, &createdUser.Phone)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("User with email: %s already exists", newUser.Email)})
			c.Abort()
			return models.User{}
		}
	}

	return createdUser
}

/*
GetUserByEmail returns a user by email

email: Email of the user to be fetched

returns: User object of the fetched user
*/
func (u *UserService) GetUserByEmail(c *gin.Context, email string) models.User {
	var user models.User
	fmt.Println(u.schema)

	query := fmt.Sprintf("SELECT id, email, password FROM %s.user WHERE email = $1 AND deleted_at IS NULL;", u.schema)
	result := u.db.QueryRow(c, query, email)

	err := result.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		c.Abort()
		return models.User{}
	}
	return user
}

/*
GetUserByID returns a user by ID

userID: ID of the user to be fetched

returns: User object of the fetched user
*/
func (u *UserService) GetUserByID(c *gin.Context, userID int64) models.User {
	var user models.User

	query := fmt.Sprintf("SELECT * FROM %s.user WHERE id = $1 AND deleted_at IS NULL;", u.schema)

	result := u.db.QueryRow(c, query, userID)

	err := result.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.DOB, &user.Phone)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		c.Abort()
		return models.User{}
	}

	return user
}

/*
GetUsers returns all users

returns: List of users ([]models.User)
*/
func (u *UserService) GetUsers(c *gin.Context) []models.User {
	query := fmt.Sprintf("SELECT id, first_name, last_name, email, dob, phone FROM %s.user WHERE deleted_at IS NULL;", u.schema)
	var users []models.User

	logger.Info("Executing query to get all users: ", query)
	result, err := u.db.Query(c, query)
	if err != nil {
		logger.Fatal(fmt.Errorf("error querying the database: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying users"})
		c.Abort()
		return nil
	}

	for result.Next() {
		var user models.User
		err := result.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.DOB, &user.Phone)
		if err != nil {
			logger.Fatal(fmt.Errorf("error scanning the database output: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing users"})
			c.Abort()
			return nil
		}
		users = append(users, user)
	}

	return users
}

/*
DeleteUser deletes a user by ID

userID: ID of the user to be deleted

returns: nil
*/
func (u *UserService) DeleteUser(c *gin.Context, userID int64) {
	query := fmt.Sprintf("UPDATE %s.user SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL;", u.schema)

	logger.Info("Executing query to delete a user by ID: ", query)
	_, err := u.db.Exec(c, query, userID)
	if err != nil {
		logger.Fatal(fmt.Errorf("error scanning the database output: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing users"})
		c.Abort()
		return
	}
}

/*
UpdateUser updates a user by ID

userID: ID of the user to be updated

updatedUser: User object with the updated details

returns: User object of the updated user
*/
func (u *UserService) UpdateUser(c *gin.Context, userID int64, updatedUser entities.UpdateUserInput) models.User {
	fmt.Println(updatedUser)

	fields := map[string]interface{}{
		"first_name": updatedUser.FirstName,
		"last_name":  updatedUser.LastName,
		"email":      updatedUser.Email,
		"dob":        updatedUser.DOB,
		"phone":      updatedUser.Phone,
	}

	if updatedUser.DOB.IsZero() {
		fields["dob"] = ""
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
		"RETURNING id, first_name, last_name, email, dob, phone;", u.schema, fieldsClause, argIndex)

	logger.Info("Executing query to update a user by ID: ", query)
	result := u.db.QueryRow(c, query, append(argValues, userID)...)

	var user models.User
	err := result.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.DOB, &user.Phone)
	if err != nil {
		logger.Fatal(fmt.Errorf("error scanning the database output: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing users"})
		c.Abort()
		return models.User{}
	}

	return user
}
