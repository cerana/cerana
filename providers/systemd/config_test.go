package systemd_test

import "fmt"

func (s *systemd) TestUnitFilePath() {
	tests := []struct {
		name   string
		result string
		err    string
	}{
		{"", "", "invalid name"},
		{".", "", "invalid name"},
		{"..", "", "invalid name"},
		{"../ ", "", "invalid name"},
		{"../", "", "invalid name"},
		{"../.", "", "invalid name"},
		{"./ ", "", "invalid name"},
		{"./.", "", "invalid name"},
		{"/ ", "", "invalid name"},
		{"/", "", "invalid name"},
		{"/.", "", "invalid name"},
		{"/..", "", "invalid name"},
		{"/foo", "", "invalid name"},
		{"./foo", "", "invalid name"},
		{"../foo", "", "invalid name"},
		{"/bar/foo", "", "invalid name"},
		{"foo/", "", "invalid name"},
		{" ", s.dir + "/ ", ""},
		{". ", s.dir + "/. ", ""},
		{".. ", s.dir + "/.. ", ""},
		{"foo", s.dir + "/foo", ""},
		{"foo ", s.dir + "/foo ", ""},
		{" foo", s.dir + "/ foo", ""},
	}

	for _, test := range tests {
		testS := fmt.Sprintf("%+v", test)
		result, err := s.config.UnitFilePath(test.name)
		if test.err == "" {
			if !s.NoError(err, testS) {
				continue
			}
			s.Equal(test.result, result, testS)
		} else {
			s.Empty(result, testS)
			s.EqualError(err, test.err, testS)
		}
	}
}

func (s *systemd) TestValidate() {
	defer s.viper.Set("unit_file_dir", s.dir)

	tests := []struct {
		dir interface{}
		err string
	}{
		{"", "missing unit_file_dir"},
		{"foo", ""},
		{".", ""},
		{5, "invalid unit_file_dir"},
	}

	for _, test := range tests {
		testS := fmt.Sprintf("%+v", test)
		s.viper.Set("unit_file_dir", test.dir)
		err := s.config.Validate()
		if test.err == "" {
			s.NoError(err, testS)
		} else {
			s.EqualError(err, test.err, testS)
		}
	}
}
