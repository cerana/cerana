package metrics_test

func (s *metrics) TestHost() {
	result, streamURL, err := s.metrics.Host(nil)
	if !s.NoError(err) {
		return
	}
	s.Nil(streamURL)
	if !s.NotNil(result) {
		return
	}
}
