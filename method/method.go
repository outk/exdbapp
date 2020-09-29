package method

import (
	"fmt"
	"log"
	"strings"
	"time"

	user_db "github.com/outk/exdbapp/userdb"

	"github.com/go-sql-driver/mysql"
)

const (
	queryInsertUser         = "INSERT INTO users(first_name, last_name, email, status) VALUES(?, ?, ?, ?);"
	queryGetUser            = "SELECT id, first_name, last_name, email, data_created, status FROM users WHERE id=?;"
	queryUpdateUser         = "UPDATE users SET first_name=?, last_name=?, email=?, status=? WHERE id=?;"
	queryDeleteUser         = "DELETE FROM users WHERE id=?;"
	querySelectUserByStatus = "SELECT id, first_name, last_name, email, data_created, status FROM users WHERE status=?;"
	errorNoRow              = "no rows in result set"
)

//User info
type User struct {
	ID        int64
	FirstName string
	LastName  string
	Email     string
	Date      *time.Time
	Status    int
}

var (
//UserDB = make(map[int64]*User)
)

func (user *User) Get() error {
	stmt, err := user_db.Client.Prepare(queryGetUser)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result := stmt.QueryRow(user.ID)
	if getErr := result.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Date, &user.Status); getErr != nil {

		return parseError(getErr)
	}
	return nil
}

func (user *User) Save() error {
	stmt, err := user_db.Client.Prepare(queryInsertUser)
	if err != nil {
		return err
	}
	defer stmt.Close()

	insertResult, saveErr := stmt.Exec(user.FirstName, user.LastName, user.Email, user.Status)
	if saveErr != nil {
		return parseError(saveErr)
	}

	userID, err := insertResult.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = userID

	log.Println("Save user data")

	return nil
}

func (user *User) Update() error {
	stmt, err := user_db.Client.Prepare(queryUpdateUser)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, updErr := stmt.Exec(user.FirstName, user.LastName, user.Email, user.Status, user.ID); updErr != nil {
		return parseError(updErr)
	}
	return nil
}

func (user *User) Delete() error {
	stmt, err := user_db.Client.Prepare(queryDeleteUser)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, delErr := stmt.Exec(user.ID); delErr != nil {
		return parseError(delErr)
	}

	return nil
}

func (user *User) FindByStatus(status int) ([]User, error) {
	stmt, err := user_db.Client.Prepare(querySelectUserByStatus)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(status)
	if err != nil {
		return nil, parseError(err)
	}
	defer rows.Close()

	res := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Date, &user.Status); err != nil {
			return nil, parseError(err)
		}
		res = append(res, user)
	}

	return res, nil
}

func parseError(err error) error {
	sqlErr, ok := err.(*mysql.MySQLError)
	if !ok {
		if strings.Contains(err.Error(), errorNoRow) {
			return fmt.Errorf("user not found")
		}
		return err
	}
	//以下,sql依存のエラー
	switch sqlErr.Number {
	case 1062:
		//email_UNIQUE error
		return fmt.Errorf("email already exist")
	}
	return err
}
