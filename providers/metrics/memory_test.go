package metrics_test

func (s *metrics) TestMemory() {
	result, streamURL, err := s.metrics.Memory(nil)
	if !s.NoError(err) {
		return
	}
	s.Nil(streamURL)
	if !s.NotNil(result) {
		return
	}
}
