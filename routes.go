package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
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

}

type authRequestPart struct {
	Username string `json:"username"`
	Key SessionKey `json:"session_key"`
}

type compileCodeRequest struct {
	Auth authRequestPart `json:"auth"`
	Source string `json:"source_code"`
	Language string `json:"language"`
}

type compileCodeResponse struct {
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Key SessionKey `json:"session_key"`
	TimeToLive time.Duration `json:"ttl"`
}

