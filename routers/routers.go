package routers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	//"encoding/json"

	//"time"

	"github.com/Sachingeek125/GolangAuth/Userdetails"
	"github.com/Sachingeek125/GolangAuth/db"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	//"github.com/jackyzha0/go-auth-w-mongo/db"
)

var Sessionduration = time.Hour * 24
var store = sessions.NewCookieStore([]byte("secert-key"))

func Routers() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	return r
}

// creates a new user session into mongodb database
func CreateSession(db *mongo.Database, userID string) (*Userdetails.Session, error) {
	session := &Userdetails.Session{
		ID:       primitive.NewObjectID().Hex(),
		CREATED:  time.Now(),
		MODIFIED: time.Now(),
	}
	collection := db.Collection("sessions")
	_, err := collection.InsertOne(context.Background(), session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// getsession retrives a session from database from id
func GetSession(db *mongo.Database, SessionID string) (*Userdetails.Session, error) {
	var session Userdetails.Session
	collection := db.Collection("sessions")
	err := collection.FindOne(context.Background(), bson.M{"_id": SessionID}).Decode(&session)
	if err != nil {
		return nil, err
	}
	session.MODIFIED = time.Now()
	return &session, nil
}

// deletesession deletes a session from database
func DeleteSession(db *mongo.Database, SessionID string) error {
	fmt.Println("Into delete session")
	collection := db.Collection("sessions")
	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": SessionID})
	return err

}

// Register creates a new user in the database and stores a session into database
func Register(w http.ResponseWriter, r *http.Request) {
	var user Userdetails.UserData
	newUser := new(Userdetails.UserData)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parseFormErr := r.ParseForm()
	if parseFormErr != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error: %v", parseFormErr)
		return
	}
	// Decode form values into the newUser struct
	newUser.EMAIL = user.EMAIL
	newUser.PASSWORD = user.PASSWORD
	newUser.BIO = user.BIO
	newUser.DATE_OF_BIRTH = user.DATE_OF_BIRTH
	newUser.FIRST_NAME = user.FIRST_NAME
	newUser.LAST_NAME = user.LAST_NAME
	newUser.USERNAME = user.USERNAME

	if newUser.EMAIL == "" || newUser.PASSWORD == "" {
		http.Error(w, "Email and password both are required", http.StatusBadRequest)
		return
	}
	db := db.Client.Database("exampleDB")

	// check if user already exists

	collection := db.Collection("users")
	fmt.Println(newUser.EMAIL)
	count, err := collection.CountDocuments(context.Background(), bson.M{"email": newUser.EMAIL})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "Email already exist.", http.StatusBadRequest)
		return
	}

	// encrypt the password using bcrypt for security reasons
	hash, err := bcrypt.GenerateFromPassword([]byte(newUser.PASSWORD), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newUser.PASSWORD = string(hash)

	// inserting newuser into database
	_, err = collection.InsertOne(context.Background(), newUser)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create a new session for user
	session, err := CreateSession(db, string(newUser.ID.Hex()))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set session header for user

	w.Header().Set("Session-ID", session.ID)
	// w.Write([]byte("User Created"))
	log.Printf("Created new user with email %v", newUser.EMAIL)
	fmt.Fprintf(w, "User Created with %s", newUser.USERNAME)
	fmt.Printf(newUser.USERNAME)
	w.WriteHeader(http.StatusOK)
	log.Printf("User created with: ")
	log.Printf(newUser.EMAIL)
}

func Login(w http.ResponseWriter, r *http.Request) {

	// 	creds := &Userdetails.Credentilas{}
	// 	if err := json.NewDecoder(r.Body).Decode(creds); err != nil {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		return
	// 	}
	// fmt.Println("Login-1:")
	email := r.FormValue("email")
	password := r.FormValue("password")
	// cred := new(Userdetails.Credentilas)
	// decoder := json.NewDecoder(r.Body)
	// err := decoder.Decode(&cred)
	// fmt.Println(cred.EMAIL)
	// fmt.Println(cred.PASSWORD)
	// fmt.Println(email)
	// fmt.Println(password)
	fmt.Println(email)
	fmt.Println(password)

	if email == "" || password == "" {
		http.Error(w, "Email and password both are required", http.StatusBadRequest)
		return
	}
	// fmt.Println("Login-2:")
	db := db.Client.Database("exampleDB")

	coll := db.Collection("users")
	var user Userdetails.UserData

	// finding a a entered email and decoding user if its exists
	err := coll.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	// fmt.Println("Login-3:")
	// fmt.Println(user.EMAIL)
	// fmt.Println(user.BIO)
	if err != nil {
		// fmt.Println("Login-4:")
		if err == mongo.ErrNoDocuments {
			http.Error(w, "No User exists", http.StatusUnauthorized)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			// fmt.Println("Login-5:")

		}
		return
	}

	// check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PASSWORD), []byte(password))
	// fmt.Println("Login-6:")
	if err != nil {
		http.Error(w, "Invalid password and email", http.StatusUnauthorized)
		return
	}
	// fmt.Println("Login-7:")
	// create a new session
	session, err := CreateSession(db, user.ID.Hex())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// fmt.Println("Login-8:")

	// set session header
	w.Header().Set("Session-ID", session.ID)
	fmt.Fprintf(w, "Your Session-ID %s", session.ID)
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "Logged in as %s", user.USERNAME)

}

// Logout handlers Logout user and deletes it's session
func Logout(w http.ResponseWriter, r *http.Request) {

	sessionID := r.Header.Get("Session-ID")
	fmt.Println(sessionID)
	if sessionID == "" {
		http.Error(w, "Session-ID required", http.StatusBadRequest)
		return
	}
	db := db.Client.Database("exampleDB")

	// retrive the session
	_, err := GetSession(db, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// delete the session
	err = DeleteSession(db, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Session %s deleted", sessionID)

}
