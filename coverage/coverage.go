package coverage

import (
	"fmt"
	"os"
)

func BranchCoverage(filePath string, id int, cond bool) bool {
	OutputCoverage(filePath, fmt.Sprintln(id, cond))
	return cond
}
func OutputCoverage(filePath string, id string) {
	coverageFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	coverageFile.WriteString(id)
	coverageFile.Close()
}
