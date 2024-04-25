package main

import "strings"

func TruncateEmail(email string) string {
	if at := strings.Index(email, "@"); at != -1 {
		return email[:at]
	}
	return email
}
