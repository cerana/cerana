package metrics_test

import "encoding/json"

func (s *metrics) TestHardware() {
	res, streamURL, err := s.metrics.Hardware(nil)
	s.Nil(streamURL)
	if !s.NoError(err) {
		return
	}
	if !s.NotNil(res) {
		return
	}
	_, err = json.Marshal(res)
	s.NoError(err)
}
