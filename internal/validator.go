package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type validator struct {
	regexpDomain *regexp.Regexp
}

func newValidator() *validator {
	return &validator{
		regexpDomain: regexp.MustCompile(`^(https?:\/\/)?(www\.)?([\w\.]+\.[a-z]+\.?)(\/[\w\.]*)*\/?$`),
	}
}

func (v *validator) parseDomain(dirtyURLs string) ([]string, error) {
	if dirtyURLs == "" {
		return nil, fmt.Errorf(`domains not found`)
	}
	var urls []string
	found := make(map[string]bool)
	fields := strings.Fields(dirtyURLs)
	for _, f := range fields {
		f := strings.TrimSpace(f)
		if f == "" {
			continue
		}
		matches := v.regexpDomain.FindAllStringSubmatch(f, -1)
		if len(matches) == 0 {
			continue
		}
		domain := strings.Trim(matches[0][3], `.`)
		if !found[domain] {
			urls = append(urls, matches[0][3])
			found[domain] = true
		}
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf(`domains not found`)
	}
	return urls, nil
}

func (v *validator) parseID(dirtyID string) (int64, error) {
	if dirtyID == "" {
		return 0, fmt.Errorf(`id is required`)
	}
	id, err := strconv.ParseInt(dirtyID, 10, 64)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, fmt.Errorf(`id must > 0`)
	}
	return id, nil
}
