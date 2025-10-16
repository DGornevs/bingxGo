package cron

import (
	"fmt"
	"time"
)

// scheduleDaily runs the given task every day at the specified HH:MM time.
func ScheduleDaily(target string, task func()) {
	go func() {
		for {
			now := time.Now()
			loc := now.Location()

			// Parse HH:MM
			var hour, min int
			fmt.Sscanf(target, "%d:%d", &hour, &min)

			// Build next occurrence
			next := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, loc)
			if next.Before(now) {
				next = next.Add(24 * time.Hour)
			}

			fmt.Printf("[Scheduler] Next daily task at %v\n", next.Format("Mon 15:04:05"))
			time.Sleep(time.Until(next))

			task()
		}
	}()
}

// ScheduleWeekly runs the given task every week on a specific weekday at the given HH:MM time.
func ScheduleWeekly(weekday time.Weekday, target string, task func()) {
	go func() {
		for {
			now := time.Now()
			loc := now.Location()

			var hour, min int
			fmt.Sscanf(target, "%d:%d", &hour, &min)

			// Start of today
			next := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, loc)

			// Adjust to correct weekday
			daysUntil := (int(weekday) - int(now.Weekday()) + 7) % 7
			if daysUntil == 0 && next.Before(now) {
				daysUntil = 7
			}
			next = next.Add(time.Duration(daysUntil) * 24 * time.Hour)

			fmt.Printf("[Scheduler] Next weekly task on %v at %v\n",
				next.Weekday(), next.Format("15:04:05"))

			time.Sleep(time.Until(next))

			task()
		}
	}()
}
