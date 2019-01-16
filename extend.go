package gorequest

import (
	"encoding/base64"
	"net/http"
	"crypto/hmac"
	"crypto/sha1"
	"sort"
	"io"
)

const qiniuHeaderPrefix = "X-Qiniu-"
//Call this method before End*
func (s *SuperAgent) Sign(sk []byte, ak string) *SuperAgent{
	req, _ := s.MakeRequest()
	sign, err := SignRequest(sk, req)
	if err != nil {
		s.logger.Println("Sign Error:", err)
	}else {
		auth := "Qiniu " + ak + ":" + base64.URLEncoding.EncodeToString(sign)
		s.Header.Set("Authorization", auth)
	}
	return s
}

func SignRequest(sk []byte, req *http.Request) ([]byte, error) {
	h := hmac.New(sha1.New, sk)
	u := req.URL
	data := req.Method + " " + u.Path
	if u.RawQuery != "" {
		data += "?" + u.RawQuery
	}
	io.WriteString(h, data+"\nHost: "+req.Host)

	ctType := req.Header.Get("Content-Type")
	if ctType != "" {
		io.WriteString(h, "\nContent-Type: "+ctType)
	}

	signQiniuHeaderValues(req.Header, h)

	io.WriteString(h, "\n\n")

	if incBody(req, ctType) {
		s2, err2 := SeekClose(req)
		if err2 != nil {
			return nil, err2
		}
		h.Write(s2.Bytes())
	}

	return h.Sum(nil), nil
}

func signQiniuHeaderValues(header http.Header, w io.Writer) {
	var keys []string
	for key, _ := range header {
		if len(key) > len(qiniuHeaderPrefix) && key[:len(qiniuHeaderPrefix)] == qiniuHeaderPrefix {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return
	}

	if len(keys) > 1 {
		sort.Sort(sortByHeaderKey(keys))
	}
	for _, key := range keys {
		io.WriteString(w, "\n"+key+": "+header.Get(key))
	}
}

type sortByHeaderKey []string

func (p sortByHeaderKey) Len() int           { return len(p) }
func (p sortByHeaderKey) Less(i, j int) bool { return p[i] < p[j] }
func (p sortByHeaderKey) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func incBody(req *http.Request, ctType string) bool {

	return req.ContentLength != 0 && req.Body != nil && ctType != "" && ctType != "application/octet-stream"
}