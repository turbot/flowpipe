package schedule

import (
	"fmt"
	"log/slog"
	"time"

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
	if interval == "weekly" {
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
	} else if interval == "daily" {
		hourCron, err := generateHourCron(id)
		if err != nil {
			return "", err
		}

		minuteCron, err := generateMinuteCron(id)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%d %d * * *", minuteCron, hourCron), nil
	} else if interval == "hourly" {
		minuteCron, err := generateMinuteCron(id)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%d * * * *", minuteCron), nil
	}

	return "", perr.BadRequestWithMessage("Invalid Interval Request passed for Pipeline")
}

func DurationToCron(duration time.Duration) (string, error) {
	if duration >= 24*time.Hour {
		return "", fmt.Errorf("duration must be less than 24 hours")
	}

	if duration%time.Hour == 0 {
		// Duration is in whole hours
		hours := duration / time.Hour
		return fmt.Sprintf("0 */%d * * *", hours), nil
	} else if duration%time.Minute == 0 {
		// Duration is in whole minutes
		minutes := duration / time.Minute
		return fmt.Sprintf("*/%d * * * *", minutes), nil
	}

	return "", fmt.Errorf("duration must be in whole minutes or hours")
}
