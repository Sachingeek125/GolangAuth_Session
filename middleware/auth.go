package middleware

import (
	"context"
	"net/http"

	// "example.com/m/v2/go/pkg/mod/go.mongodb.org/mongo-driver@v1.2.1/mongo"
	"github.com/Sachingeek125/GolangAuth/Userdetails"
	"github.com/Sachingeek125/GolangAuth/db"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func authentication(req http.HandlerFunc, admincheck bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, cookieFetchErr := r.Cookie("session_token")

		if cookieFetchErr != nil {
			if cookieFetchErr == http.ErrNoCookie {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			log.Warn("Bad Auth Attempt: Could not read cookie.")
			return
		}
		sessionToken := c.Value

		filter := bson.M{"sessionToken": sessionToken}
		var res Userdetails.UserData
		err := db.Users.FindOne(context.Background(), filter).Decode(&res)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				log.Warnf("Bad Attempt: No User Exists with token %s", sessionToken)
				return

			}
			w.WriteHeader(http.StatusInternalServerError)
			log.Warn("Bad Auth Attempt: Internal server error while finding a user.")
			return

		}
		// expireTime, TimeparseErr := time.Parse(time.RFC3339, res.Expires)

		// if TimeparseErr != nil {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	log.Warn("Bad Auth Attempt: Session expiry date is wrong")
		// 	return
		// }
		// if time.Now().After(expireTime) {
		// 	http.Redirect(w, r, "/login", http.StatusSeeOther)
		// 	return

		// }

		r.Header.Set("x-res-Email", res.EMAIL)
		req(w, r)

	}

}
