package metrics_test

func (s *metrics) TestNetwork() {
	result, streamURL, err := s.metrics.Network(nil)
	if !s.NoError(err) {
		return
	}
	s.Nil(streamURL)
	if !s.NotNil(result) {
		return
	}
}
