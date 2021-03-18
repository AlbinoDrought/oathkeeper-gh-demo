package main

import (
	"net/http"
	"os"
	"strings"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		w.Header().Set("Content-Type", "text/plain")

		if userID == "" {
			w.Write([]byte("Hello logged-out user"))
		} else {

			provider := r.Header.Get("X-User-Provider")
			username := r.Header.Get("X-User-Username")
			email := r.Header.Get("X-User-Email")

			sb := strings.Builder{}
			sb.WriteString("Hello user ")
			sb.WriteString(userID)
			sb.WriteString("\nYou logged in with ")
			sb.WriteString(provider)
			sb.WriteString(" using the username ")
			sb.WriteString(username)
			sb.WriteString(" and the email ")
			sb.WriteString(email)
			sb.WriteRune('\n')

			w.Write([]byte(sb.String()))
		}
	})

	os.Stdout.WriteString("Listening at :3000\n")

	panic(http.ListenAndServe(":3000", nil))
}
