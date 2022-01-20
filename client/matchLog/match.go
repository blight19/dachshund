package matchLog

import "strings"

type Matcher struct {
	KeyWords []string
}

func (m Matcher) Match() (chan<- map[string]interface{}, <-chan map[string]interface{}) {
	in := make(chan map[string]interface{})
	out := make(chan map[string]interface{})
	go func() {
		for {
			s := <-in
			for _, keyword := range m.KeyWords {
				if value, ok := s["Error"].(string); ok {
					if strings.Contains(value, keyword) {
						s["Error"] = strings.TrimSpace(value)
						out <- s
					}
				}
			}
		}
	}()
	return in, out
}
