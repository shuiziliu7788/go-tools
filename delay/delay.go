package delay

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptrace"
	"sort"
	"sync"
	"time"
)

type Delay struct {
	Uri      string
	Duration time.Duration
}

func TestDelay(urls []string) ([]Delay, error) {
	var (
		wg    sync.WaitGroup
		hosts []Delay
	)
	wg.Add(len(urls))

	for _, url := range urls {
		go func(url string) {
			var (
				connectStart, tlsStart       time.Time
				connectDuration, tlsDuration time.Duration
			)
			defer wg.Done()
			trace := &httptrace.ClientTrace{
				ConnectStart: func(network, addr string) {
					connectStart = time.Now()
				},
				ConnectDone: func(network, addr string, err error) {
					connectDuration = time.Since(connectStart)
				},
				TLSHandshakeStart: func() {
					tlsStart = time.Now()
				},
				TLSHandshakeDone: func(state tls.ConnectionState, err error) {
					tlsDuration = time.Since(tlsStart)
				},
			}
			req, _ := http.NewRequest("HEAD", url, nil)
			req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
			resp, err := http.DefaultTransport.RoundTrip(req)
			if err != nil {
				return
			}
			resp.Body.Close()
			hosts = append(hosts, Delay{
				Uri:      url,
				Duration: connectDuration + tlsDuration,
			})
		}(url)
	}
	wg.Wait()
	if len(hosts) == 0 {
		return nil, errors.New("no valid hosts found")
	}
	sort.Slice(hosts, func(i, j int) bool {
		return hosts[i].Duration < hosts[j].Duration
	})
	return hosts, nil
}

func TestRegion(urls map[string]string) (string, error) {
	var ip struct {
		CountryCode string `json:"country_code"`
	}
	if resp, err := http.Get("https://ip234.in/ip.json"); err == nil {
		_ = json.NewDecoder(resp.Body).Decode(&ip)
		resp.Body.Close()
	}
	if _, ok := urls[ip.CountryCode]; !ok {
		return "", errors.New("no valid region found")
	}
	return urls[ip.CountryCode], nil
}
