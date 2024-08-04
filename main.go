package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"unicode"

	_ "embed"

	"github.com/labstack/echo/v4"
	_ "modernc.org/sqlite"
)

func main(){
	e := echo.New()
	setupRoutes(e)

	db, dbErr := ConnectToDB("data.db")
	ensureNil(dbErr)
	defer db.Close()

	initErr := InitializeDB(db)
	ensureNil(initErr)

	adminErr := createAdminUserIfNotExists(db)
	ensureNil(adminErr)


	log.Fatalf(e.Start(":8080").Error())
}

const ADMIN_PASS = "123"
func createAdminUserIfNotExists(db *sql.DB) (err error) {
	_, found, err := GetUserByID(db, 0)
	ensureNil(err)
	if !found {
		var user User

		user, err = NewUser("admin", ADMIN_PASS)
		if err != nil { return }

		_, err = db.Exec(`
			INSERT INTO Users(id, name, auth_hash, auth_salt, admin)
				VALUES(0, $1, $2, $3, 1);
		`, user.name, base64Encode(user.authHash), base64Encode(user.authSalt))
		if err != nil { return }
	}

	return
}

func GetUserByID(db *sql.DB, id Id) (user User, found bool, err error){
	var dbHash, dbSalt string
	row := db.QueryRow("select id, name, auth_hash, auth_salt from Users where id = $1", id)
	err = row.Scan(&user.id, &user.name, &dbHash, &dbSalt)

	if err == sql.ErrNoRows {
		return user, false, nil
	} else if err != nil {
		return user, false, err
	}

	user.authHash, _ = base64Decode(dbHash)
	user.authSalt, _ = base64Decode(dbSalt)

	return user, true, nil
}

//go:embed init_db.sql
var initDatabaseSQLSource string

func InitializeDB(db *sql.DB) error {
	_, err := db.Exec(initDatabaseSQLSource)
	return err
}

func ensureNil(e error){
	if e != nil {
		log.Fatal(e.Error())
	}
}

func ConnectToDB(path string) (db *sql.DB, err error){
	return sql.Open("sqlite", path)
}

type Id int64

type User struct {
	id Id
	name string
	authHash []byte
	authSalt []byte
}

func NewUser(username, password string) (user User, err error){
	user.authHash, user.authSalt, err = CreatePassword(password)
	if err != nil { return }
	if ok := validateUsername(username); !ok {
		err = fmt.Errorf("invalid username: %q", username)
		return
	}
	user.name = username
	return
}

func CreateUser(db *sql.DB, username, password string) (user User, err error) {
	user, err = NewUser(username, password)
	if err != nil { return }

	_, err = db.Exec(`
		INSERT INTO Users(name, auth_hash, auth_salt)
			VALUES($1, $2, $3);
	`, user.name, base64Encode(user.authHash), base64Encode(user.authSalt))

	return
}

func validateUsername(name string) bool {
	if len(name) < 2 { return false }

	for _, c := range name {
		ascii := c < 127
		ok := ascii && (unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' || c == '-')
		if !ok { return false }
	}

	return true
}

func base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func base64Decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}


