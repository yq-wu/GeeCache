package geecache

import (
	"GeeCache/geecachepb"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"net/http"
	"net/url"
)

type PeerPicker interface {
	PeerPick(key string) (peer PeerGetter, ok bool)
}
type PeerGetter interface {
	Get(in *geecachepb.Request, out *geecachepb.Response) error
}

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(in *geecachepb.Request, out *geecachepb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}
