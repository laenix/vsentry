package forensic

import (
	"bufio"
	"os"
)

type TextParser struct{}

func (p *TextParser) Parse(filePath string) ([]ForensicEvent, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []ForensicEvent
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		events = append(events, ForensicEvent{
			"raw_data":      line,
			"category_name": "Findings",
			"class_name":    "Forensic Evidence",
		})
	}
	return events, nil
}