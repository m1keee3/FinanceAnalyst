package chart

import (
	"fmt"
	"os/exec"
	"time"
)

func Save(ticker string, from time.Time, till time.Time, dirName, fileName string) error {
	fromStr := from.Format("2006-01-02")
	tillStr := till.Format("2006-01-02")

	cmd := exec.Command("python", "pkg\\chart\\save_chart.py", ticker, fromStr, tillStr, dirName, fileName)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error to run python scrypt: %v\n%s", err, string(output))
	}

	fmt.Println(string(output))
	return nil
}
