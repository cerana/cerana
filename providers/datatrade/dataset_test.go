package datatrade_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/datatrade"
)

func (p *Provider) TestDatasetImport() {
	data := []byte("foobar")
	stream := bytes.NewBuffer(data)
	streamURL, err := p.tracker.NewStreamUnix(p.config.StreamDir("dataset-import"), ioutil.NopCloser(stream))
	p.Require().NoError(err)

	p.clusterConf.Data.Nodes["localhost"] = &clusterconf.Node{ID: "localhost"}

	tests := []struct {
		redundancy  int
		streamURL   *url.URL
		expectedErr string
	}{
		{0, nil, "missing arg: redundancy"},
		{1, nil, "missing request streamURL"},
		{1, streamURL, ""},
	}

	for _, test := range tests {
		desc := fmt.Sprintf("%+v", test)
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task:         "import-dataset",
			ResponseHook: p.responseHook,
			StreamURL:    test.streamURL,
			Args: datatrade.DatasetImportArgs{
				Redundancy: test.redundancy,
			},
		})
		p.Require().NoError(err)

		res, streamURL, err := p.provider.DatasetImport(req)
		p.Nil(streamURL)
		if test.expectedErr != "" {
			p.EqualError(err, test.expectedErr, desc)
			p.Nil(res, desc)
			continue
		}
		if !p.Nil(err, desc) {
			continue
		}
		if !p.NotNil(res) {
			continue
		}

		result, ok := res.(datatrade.DatasetImportResult)
		if !p.True(ok, desc) {
			continue
		}

		p.NotNil(p.zfs.Data.Data[filepath.Join(p.config.DatasetDir(), result.Dataset.ID)], desc)
		p.NotNil(p.clusterConf.Data.Datasets[result.Dataset.ID], desc)
	}
}
