package wsclient

import "fmt"

type Subscription struct {
	max      int                 // 30
	channels map[string]struct{} // only unique
}

func NewSubscription(max int) Subscription {
	return Subscription{max: max, channels: make(map[string]struct{}, max)}
}

func (s *Subscription) Push(channel string) bool {
	if len(s.channels) < s.max {

		s.channels[channel] = struct{}{}
		return true
	}
	return false
}

func (s *Subscription) TryRemove(channel string) bool {
	if _, ok := s.channels[channel]; ok {
		delete(s.channels, channel)
		return true
	}
	return false
}

func (s *Subscription) Contains(channel string) bool { _, ok := s.channels[channel]; return ok }

func (s *Subscription) Size() int    { return len(s.channels) }
func (s *Subscription) IsFull() bool { return len(s.channels) == s.max }

func (s *Subscription) ToString() string {
	var str string

	i := 0
	for channel := range s.channels {
		str += fmt.Sprintf("\"%s\"", channel)
		if i != len(s.channels)-1 {
			str += ", "
		}

		i++
	}

	return str
}
