package metrics_test

func (s *metrics) TestCPU() {
	result, streamURL, err := s.metrics.CPU(nil)
	if !s.NoError(err) {
		return
	}
	s.Nil(streamURL)
	if !s.NotNil(result) {
		return
	}
}
