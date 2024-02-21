package kratos

import "time"

var (
	locationJst                  *time.Location
	privilegedAccessLimitMinutes time.Duration
)

type InitInput struct {
	PrivilegedAccessLimitMinutes time.Duration
}

func Init(i InitInput) {
	privilegedAccessLimitMinutes = i.PrivilegedAccessLimitMinutes

	var err error
	locationJst, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
}
