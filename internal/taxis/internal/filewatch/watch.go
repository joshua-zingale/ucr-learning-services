package filewatch

import (
	"context"
	"os"
	"time"
)

func RunOnUpdate(ctx context.Context, filePath string, pollingPeriod time.Duration, runOnUpdateProc func()) error {
	for {
		if err := watchFile(ctx, filePath, pollingPeriod); err != nil {
			return err
		}
		runOnUpdateProc()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			continue
		}
	}
}

// Source - https://stackoverflow.com/questions/8270441/go-language-how-detect-file-changing
// Posted by laurent, modified by community. See post 'Timeline' for change history
// Retrieved 2026-01-27, License - CC BY-SA 3.0
// Modified by Joshua Zingale
func watchFile(ctx context.Context, filePath string, pollingPeriod time.Duration) error {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(pollingPeriod)
	defer ticker.Stop()

	for {

		ticker := time.NewTicker(pollingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				stat, err := os.Stat(filePath)
				if err != nil {
					return err
				}

				if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
					return nil
				}
			}
		}
	}
}
