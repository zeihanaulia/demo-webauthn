package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/png"
	"io/ioutil"
	"net/http"

	"github.com/duo-labs/webauthn.io/models"
	"github.com/duo-labs/webauthn.io/session"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var otpStore = make(map[string]string)

func (ws *Server) GetTOTP(w http.ResponseWriter, r *http.Request) {
	var resp TOTP
	session, _ := ws.store.Get(r, session.WebauthnSession)
	// Load the user from the database and store it in the request context
	if id, ok := session.Values["user_id"]; ok {
		u, err := models.GetUser(id.(uint))
		if err != nil {
			jsonResponse(w, "Cannot create TOTP", http.StatusBadRequest)
			return
		}

		resp, err = generate(u.Username)
		if err != nil {
			jsonResponse(w, "Cannot create TOTP", http.StatusBadRequest)
			return
		}

		otpStore[u.Username] = resp.Secret
	}

	jsonResponse(w, resp, http.StatusOK)
}

type VerifiyTOTPRequest struct {
	Username string `json:"username"`
	Passcode string `json:"passcode"`
}

func (ws *Server) VerifiyTOTP(w http.ResponseWriter, r *http.Request) {
	var p VerifiyTOTPRequest

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get secret
	var secret string
	if val, ok := otpStore[p.Username]; ok {
		secret = val
	}

	fmt.Println("TOTP", p.Passcode, secret)
	valid := totp.Validate(p.Passcode, secret)
	if !valid {
		jsonResponse(w, "Invalid otp code", http.StatusForbidden)
		return
	}

	user, err := models.GetUserByUsername(p.Username)
	if err != nil {
		jsonResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = ws.store.Set("user_id", user.ID, r, w)
	if err != nil {
		jsonResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, "Login Success", http.StatusOK)
}

func generate(username string) (TOTP, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Example.com",
		AccountName: username,
	})
	if err != nil {
		return TOTP{}, err
	}

	// Convert TOTP key into a PNG
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return TOTP{}, err
	}
	png.Encode(&buf, img)

	// display the QR code to the user.
	totp := toStruct(key, buf.Bytes())

	return totp, nil
}

func display(key *otp.Key, data []byte) {
	fmt.Printf("Issuer:       %s\n", key.Issuer())
	fmt.Printf("Account Name: %s\n", key.AccountName())
	fmt.Printf("Secret:       %s\n", key.Secret())
	fmt.Println("Writing PNG to qr-code.png....")
	ioutil.WriteFile("qr-code.png", data, 0644)
	fmt.Println("")
	fmt.Println("Please add your TOTP to your OTP Application now!")
	fmt.Println("")
}

type TOTP struct {
	Issuer      string `json:"issuer"`
	AccountName string `json:"account_name"`
	Secret      string `json:"secret"`
}

func toStruct(key *otp.Key, data []byte) TOTP {
	return TOTP{
		Issuer:      key.Issuer(),
		AccountName: key.AccountName(),
		Secret:      key.Secret(),
	}
}
