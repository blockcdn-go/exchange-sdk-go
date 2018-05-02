package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

func floatFromString(raw interface{}) (float64, error) {
	str, ok := raw.(string)
	if !ok {
		return 0, fmt.Errorf(fmt.Sprintf("unable to parse, value not string: %T", raw))
	}
	flt, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, warpError(err, fmt.Sprintf("unable to parse as float: %s", str))
	}
	return flt, nil
}

func intFromString(raw interface{}) (int, error) {
	str, ok := raw.(string)
	if !ok {
		return 0, fmt.Errorf(fmt.Sprintf("unable to parse, value not string: %T", raw))
	}
	n, err := strconv.Atoi(str)
	if err != nil {
		return 0, warpError(err, fmt.Sprintf("unable to parse as int: %s", str))
	}
	return n, nil
}

func timeFromUnixTimestampString(raw interface{}) (time.Time, error) {
	str, ok := raw.(string)
	if !ok {
		return time.Time{}, fmt.Errorf(fmt.Sprintf("unable to parse, value not string"))
	}
	ts, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return time.Time{}, warpError(err, fmt.Sprintf("unable to parse as int: %s", str))
	}
	return time.Unix(0, ts*int64(time.Millisecond)), nil
}

func timeFromUnixTimestampFloat(raw interface{}) (time.Time, error) {
	ts, ok := raw.(float64)
	if !ok {
		return time.Time{}, fmt.Errorf(fmt.Sprintf("unable to parse, value not int64: %T", raw))
	}
	return time.Unix(0, int64(ts)*int64(time.Millisecond)), nil
}

func unixMillis(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func recvWindow(d time.Duration) int64 {
	return int64(d) / int64(time.Millisecond)
}

func (as *apiService) handleError(textRes []byte) error {
	err := &Error{}
	log.Println("errorResponse:", string(textRes))
	if err := json.Unmarshal(textRes, err); err != nil {
		return warpError(err, "error unmarshal failed")
	}
	return err
}

func warpError(err error, msg string) error {
	return fmt.Errorf(err.Error() + msg)
}
