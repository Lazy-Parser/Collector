package wsclient

type Subscription struct {
	max      int                 // 30
	channels map[string]struct{} // only unique
}

func NewSubscription(max int) *Subscription {
	return &Subscription{max: max, channels: make(map[string]struct{}, max)}
}

func (s *Subscription) Push(channel string) bool {
	if len(s.channels) < s.max {

		s.channels[channel] = struct{}{}
		return true
	}
	return false
}

// if not found, do nothing
func (s *Subscription) TryRemove(channel string) bool {
	ok := s.Exists(channel)
	if ok {
		delete(s.channels, channel)
	}
	return ok
}

func (s *Subscription) GetChannelsList() []string {
	list := make([]string, len(s.channels))
	i := 0
	for key := range s.channels {
		list[i] = key
		i++
	}

	return list
}

func (s *Subscription) ClearChannels() {
	for key := range s.channels {
		delete(s.channels, key)
	}
}

func (s *Subscription) Exists(channel string) bool { _, ok := s.channels[channel]; return ok }
func (s *Subscription) Size() int                  { return len(s.channels) }
func (s *Subscription) IsFull() bool               { return len(s.channels) == s.max }