package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", nil
	}
	dateTask, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("failed to convert string to date: %w", err)
	}

	repeatRune := []rune(repeat)
	firstRepeatLetter := string(repeatRune[0])
	if firstRepeatLetter != "y" && firstRepeatLetter != "d" && firstRepeatLetter != "w" && firstRepeatLetter != "m" {
		return "", fmt.Errorf("wrong repetition rule: %w")
	}
	var dateAnswer time.Time
	switch firstRepeatLetter {
	case "y":
		dateAnswer = addDateTask(now, dateTask, 1, 0, 0)
	case "d":
		repeatSplit := strings.Split(repeat, " ")
		if len(repeatSplit) != 2 {
			return "", fmt.Errorf("wrong repetition rule ")
		}
		digit, err := convertToIntAndCheck(repeatSplit[1])
		if err != nil {
			return "", err
		}
		dateAnswer = addDateTask(now, dateTask, 0, 0, digit)
	case "w":
		repeatCheck := strings.Split(repeat, " ")
		if len(repeatCheck) != 2 {
			return "", fmt.Errorf("wrong repetition rule")
		}
		lastRepeatRune := repeatRune[2:]
		repeatSplit := strings.Split(string(lastRepeatRune), ",")
		dateAnswer = now.AddDate(1, 0, 0)
		dateComparison := dateTask
		for _, val := range repeatSplit {
			valInt, err := strconv.Atoi(val)
			if err != nil {
				return "", fmt.Errorf("wrong repetition rule")
			}
			if 1 <= valInt && valInt <= 7 {
				weekday := time.Weekday((valInt + 7) % 7)
				daysToTarget := int(weekday - dateTask.Weekday())
				if daysToTarget < 0 {
					daysToTarget += 7
				}
				dateComparison = dateComparison.AddDate(0, 0, daysToTarget)
				if now.After(dateComparison) {
					dateComparison = addDateTask(now, dateComparison, 0, 0, 7)
				}
				if dateAnswer.After(dateComparison) {
					dateAnswer = dateComparison
				}
			} else {
				return "", fmt.Errorf("error when entering the day of the week")
			}
		}
	case "m":
		lastRepeatRune := repeatRune[2:]
		lastStringSplit := strings.Split(string(lastRepeatRune), " ")
		var masDays []string
		var masMonth []string
		var flagWithMonth bool
		dateAnswer = time.Date(now.Year()+1, now.Month(), 1, 0, 0, 0, 0, now.Location())
		dateComparison := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		if len(lastStringSplit) == 2 {
			masDays = strings.Split(lastStringSplit[0], ",")
			masMonth = strings.Split(lastStringSplit[1], ",")
			flagWithMonth = true
		}
		if flagWithMonth {
			for _, valDays := range masDays {
				valDaysInt, err := strconv.Atoi(valDays)
				if err != nil {
					return "", fmt.Errorf("wrong repetition rule: %w", err)
				}
				if 1 < valDaysInt && valDaysInt > 31 {
					return "", fmt.Errorf("wrong repetition rule")
				}
				for _, valMonth := range masMonth {
					valMonthInt, err := strconv.Atoi(valMonth)
					if err != nil {
						return "", fmt.Errorf("wrong repetition rule: %w", err)
					}
					if 1 < valMonthInt && valMonthInt > 12 {
						return "", fmt.Errorf("wrong repetition rule")
					}
					if valDaysInt == -1 && valDaysInt == -2 {
						dateComparison = time.Date(now.Year(), time.Month(valMonthInt)+1, valDaysInt, 0, 0, 0, 0, now.Location())
					} else {
						dateComparison = time.Date(now.Year(), time.Month(valMonthInt), valDaysInt, 0, 0, 0, 0, now.Location())
					}
					if now.Before(dateComparison) && dateAnswer.After(dateComparison) && !now.Equal(dateComparison) {
						dateAnswer = dateComparison
					}
				}
			}

		} else {
			repeatSplit := strings.Split(lastStringSplit[0], ",")
			for _, val := range repeatSplit {
				valInt, err := strconv.Atoi(val)
				if err != nil {
					return "", fmt.Errorf("wrong repetition rule: %w", err)
				}
				if valInt == -1 || valInt == -2 {
					dateComparison = time.Date(dateTask.Year(), dateTask.Month()+1, valInt+1, 0, 0, 0, 0, dateTask.Location())
					if dateAnswer.After(dateComparison) {
						dateAnswer = dateComparison
					}
				} else if 1 <= valInt && valInt < 31 {
					dateComparison = time.Date(dateTask.Year(), dateTask.Month(), valInt, 0, 0, 0, 0, dateTask.Location())
					if now.After(dateComparison) || now.Equal(dateComparison) {
						dateComparison = addDateTask(now, dateComparison, 0, 1, 0)
					}
					if dateAnswer.After(dateComparison) {
						dateAnswer = dateComparison
					}

				} else if valInt == 31 {
					nowMonth := dateTask.Month()
					if nowMonth == time.February || nowMonth == time.April || nowMonth == time.June || nowMonth == time.September || nowMonth == time.November {
						nowMonth = nowMonth + 1
					}
					dateComparison = time.Date(dateTask.Year(), nowMonth, valInt, 0, 0, 0, 0, dateTask.Location())
					if now.After(dateComparison) || now.Equal(dateComparison) {
						dateComparison = addDateTask(now, dateComparison, 0, 1, 0)
					}
					if dateAnswer.After(dateComparison) {
						dateAnswer = dateComparison
					}

				} else {
					return "", fmt.Errorf("wrong repetition rule")
				}
			}
		}

	}

	return dateAnswer.Format("20060102"), nil
}

func convertToIntAndCheck(num string) (int, error) {
	digit, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}
	if digit >= 400 || digit < 0 {
		return 0, fmt.Errorf("the number of days to reschedule the task is greater than 400 or a negative number is specified")
	}
	return digit, nil
}

func addDateTask(now time.Time, dateTask time.Time, y int, m int, d int) time.Time {
	dateTask = dateTask.AddDate(y, m, d)

	for dateTask.Before(now) {
		dateTask = dateTask.AddDate(y, m, d)
	}
	return dateTask
}
