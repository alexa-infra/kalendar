package calendar

import (
	"fmt"
	"time"
)

const (
	oneDay       = 24 * time.Hour
	headerFormat = "January 2006"
	weekdaysLine = "Mo Di Mi Do Fr Sa So"
)

func firstMonthDay(dt time.Time) time.Time {
	month := dt.Month()
	for dt.Add(-oneDay).Month() == month {
		dt = dt.Add(-oneDay)
	}
	return dt
}

func lastMonthDay(dt time.Time) time.Time {
	month := dt.Month()
	for dt.Month() == month {
		dt = dt.Add(oneDay)
	}
	return dt
}

func fistWeekDay(dt time.Time) time.Time {
	for dt.Weekday() != time.Monday {
		dt = dt.Add(-oneDay)
	}
	return dt
}

func lastWeekDay(dt time.Time) time.Time {
	for dt.Weekday() != time.Sunday {
		dt = dt.Add(oneDay)
	}
	return dt
}

func GetCalendarText(t time.Time) []string {
	t = t.Truncate(oneDay)
	month := t.Month()
	firstDay := fistWeekDay(firstMonthDay(t))
	lastDay := lastWeekDay(lastMonthDay(t))
	lines := []string{}
	space := "   "
	lines = append(lines, space+t.Format(headerFormat))
	lines = append(lines, weekdaysLine)
	line := ""
	for _, dt := range iterDays(firstDay, lastDay) {
		if dt.Month() == month {
			line += fmt.Sprintf("%2d ", dt.Day())
		} else {
			line += space
		}
		if dt.Weekday() == time.Sunday {
			lines = append(lines, line)
			line = ""
		}
	}
	return lines
}

func iterDays(firstDay, lastDay time.Time) (days []time.Time) {
	lastDay = lastDay.Add(oneDay)
	for dt := firstDay; dt.Before(lastDay); dt = dt.Add(oneDay) {
		days = append(days, dt)
	}
	return
}

func iterCalendarDays(firstDay, lastDay, today time.Time) (days []CalendarDay) {
	for _, dt := range iterDays(firstDay, lastDay) {
		days = append(days, NewCalendarDay(dt, today))
	}
	return
}

func GetCalendar(t time.Time) (days []CalendarDay) {
	today := t.Truncate(oneDay)
	firstDay := fistWeekDay(firstMonthDay(today))
	lastDay := lastWeekDay(lastMonthDay(today))
	for _, day := range iterCalendarDays(firstDay, lastDay, today) {
		days = append(days, day)
	}
	return
}

type CalendarDay struct {
	today, weekend, thisMonth bool
	time.Time
}

func (day CalendarDay) Today() bool {
	return day.today
}

func (day CalendarDay) Weekend() bool {
	return day.weekend
}

func (day CalendarDay) ThisMonth() bool {
	return day.thisMonth
}

func (day CalendarDay) Day() int {
	return day.Day()
}

func NewCalendarDay(t, today time.Time) CalendarDay {
	isToday := t == today
	isWeekend := t.Weekday() == time.Sunday || t.Weekday() == time.Saturday
	isThisMonth := t.Month() == today.Month() && t.Year() == today.Year()
	return CalendarDay{isToday, isWeekend, isThisMonth, t}
}
