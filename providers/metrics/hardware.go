package metrics

import (
	"encoding/json"
	"net/url"
	"os/exec"

	"github.com/cerana/cerana/acomm"
)

// Hardware returns information about the hardware.
func (m *Metrics) Hardware(req *acomm.Request) (interface{}, *url.URL, error) {
	// Note: json output from lshw is broken when specifying classes with `-C`
	lshw := exec.Command("lshw", "-json")
	out, err := lshw.Output()
	if err != nil {
		return nil, nil, err
	}

	var outI interface{}
	if err := json.Unmarshal(out, &outI); err != nil {
		return nil, nil, err
	}

	return outI, nil, nil
}
