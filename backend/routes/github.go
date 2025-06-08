package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"oneclickdevenv/backend/db"
	"oneclickdevenv/backend/models"
	"oneclickdevenv/backend/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// Helper to always get fresh config with up-to-date env vars
func getGithubOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
		RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
	}
}

func GitHubLogin(c *gin.Context) {
	fmt.Println("---- GitHubLogin called ----")
	fmt.Println("GITHUB_CLIENT_ID:", os.Getenv("GITHUB_CLIENT_ID"))
	fmt.Println("GITHUB_CLIENT_SECRET:", os.Getenv("GITHUB_CLIENT_SECRET"))
	fmt.Println("GITHUB_REDIRECT_URL:", os.Getenv("GITHUB_REDIRECT_URL"))

	oauthConfig := getGithubOauthConfig()
	url := oauthConfig.AuthCodeURL("random-state", oauth2.AccessTypeOnline)
	fmt.Println("Redirecting to GitHub OAuth URL:", url)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GitHubCallback(c *gin.Context) {
	fmt.Println("---- GitHubCallback called ----")
	code := c.Query("code")
	fmt.Println("Received code:", code)
	if code == "" {
		fmt.Println("Error: Missing code in callback")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
		return
	}

	oauthConfig := getGithubOauthConfig()
	fmt.Println("Exchanging code for token...")
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("Token exchange failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token exchange failed", "details": err.Error()})
		return
	}
	fmt.Println("Token exchange successful.")

	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		fmt.Printf("Failed to fetch user info: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		fmt.Printf("Failed to decode user info: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}
	fmt.Printf("GitHub user info: %+v\n", userInfo)

	var user models.User
	if err := db.DB.FirstOrCreate(&user, models.User{
		GitHubID:  strconv.Itoa(userInfo.ID),
		Name:      userInfo.Login,
		Email:     userInfo.Email,
		AvatarURL: userInfo.AvatarURL,
	}).Error; err != nil {
		fmt.Printf("Database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	fmt.Printf("User record in DB: %+v\n", user)

	jwtToken, err := services.GenerateJWT(user.ID.String(), user.GitHubID, user.Email)
	if err != nil {
		fmt.Printf("Failed to generate JWT: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
		return
	}
	fmt.Println("JWT generated successfully.")

	c.JSON(http.StatusOK, gin.H{
		"message":    "GitHub auth successful",
		"token":      jwtToken,
		"login":      user.Name,
		"email":      user.Email,
		"name":       userInfo.Name,
		"avatar_url": user.AvatarURL,
	})
}
