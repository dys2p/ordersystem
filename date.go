package main

import (
	"time"
)

const dateFmt = "2006-01-02" // less metadata than hours, minutes, seconds

type Date string // dateFmt

func Today() Date {
	return Date(time.Now().Format(dateFmt))
}

func (d Date) Format() (string, error) {
	var t, err = d.Parse()
	if err != nil {
		return "", err
	}
	return t.Format("02.01.2006"), nil
}

func (d Date) Parse() (time.Time, error) {
	return time.Parse(dateFmt, string(d))
}
