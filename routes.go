package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const DefaultTTL = 10 * time.Second

func setupRoutes(e *echo.Echo, sessionStore *SessionStore, db *sql.DB){
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

	e.POST("/compile/:language", func(c echo.Context) error {
		lang := c.Param("language")
		req := c.Request()
		compileReq := compileCodeRequest{}

		blob, _ := io.ReadAll(req.Body) // TODO: Limit size
		err := json.Unmarshal(blob, &compileReq)
		if err != nil { return c.String(http.StatusBadRequest, "Bad Request") }

		authorized := sessionStore.ValidateSession(compileReq.Auth.Username, compileReq.Auth.Key)

		if !authorized {
			return c.String(http.StatusUnauthorized, "Failed to verify session")
		}

		var builder LanguageBuilder

		switch lang {
		case "odin":
			builder = OdinBuilder{}

		default:
			return c.String(http.StatusNotFound, "Unavailable language")
		}

		status := buildSource("main-odin",
			`package main

			main :: proc(){
			}
		`, builder)

		compileRes := compileCodeResponse {
			Success: status.Success,
			Stdout: strings.TrimRight(string(status.Stdout), "\u0000"),
			Stderr: strings.TrimRight(string(status.Stderr), "\u0000"),
			ElapsedMs: status.Elapsed.Milliseconds(),
		}

		jsonBlob, _ := json.Marshal(compileRes)

		return c.JSONBlob(200, jsonBlob)
	})
}


type authRequestPart struct {
	Username string `json:"username"`
	Key SessionKey `json:"session_key"`
}

type compileCodeRequest struct {
	Auth authRequestPart `json:"auth"`
	Source string `json:"source_code"`
}

type compileCodeResponse struct {
	Success bool `json:"success"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	ElapsedMs int64 `json:"elapsed_time"`
}

type runCodeRequest struct {
	Auth authRequestPart `json:"auth"`
	Args []string `json:"args"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Key SessionKey `json:"session_key"`
	TimeToLive time.Duration `json:"ttl"`
}

