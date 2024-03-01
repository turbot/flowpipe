package schedule

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/utils"
)

func generateHourCron(distributionID string) (int64, error) {
	// Will produce number in range 0-23
	hourCron, err := utils.DistributedStringIndex(distributionID, "", 23)
	if err != nil {
		slog.Error("Unable to generate hour distributed string index", "error", err)
		return 0, err
	}

	return hourCron, nil
}

func generateMinuteCron(distributionID string) (int64, error) {
	// Will produce number in range 0-23
	minuteCron, err := utils.DistributedStringIndex(distributionID, "", 59)
	if err != nil {
		slog.Error("Unable to generate minute distributed string index", "error", err)
		return 0, err
	}

	return minuteCron, nil
}

func generateDayCron(distributionID string) (int64, error) {
	// Will produce number in range 0-6
	dayCron, err := utils.DistributedStringIndex(distributionID, "", 6)
	if err != nil {
		slog.Error("Unable to generate day distributed string index", "error", err)
		return 0, err
	}

	return dayCron, nil
}

func IntervalToCronExpression(id, interval string) (string, error) {

	switch interval {

	case "weekly":
		hourCron, err := generateHourCron(id)
		if err != nil {
			return "", err
		}

		minuteCron, err := generateMinuteCron(id)
		if err != nil {
			return "", err
		}

		dayCron, err := generateDayCron(id)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%d %d * * %d", minuteCron, hourCron, dayCron), nil
	case "daily", "24h":
		hourCron, err := generateHourCron(id)
		if err != nil {
			return "", err
		}

		minuteCron, err := generateMinuteCron(id)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%d %d * * *", minuteCron, hourCron), nil
	case "hourly", "60m", "1h":
		minuteCron, err := generateMinuteCron(id)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%d * * * *", minuteCron), nil

	case "5m", "10m", "15m", "30m", "2h", "4h", "6h", "8h", "12h":
		return durationToCron(id, interval)

	default:
		return "", perr.BadRequestWithMessage("Invalid Interval Request passed for Pipeline")
	}
}

func durationToCron(id, duration string) (string, error) {
	if strings.HasSuffix(duration, "m") {
		minuteString := strings.TrimSuffix(duration, "m")
		minuteInt, err := strconv.Atoi(minuteString)
		if err != nil {
			return "", perr.BadRequestWithMessage("Invalid Duration Request passed for Pipeline")
		}

		offset, err := utils.DistributedStringIndex(id, "", int64(minuteInt-1))
		if err != nil {
			return "", err
		}

		cron := fmt.Sprintf("%d-59/%d * * * *", offset, minuteInt)
		return cron, err

	} else if strings.HasSuffix(duration, "h") {
		hourString := strings.TrimSuffix(duration, "h")
		hourInt, err := strconv.Atoi(hourString)
		if err != nil {
			return "", perr.BadRequestWithMessage("Invalid Duration Request passed for Pipeline")
		}

		offset, err := utils.DistributedStringIndex(id, "", int64(hourInt-1))
		if err != nil {
			return "", err
		}

		minute, err := utils.DistributedStringIndex(id, "", 59)
		if err != nil {
			return "", err
		}

		cron := fmt.Sprintf("%d %d-23/%d * * *", minute, offset, hourInt)
		return cron, nil
	}

	return "", perr.BadRequestWithMessage("Invalid Duration Request passed for Pipeline")
}
