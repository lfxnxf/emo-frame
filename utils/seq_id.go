package utils

import "github.com/google/uuid"

func GetSeqID() string {
	return uuid.New().String()
}
