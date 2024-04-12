package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/user"
)

func isAlreadyLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	userInfo := user.GetUserInfo(r)
	if userInfo.Permission != user.View {
		log.Printf("user %s (%s) is already logged in", userInfo.Name, r.RemoteAddr)
		template.ErrorTemplate(w, r, "You're already logged in", http.StatusBadRequest)
		return true
	}
	return false
}

type githubAccessInfo struct {
	AccessToken           string `json:"access_token"`
	ExpiresIn             int `json:"expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int `json:"refresh_token_expires_in"`
	TokenType             string `json:"token_type"`
}

func getGitHubApi(path string, accessInfo githubAccessInfo) (responseBody []byte, responseCode int, err error) {
	log.Printf("Accessing GitHub API: %s", path)
	var request *http.Request
	request, err = http.NewRequest("GET", fmt.Sprintf("https://api.github.com/%s", path), nil)
	if err != nil {
		panic(err)
	}

	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", fmt.Sprintf("%s %s", accessInfo.TokenType, accessInfo.AccessToken))
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	var response *http.Response
	response, err = http.DefaultClient.Do(request)
	responseCode = response.StatusCode
	if err != nil {
		return
	}

	responseBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	return
}

func getGitHubUsername(accessInfo githubAccessInfo) (string, error) {
	data, statusCode, err := getGitHubApi("user", accessInfo)
	if err != nil {
		return "", err
	}
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("Request error: %s", string(data))
	}

	var userInfo struct {
		Login string `json:"login"`
	}

	err = json.Unmarshal(data, &userInfo)
	if err != nil {
		return "", fmt.Errorf("Failed to unmarshal GitHub user info JSON: %v", err)
	}

	if userInfo.Login != "" {
		return userInfo.Login, nil
	}

	return "", errors.New("Failed to locate GitHub username")
}

func hasAdminGitHubTeam(username string, orgName string, teamSlug string, accessInfo githubAccessInfo) (bool, error) {
	data, statusCode, err := getGitHubApi(fmt.Sprintf("orgs/%s/teams/%s/memberships/%s", orgName, teamSlug, username), accessInfo)
	if err != nil {
		return false, err
	}
	if statusCode != http.StatusOK {
		return false, nil
	}

	var responseData struct {
		State string `json:"state"`
		Role  string `json:"role"`
		Url   string `json:"url"`
	}

	err = json.Unmarshal(data, &responseData)
	if err != nil {
		return false, fmt.Errorf("Failed to unmarshal GitHub teams info JSON: %v", err)
	}

	return responseData.State == "active", nil
}

func completeAuthentication(w http.ResponseWriter, r *http.Request, userInfo user.UserInfo, redirectUrl string) error {
	cookie, err := user.CreateCookie(userInfo)
	if err != nil {
		log.Printf("error logging in %s: %v", r.RemoteAddr, err)
		template.ErrorTemplate(w, r, "", http.StatusInternalServerError)
		return err
	}
	http.SetCookie(w, cookie)
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, redirectUrl, http.StatusMovedPermanently)

	return nil
}

func LoginGetGithubCallback(githubAppConfig *app.GitHubAuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		if codeParam := params.Get("code"); codeParam != "" {
			//TODO: while we technically ought to put a CSRF token in the state here, since
			//we only currently use it for redirecting, we don't actually need it
			//the redirect target should be CSRF-protected anyway
			redirectUrl := params.Get("state")

			exchangeParams := url.Values{}
			exchangeParams.Add("code", codeParam)
			exchangeParams.Add("client_id",  githubAppConfig.ClientId)
			exchangeParams.Add("client_secret", githubAppConfig.ClientSecret)

			githubAccessToken, err := url.Parse("https://github.com/login/oauth/access_token")
			if err != nil {
				panic(err)
			}

			githubAccessToken.RawQuery = exchangeParams.Encode()
			var emptyBody []byte

			request, err := http.NewRequest("POST", githubAccessToken.String(), bytes.NewBuffer(emptyBody))
			if err != nil {
				log.Println("Request creation error: %v", err)
				return
			}

			request.Header.Set("Accept", "application/json")

			log.Println("Contacting GitHub for API token")
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Printf("Error requesting user access token from GitHub: %v", err)
				template.ErrorTemplate(w, r, "Unable to log you in with GitHub", http.StatusInternalServerError)
				return
			}
			log.Println("Processing response")

			defer response.Body.Close()
			bodyStr, _ := ioutil.ReadAll(response.Body)

			var accessInfo githubAccessInfo
			err = json.Unmarshal(bodyStr, &accessInfo)
			if err != nil {
				log.Printf("Error parsing user access token response JSON from GitHub: %v", err)
				template.ErrorTemplate(w, r, "Error authenticating you with GitHub", http.StatusInternalServerError)
			}

			username, err := getGitHubUsername(accessInfo)
			if err != nil {
				log.Printf("Error getting GitHub username: %v", err)
				template.ErrorTemplate(w, r, "Error getting GitHub username", http.StatusInternalServerError)
				return
			}

			isAdmin, err := hasAdminGitHubTeam(username, githubAppConfig.OrgName, githubAppConfig.TeamSlug, accessInfo)
			if isAdmin {
				log.Printf("GitHub user %s is an administrator :)", username)
				completeAuthentication(w, r, user.UserInfo{
					Name: username,
					Permission: user.Admin,
				}, redirectUrl)
			} else {
				log.Printf("GitHub user %s is NOT an administrator", username)
				template.ErrorTemplate(w, r, "Can't log you in because you're not an administrator", http.StatusUnauthorized)
				return
			}
		} else {
			log.Println("Weird GitHub callback query: %s", r.URL.String())
		}
	}
}

func LoginGetUserPassword(w http.ResponseWriter, r *http.Request) {
	if(!isAlreadyLoggedIn(w, r)){
		template.ExecuteTemplate(w, r, "login")
	}
}

func LoginPostUserPassword(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Println("bad login post from %s: %v", r.RemoteAddr, err)
			template.ErrorTemplate(w, r, "", http.StatusBadRequest)
			return
		}
		if isAlreadyLoggedIn(w, r) {
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		redirectUrl := r.FormValue("redirect_url")

		userInfo, err := db.AuthenticateUser(username, []byte(password))
		//TODO: check the type of error (unknown user, wrong password, etc)
		if err != nil {
			log.Println("%v", err)
			template.ErrorTemplate(w, r, "Failed to login", http.StatusUnauthorized)
			return
		}

		completeAuthentication(w, r, userInfo, redirectUrl)
	}
}

func LogoutGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("logging out user on %s", r.RemoteAddr)
	http.SetCookie(w, user.DeleteCookie())
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, r.Referer(), http.StatusMovedPermanently)
}
