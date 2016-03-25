package metrics_test

import "encoding/json"

func (s *metrics) TestDisk() {
	result, streamURL, err := s.metrics.Disk(nil)
	if !s.NoError(err) {
		return
	}
	s.Nil(streamURL)
	if !s.NotNil(result) {
		return
	}
	_, err = json.Marshal(result)
	s.NoError(err)
}
