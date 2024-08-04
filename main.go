package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"unicode"

	_ "embed"

	"github.com/labstack/echo/v4"
	_ "modernc.org/sqlite"
)

func main(){
	e := echo.New()
	e.HideBanner = true
	setupRoutes(e)

	db, dbErr := ConnectToDB("data.db")
	ensureNil(dbErr)
	defer db.Close()

	initErr := InitializeDB(db)
	ensureNil(initErr)

	adminErr := createAdminUserIfNotExists(db)
	ensureNil(adminErr)

	sessionStore := NewSessionStore()

	const DefaultTTL = 10 * time.Second

	/*** Routes ***/
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello")
	})

	e.POST("/login", func(c echo.Context) error {
		req := c.Request()
		loginReq := loginRequest{}
		data, _ := io.ReadAll(req.Body) // TODO: Limit request size
		jsonErr := json.Unmarshal(data, &loginReq)
		if jsonErr != nil {
			return c.String(http.StatusBadRequest, "Bad request")
		}

		sessionKey, sessErr := sessionStore.BeginSession(db, loginReq.Username, loginReq.Password, DefaultTTL)
		if sessErr != nil {
			return c.String(http.StatusUnauthorized, "Failed to initialize session, invalid credentials")
		}
		loginRes := loginResponse{
			Key: sessionKey,
			TimeToLive: DefaultTTL,
		}

		resData, _ := json.Marshal(&loginRes)
		return c.JSONBlob(http.StatusOK, resData)
	})
	/**************/

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

func GetUserByName(db *sql.DB, name string) (user User, found bool, err error){
	var dbHash, dbSalt string
	row := db.QueryRow("select id, name, auth_hash, auth_salt from Users where name = $1", name)
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


