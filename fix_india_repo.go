package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	filePath := "/home/abhishek/Desktop/flyfleet/services/microservices/india-integration-service/internal/repository/india_repository_impl.go"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(content), "\n")
	newLines := []string{}

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		newLines = append(newLines, line)

		// Fix GetGSTCompliance Scan
		if i > 1500 && i < 1550 && strings.Contains(line, "&compliance.CreatedAt,") && !strings.Contains(lines[i+1], "&compliance.CreatedBy,") {
			newLines = append(newLines, "\t\t&compliance.CreatedBy,")
		}
		if i > 1500 && i < 1550 && strings.Contains(line, "&compliance.UpdatedAt,") && !strings.Contains(lines[i+1], "&compliance.UpdatedBy,") {
			newLines = append(newLines, "\t\t&compliance.UpdatedBy,")
		}

		// Fix GetGSTComplianceByGSTIN Scan
		if i > 1550 && i < 1600 && strings.Contains(line, "&compliance.CreatedAt,") && !strings.Contains(lines[i+1], "&compliance.CreatedBy,") {
			newLines = append(newLines, "\t\t&compliance.CreatedBy,")
		}
		if i > 1550 && i < 1600 && strings.Contains(line, "&compliance.UpdatedAt,") && !strings.Contains(lines[i+1], "&compliance.UpdatedBy,") {
			newLines = append(newLines, "\t\t&compliance.UpdatedBy,")
		}
	}

	err = ioutil.WriteFile(filePath, []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Fixed GST compliance scans")
}
