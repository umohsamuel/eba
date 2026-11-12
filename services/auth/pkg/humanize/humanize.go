package humanize

import (
	"fmt"
	"time"
)

func Duration(d time.Duration) string {
	unit, n := "second", int(d/time.Second)
	switch {
	case d >= time.Hour && d%time.Hour == 0:
		unit, n = "hour", int(d/time.Hour)
	case d >= time.Minute:
		unit, n = "minute", int(d/time.Minute)
	}

	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}
