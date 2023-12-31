package controllers

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/mhdianrush/go-login-signup-auth/config"
	"github.com/mhdianrush/go-login-signup-auth/entities"
	"github.com/mhdianrush/go-login-signup-auth/libraries"
	"github.com/mhdianrush/go-login-signup-auth/models"
	"golang.org/x/crypto/bcrypt"
)

type USerInput struct {
	Username string `validate:"required"`
	Password string `validate:"required"`
}

func Index(w http.ResponseWriter, r *http.Request) {
	// checked session login of each client, if empty, will redirect to login page
	session, err := config.Store.Get(r, config.SESSION_ID)
	if err != nil {
		panic(err)
	}

	if len(session.Values) == 0 {
		// empty session
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		if session.Values["LoggedIn"] != true {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		} else {
			data := map[string]any{
				"full_name_of_user": session.Values["full_name"],
			}
			temp, err := template.ParseFiles("views/index.html")
			if err != nil {
				panic(err)
			}
			temp.Execute(w, data)
		}
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		temp, err := template.ParseFiles("views/login.html")
		if err != nil {
			panic(err)
		}
		temp.Execute(w, nil)
	} else if r.Method == http.MethodPost {
		// login process
		r.ParseForm()

		userInput := &USerInput{
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
		}

		// validation
		var errorMessages = libraries.NewValidation().Struct(userInput)

		if errorMessages != nil {
			data := map[string]any{
				"validation": errorMessages,
			}

			temp, err := template.ParseFiles("views/login.html")
			if err != nil {
				panic(err)
			}
			temp.Execute(w, data)
		} else {
			var user entities.User
			err := models.NewUserModel().Find(&user, "username", userInput.Username)
			if err != nil {
				panic(err)
			}

			var message error
			// if login failed

			if user.Username == "" {
				// nothing match data with database
				message = errors.New("username or password doesn't match")
			} else {
				// exist username data in database
				// compare password userInput with password in database with hashing
				errPassword := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userInput.Password))
				if errPassword != nil {
					// password not match
					message = errors.New("username or password doesn't match")
				}
			}
			// if failed login
			if message != nil {
				data := map[string]any{
					"error": message,
				}
				temp, err := template.ParseFiles("views/login.html")
				if err != nil {
					panic(err)
				}
				temp.Execute(w, data)
			} else {
				// if username and password correct, set session
				session, _ := config.Store.Get(r, config.SESSION_ID)
				session.Values["LoggedIn"] = true
				session.Values["email"] = user.Email
				session.Values["username"] = user.Username
				session.Values["full_name"] = user.FullName

				session.Save(r, w)

				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
		}
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := config.Store.Get(r, config.SESSION_ID)
	if err != nil {
		panic(err)
	}
	// delete the session
	session.Options.MaxAge = -1
	// -1 is the default number to delete session

	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		temp, err := template.ParseFiles("views/register.html")
		if err != nil {
			panic(err)
		}
		temp.Execute(w, nil)
	}

	if r.Method == http.MethodPost {
		// manual validation
		// helper.RegistrationManualValidation(w, r)

		user := entities.User{
			FullName:        r.FormValue("full_name"),
			Email:           r.FormValue("email"),
			Username:        r.FormValue("username"),
			Password:        r.FormValue("password"),
			ConfirmPassword: r.FormValue("confirm_password"),
		}

		// validation with validator package
		var errorMessages = libraries.NewValidation().Struct(user)

		if errorMessages != nil {
			data := map[string]any{
				"validation": errorMessages,
				// so that the filled text not lost if all of the input text doesn't exist
				"user": user,
			}

			temp, err := template.ParseFiles("views/register.html")
			if err != nil {
				panic(err)
			}
			temp.Execute(w, data)
		} else {
			// hash password
			hashPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
			user.Password = string(hashPassword)

			// insert to db
			_, err := models.NewUserModel().Create(user)
			if err != nil {
				panic(err)
			}

			data := map[string]any{
				"message": "Registration Successfully",
			}
			temp, _ := template.ParseFiles("views/register.html")
			temp.Execute(w, data)
		}
	}
}
