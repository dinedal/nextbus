package nextbus

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

const baseURL = "http://webservices.nextbus.com/service/publicXMLFeed"

func makeURL(command string, params ...string) string {
	if len(params) != 0 && len(params)%2 != 0 {
		panic(fmt.Sprintf("Illegal params %q.  Must be 0 len or an even length", params))
	}
	var values []string
	values = append(values, "command="+url.QueryEscape(command))
	for len(params) != 0 {
		values = append(values, url.QueryEscape(params[0])+"="+url.QueryEscape(params[1]))
		params = params[2:]
	}
	return baseURL + "?" + strings.Join(values, "&")
}

// Maps from expected url to xml response
var fakes = map[string]string{
	makeURL("agencyList"): `
<body copyright="just testing">
<agency tag="alpha" title="The First" regionTitle="What a Transit Agency"/>
<agency tag="beta" title="The Second" regionTitle="Never never land"/>
</body>
`,
	makeURL("routeList", "a", "alpha"): `
<body copyright="All data copyright some transit company.">
<route tag="1" title="1-first"/>
<route tag="2" title="2-second"/>
</body>
`,
	makeURL("vehicleLocations", "a", "alpha", "t", "0"): `
<body copyright="All data copyright some transit company.">
<vehicle id="1111" routeTag="1" dirTag="1_outbound" lat="37.77513" lon="-122.41946" secsSinceReport="4" predictable="true" heading="225" speedKmHr="0" leadingVehicleId="1112"/>
<vehicle id="2222" routeTag="2" dirTag="2_inbound" lat="37.74891" lon="-122.45848" secsSinceReport="5" predictable="true" heading="217" speedKmHr="0" leadingVehicleId="2223"/>
<lastTime time="1234567890123"/>
</body>
`,
}

type fakeRoundTripper struct {
	t *testing.T
}

func (f fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		req.Body.Close()
		req.Body = nil
	}

	url := req.URL.String()
	xml, ok := fakes[url]
	if !ok {
		f.t.Fatalf("Unexpected url %q.  fakes=%v", url, fakes)
		return nil, nil
	}

	res := http.Response{}
	res.StatusCode = http.StatusOK
	res.Body = ioutil.NopCloser(strings.NewReader(xml))
	res.Request = req
	return &res, nil
}

func testingClient(t *testing.T) *http.Client {
	httpClient := http.Client{}
	httpClient.Transport = fakeRoundTripper{t}
	return &httpClient
}

func xmlName(s string) xml.Name {
	return xml.Name{Space: "", Local: s}
}

func TestGetAgencyList(t *testing.T) {
	nb := NewClient(testingClient(t))
	found, err := nb.GetAgencyList()
	ok(t, err)

	expected := []Agency{
		Agency{xmlName("agency"), "alpha", "The First", "What a Transit Agency"},
		Agency{xmlName("agency"), "beta", "The Second", "Never never land"},
	}
	equals(t, expected, found)
}

func TestGetRouteList(t *testing.T) {
	nb := NewClient(testingClient(t))
	found, err := nb.GetRouteList("alpha")
	ok(t, err)

	expected := []Route{
		Route{xmlName("route"), "1", "1-first"},
		Route{xmlName("route"), "2", "2-second"},
	}
	equals(t, expected, found)
}

func TestGetVehicleLocations(t *testing.T) {
	nb := NewClient(testingClient(t))
	found, err := nb.GetVehicleLocations("alpha")
	ok(t, err)

	expected := LocationResponse{
		xmlName("body"),
		[]VehicleLocation{
			VehicleLocation{
				xmlName("vehicle"),
				"1111",
				"1",
				"1_outbound",
				"37.77513",
				"-122.41946",
				"4",
				"true",
				"225",
				"0",
				"1112",
			},
			VehicleLocation{
				xmlName("vehicle"),
				"2222",
				"2",
				"2_inbound",
				"37.74891",
				"-122.45848",
				"5",
				"true",
				"217",
				"0",
				"2223",
			},
		},
		LocationLastTime{xmlName("lastTime"), "1234567890123"},
	}
	equals(t, &expected, found)
}

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
