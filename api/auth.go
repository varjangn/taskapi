package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/varjangn/taskapi/types"
	"golang.org/x/crypto/bcrypt"
)

type regReqType struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
}

type loginReqType struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRespType struct {
	Token string     `json:"access"`
	User  types.User `json:"user"`
}

func validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password can not be empty")
	}
	if len(password) < 8 {
		return fmt.Errorf("length of password must be greater then 8 charcters")
	}
	done, err := regexp.MatchString("([a-z])+", password)
	if err != nil {
		log.Println("ValidationError:", err.Error())
		return fmt.Errorf("invalid password")
	}
	if !done {
		return fmt.Errorf("password should contain atleast one lower case character")
	}
	done, err = regexp.MatchString("([A-Z])+", password)
	if err != nil {
		log.Println("ValidationError:", err.Error())
		return fmt.Errorf("invalid password")
	}
	if !done {
		return fmt.Errorf("password should contain atleast one upper case character")
	}
	done, err = regexp.MatchString("([0-9])+", password)
	if err != nil {
		log.Println("ValidationError:", err.Error())
		return fmt.Errorf("invalid password")
	}
	if !done {
		return fmt.Errorf("password should contain atleast one digit")
	}

	done, err = regexp.MatchString("([!@#$%^&*.?-])+", password)
	if err != nil {
		log.Println("ValidationError:", err.Error())
		return fmt.Errorf("invalid password")
	}
	if !done {
		return fmt.Errorf("password should contain atleast one special character")
	}
	return nil
}

func (s *APIServer) RegisterEndpoint(w http.ResponseWriter, r *http.Request) {
	regReq := new(regReqType)
	if err := json.NewDecoder(r.Body).Decode(regReq); err != nil {
		log.Println("RegisterError:", err.Error())
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "invalid Request body"})
		return
	}
	if err := validatePassword(regReq.Password); err != nil {
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": err.Error()})
		return
	}
	if regReq.ConfirmPassword != regReq.Password {
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "password does not match"})
		return
	}

	addr, err := mail.ParseAddress(regReq.Email)
	if err != nil {
		log.Println("RegisterError:", err.Error())
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "invalid Email"})
		return
	}
	regReq.Email = addr.Address

	userId, err := s.store.GetUserId(regReq.Email)
	if err != nil {
		log.Println("RegisterError:", err.Error())
		WriteJSON(w, http.StatusInternalServerError,
			Map{"error": "unknown error"})
		return
	}

	if userId > 0 {
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "user with this email already exists"})
		return
	}

	user, err := types.NewUser(regReq.Email,
		regReq.Password, regReq.FirstName, regReq.LastName,
	)
	if err != nil {
		log.Println("RegisterError:", err.Error())
		WriteJSON(w, http.StatusInternalServerError,
			Map{"error": "unknown error"})
	}

	err = s.store.CreateUser(user)
	if err != nil {
		log.Println("RegisterError:", err.Error())
		WriteJSON(w, http.StatusInternalServerError,
			Map{"error": "unknown error"})
		return
	}

	WriteJSON(w, http.StatusCreated, user)
}

func (s *APIServer) LoginEndpoint(w http.ResponseWriter, r *http.Request) {
	loginReq := new(loginReqType)
	if err := json.NewDecoder(r.Body).Decode(loginReq); err != nil {
		log.Println("LoginError:", err.Error())
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "invalid request body"})
		return
	}

	if err := validatePassword(loginReq.Password); err != nil {
		WriteJSON(w, http.StatusBadRequest, Map{"error": "invalid password"})
		return
	}

	user, err := s.store.GetUser(loginReq.Email)
	if err != nil {
		log.Println("LoginError:", err.Error())
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "invalid request body"})
		return
	}

	if user == nil {
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "invalid email address"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
	if err != nil {
		WriteJSON(w, http.StatusBadRequest,
			Map{"error": "invalid password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Email,
		"exp": time.Now().Add(120 * time.Minute).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	resp := loginRespType{
		Token: tokenStr,
		User:  *user,
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (s *APIServer) ProfileEndpoint(w http.ResponseWriter, r *http.Request) {
	reqUser := r.Context().Value(authUserKey).(*types.User)
	WriteJSON(w, http.StatusOK, reqUser)
}
