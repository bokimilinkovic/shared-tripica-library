package tripica

import (
	"strconv"
	"strings"
	"time"
)

// Date represents a triPica date formatted in timestamp in milliseconds.
type Date struct {
	time.Time
}

// UnmarshalJSON converts a timestamp in milliseconds to time.Time.
func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	timestamp, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}

	d.Time = time.Unix(0, timestamp*int64(time.Millisecond))

	return nil
}
