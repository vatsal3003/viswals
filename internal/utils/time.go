package utils

import (
	"strconv"
	"time"
)

func MillisToTime(ms string) *time.Time {
	// if it is -1 then set to nil
	if ms == "-1" {
		return nil
	}

	// convert milliseconds to timestamp
	millis, _ := strconv.Atoi(ms)
	t := time.Unix(0, int64(millis)*int64(time.Millisecond))

	return &t
}
