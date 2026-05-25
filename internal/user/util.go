package user

import "time"

func nowUTC() time.Time { return time.Now().UTC() }

func ptrString(s string) *string { return &s }
