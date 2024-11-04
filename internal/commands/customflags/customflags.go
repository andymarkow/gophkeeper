package customflags

import (
	"errors"
	"fmt"
	"strings"
)

type StringMapFlag map[string]string

// Реализация интерфейса `pflag.Value` для типа StringMapFlag

func (s *StringMapFlag) String() string {
	pairs := make([]string, 0, len(*s))

	for key, value := range *s {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(pairs, ",")
}

func (s *StringMapFlag) Set(value string) error {
	parts := strings.Split(value, "=")
	if len(parts) != 2 {
		return errors.New("value must be in format key=value")
	}
	if *s == nil {
		*s = make(map[string]string)
	}

	(*s)[parts[0]] = parts[1]

	return nil
}

func (s *StringMapFlag) Type() string {
	return "map[string]string"
}

func (s *StringMapFlag) GetMap() map[string]string {
	return *s
}
