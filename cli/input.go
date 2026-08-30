package cli

import "strings"

func (s *Session) GetInput() string {
	go func() {
		input, err := s.stdinReader.ReadString('\n')
		if err != nil {
			return
		}
		s.inc <- strings.ToLower(strings.TrimSpace(input))
	}()

	select {
	case in := <-s.inc:
		return in
	case <-s.sc:
		return ""
	}
}
