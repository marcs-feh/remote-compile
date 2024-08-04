package main

import (
	"time"

	"github.com/labstack/echo/v4"
)

func setupRoutes(e *echo.Echo){
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Key SessionKey `json:"session_key"`
	TimeToLive time.Duration `json:"ttl"`
}

