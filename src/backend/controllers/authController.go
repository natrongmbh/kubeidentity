package controllers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/natrongmbh/kubeperm/util"
)

type GithubUser struct {
	ID              float64  `json:"id"`
	Login           string   `json:"login"`
	Email           string   `json:"email"`
	Name            string   `json:"name"`
	AvatarURL       string   `json:"avatar_url"`
	GithubTeamSlugs []string `json:"github_team_slugs"`
	Organization    string   `json:"organization"`
}

// GetGithubTeams returns the redirect url
func CheckGithubLogin(c *fiber.Ctx) error {

	util.InfoLogger.Printf("%s %s %s", c.IP(), c.Method(), c.Path())

	githubUser := CheckAuth(c)
	if githubUser.ID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	return c.JSON(githubUser)
}

// FrontendGithubLogin gets the github data and sends it to LoggedIn()
func FrontendGithubLogin(c *fiber.Ctx) error {

	util.InfoLogger.Printf("%s %s %s", c.IP(), c.Method(), c.Path())

	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body (body parser)",
		})
	}

	// get access_token from data
	if githubCode := data["github_code"]; githubCode == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body (github_code)",
		})
	} else {
		return LoggedIn(c, githubCode)
	}

}

// GithubCallback handles the callback with the code query param
func GithubCallback(c *fiber.Ctx) error {

	util.InfoLogger.Printf("%s %s %s", c.IP(), c.Method(), c.Path())

	// get code from "code" query param
	if githubCode := c.Query("code"); githubCode == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	} else {
		return LoggedIn(c, githubCode)
	}
}

// LoggedIn handles the login and returns the token
func LoggedIn(c *fiber.Ctx, githubCode string) error {

	githubAccessToken := util.GetGithubAccessToken(githubCode)
	githubTeamsData := util.GetGithubTeams(githubAccessToken)
	githubUserData := util.GetGithubUser(githubAccessToken)

	if githubTeamsData == "" || githubUserData == "" {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to get github data",
		})
	}

	var githubTeamsDataMap []map[string]interface{}
	if err := json.Unmarshal([]byte(githubTeamsData), &githubTeamsDataMap); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message":       "Internal server error",
			"error":         err.Error(),
			"github_mesage": githubTeamsData,
		})
	}

	var githubUserDataMap map[string]interface{}
	if err := json.Unmarshal([]byte(githubUserData), &githubUserDataMap); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": "Internal server error",
			"error":   err.Error(),
		})
	}

	var githubTeamSlugs []string

	for _, githubTeam := range githubTeamsDataMap {
		githubTeamSlugs = append(githubTeamSlugs, githubTeam["slug"].(string))
	}

	githubUser := GithubUser{
		ID:              githubUserDataMap["id"].(float64),
		Login:           githubUserDataMap["login"].(string),
		Email:           githubUserDataMap["email"].(string),
		Name:            githubUserDataMap["name"].(string),
		AvatarURL:       githubUserDataMap["avatar_url"].(string),
		GithubTeamSlugs: githubTeamSlugs,
		Organization:    "",
	}

	exp := time.Now().Add(time.Hour * 24).Unix()

	claims := jwt.MapClaims{
		"github_team_slugs": githubTeamSlugs,
		"id":                githubUser.ID,
		"login":             githubUser.Login,
		"email":             githubUser.Email,
		"name":              githubUser.Name,
		"avatar_url":        githubUser.AvatarURL,
		"organization":      util.GITHUB_ORGANIZATION,
		"exp":               exp,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(util.JWT_SECRET_KEY))

	return c.JSON(fiber.Map{
		"githubUser": githubUser,
		"token":      tokenString,
	})
}

// CheckAuth checks if the token is valid and returns the github team slugs
func CheckAuth(c *fiber.Ctx) GithubUser {
	var token *jwt.Token
	var tokenString string

	// get bearer token from header
	bearerToken := c.Get("Authorization")

	// split bearer token to get token
	bearerTokenSplit := strings.Split(bearerToken, " ")
	if len(bearerTokenSplit) == 2 {
		tokenString = bearerTokenSplit[1]
	} else {
		return GithubUser{}
	}

	if tokenString == "" {
		// return unauthorized
		return GithubUser{}
	}

	var err error
	// parse token with secret key
	token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(util.JWT_SECRET_KEY), nil
	})

	if err != nil {
		return GithubUser{}
	}

	if token == nil {
		return GithubUser{}
	}

	// validate expiration
	if !token.Valid {
		return GithubUser{}
	}

	// validate claims
	claims := token.Claims.(jwt.MapClaims)

	if claims["exp"] == nil {
		return GithubUser{}
	} else {
		exp := claims["exp"]
		// convert exp to int64
		expInt64 := int64(exp.(float64))
		if expInt64 < time.Now().Unix() {
			return GithubUser{}
		}
	}

	if claims["github_team_slugs"] == nil {
		util.WarningLogger.Printf("IP %s is not authorized", c.IP())
		return GithubUser{}
	}

	var githubTeamSlugs []string
	for _, githubTeam := range claims["github_team_slugs"].([]interface{}) {
		githubTeamSlugs = append(githubTeamSlugs, githubTeam.(string))
	}

	// if githubTeamSlugs == nil {
	// 	util.WarningLogger.Printf("IP %s is not authorized", c.IP())
	// 	return nil
	// }

	// return claims map as json

	return GithubUser{
		ID:              claims["id"].(float64),
		GithubTeamSlugs: githubTeamSlugs,
		Login:           claims["login"].(string),
		Email:           claims["email"].(string),
		Name:            claims["name"].(string),
		AvatarURL:       claims["avatar_url"].(string),
		Organization:    claims["organization"].(string),
	}
}