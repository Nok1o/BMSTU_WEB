package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func ReadString(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func ReadInt(prompt string) int {
	for {
		input := ReadString(prompt)
		if value, err := strconv.Atoi(input); err == nil {
			return value
		}
		fmt.Println("Please enter a valid number")
	}
}

func ReadDate(prompt string) time.Time {
	for {
		input := ReadString(prompt + " (YYYY-MM-DD): ")
		if date, err := time.Parse("2006-01-02", input); err == nil {
			return date
		}
		fmt.Println("Please enter a valid date in YYYY-MM-DD format")
	}
}

func ReadBool(prompt string) bool {
	for {
		input := strings.ToLower(ReadString(prompt + " (y/n): "))
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}
		fmt.Println("Please enter 'y' for yes or 'n' for no")
	}
}

func PrintJSON(data interface{}) {
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonData))
}
