package nextbus

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// DefaultClient uses the default http client to make requests
var DefaultClient = &Client{http.DefaultClient}

// Client is used to make requests
type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient}
}

type AgencyResponse struct {
	XMLName    xml.Name `xml:"body"`
	AgencyList []Agency `xml:"agency"`
}

type Agency struct {
	XMLName     xml.Name `xml:"agency"`
	Tag         string   `xml:"tag,attr"`
	Title       string   `xml:"title,attr"`
	RegionTitle string   `xml:"regionTitle,attr"`
}

func (c *Client) GetAgencyList() ([]Agency, error) {
	resp, httpErr := c.httpClient.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=agencyList")
	if httpErr != nil {
		return nil, httpErr
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	var a AgencyResponse
	xmlErr := xml.Unmarshal(body, &a)
	if xmlErr != nil {
		return nil, xmlErr
	}
	return a.AgencyList, nil
}

type RouteResponse struct {
	XMLName   xml.Name `xml:"body"`
	RouteList []Route  `xml:"route"`
}

type Route struct {
	XMLName xml.Name `xml:"route"`
	Tag     string   `xml:"tag,attr"`
	Title   string   `xml:"title,attr"`
}

func (c *Client) GetRouteList(agencyTag string) ([]Route, error) {
	resp, httpErr := c.httpClient.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=routeList&a=" + agencyTag)
	if httpErr != nil {
		return nil, httpErr
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	var a RouteResponse
	xmlErr := xml.Unmarshal(body, &a)
	if xmlErr != nil {
		return nil, xmlErr
	}
	return a.RouteList, nil
}

type RouteConfigResponse struct {
	XMLName   xml.Name      `xml:"body"`
	RouteList []RouteConfig `xml:"route"`
}

type RouteConfig struct {
	XMLName       xml.Name  `xml:"route"`
	StopList      []Stop    `xml:"stop"`
	Tag           string    `xml:"tag,attr"`
	Title         string    `xml:"title,attr"`
	Color         string    `xml:"color,attr"`
	OppositeColor string    `xml:"oppositeColor,attr"`
	LatMin        string    `xml:"latMin,attr"`
	LatMax        string    `xml:"latMax,attr"`
	LonMin        string    `xml:"lonMin,attr"`
	LonMax        string    `xml:"lonMax,attr"`
	Dir           Direction `xml:"direction"`
	PathList      []Path    `xml:"path"`
}

type Stop struct {
	XMLName xml.Name `xml:"stop"`
	Tag     string   `xml:"tag,attr"`
	Title   string   `xml:"title,attr"`
	Lat     string   `xml:"lat,attr"`
	Lon     string   `xml:"lon,attr"`
	StopId  string   `xml:"stopId,attr"`
}

type Direction struct {
	XMLName        xml.Name     `xml:"direction"`
	Tag            string       `xml:"tag,attr"`
	Title          string       `xml:"title,attr"`
	Name           string       `xml:"name,attr"`
	UseForUI       string       `xml:"useForUI,attr"`
	StopMarkerList []StopMarker `xml:"stop"`
}

type StopMarker struct {
	XMLName xml.Name `xml:"stop"`
	Tag     string   `xml:"tag,attr"`
}

type Path struct {
	XMLName   xml.Name `xml:"path"`
	PointList []Point  `xml:"point"`
}

type Point struct {
	XMLName xml.Name `xml:"point"`
	Lat     string   `xml:"lat,attr"`
	Lon     string   `xml:"lon,attr"`
}

// RouteConfigParam is a configuration parameters for GetRouteConfig.
type RouteConfigParam func() string

// RouteConfigTag creates a RouteConfigParam that restricts a
// GetRouteConfig call to a single route.
func RouteConfigTag(tag string) RouteConfigParam {
	return func() string {
		return "r=" + url.QueryEscape(tag)
	}
}

// RouteConfigTerse configures a GetRouteConfig call to avoid path results
func RouteConfigTerse() RouteConfigParam {
	return func() string {
		return "terse"
	}
}

// RouteConfigVerbose configures a GetRouteConfig call to include directions
// not normally shown in UIs.
func RouteConfigVerbose() RouteConfigParam {
	return func() string {
		return "verbose"
	}
}

func (c *Client) GetRouteConfig(agencyTag string, configParams ...RouteConfigParam) ([]RouteConfig, error) {
	params := []string{"command=routeConfig", "a=" + url.QueryEscape(agencyTag)}
	for _, cp := range configParams {
		params = append(params, cp())
	}
	resp, httpErr := c.httpClient.Get("http://webservices.nextbus.com/service/publicXMLFeed?" + strings.Join(params, "&"))
	if httpErr != nil {
		return nil, httpErr
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	var a RouteConfigResponse
	xmlErr := xml.Unmarshal(body, &a)
	if xmlErr != nil {
		return nil, xmlErr
	}
	return a.RouteList, nil
}

type PredictionResponse struct {
	XMLName            xml.Name         `xml:"body"`
	PredictionDataList []PredictionData `xml:"predictions"`
}

type PredictionData struct {
	XMLName                 xml.Name              `xml:"predictions"`
	PredictionDirectionList []PredictionDirection `xml:"direction"`
	MessageList             []Message             `xml:"message"`
	AgencyTitle             string                `xml:"agencyTitle,attr"`
	RouteTitle              string                `xml:"routeTitle,attr"`
	RouteTag                string                `xml:"routeTag,attr"`
	StopTitle               string                `xml:"stopTitle,attr"`
	StopTag                 string                `xml:"stopTag,attr"`
}

type PredictionDirection struct {
	XMLName        xml.Name     `xml:"direction"`
	PredictionList []Prediction `xml:"prediction"`
	Title          string       `xml:"title,attr"`
}

type Prediction struct {
	XMLName           xml.Name `xml:"prediction"`
	EpochTime         string   `xml:"epochTime,attr"`
	Seconds           string   `xml:"seconds,attr"`
	Minutes           string   `xml:"minutes,attr"`
	IsDeparture       string   `xml:"isDeparture,attr"`
	DirTag            string   `xml:"dirTag,attr"`
	Vehicle           string   `xml:"vehicle,attr"`
	VehiclesInConsist string   `xml:"vehiclesInConsist,attr"`
	Block             string   `xml:"block,attr"`
	TripTag           string   `xml:"tripTag,attr"`
}

type Message struct {
	XMLName  xml.Name `xml:"message"`
	Text     string   `xml:"text,attr"`
	Priority string   `xml:"priority,attr"`
}

func (c *Client) GetPredictions(agencyTag string, routeTag string, stopTag string) ([]PredictionData, error) {
	resp, httpErr := c.httpClient.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=predictions&a=" + agencyTag + "&r=" + routeTag + "&s=" + stopTag)
	if httpErr != nil {
		return nil, httpErr
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	var a PredictionResponse
	xmlErr := xml.Unmarshal(body, &a)
	if xmlErr != nil {
		return nil, xmlErr
	}
	return a.PredictionDataList, nil
}

type LocationResponse struct {
	XMLName     xml.Name          `xml:"body"`
	VehicleList []VehicleLocation `xml:"vehicle"`
	LastTime    LocationLastTime  `xml:"lastTime"`
}

type VehicleLocation struct {
	XMLName          xml.Name `xml:"vehicle"`
	ID               string   `xml:"id,attr"`
	RouteTag         string   `xml:"routeTag,attr"`
	DirTag           string   `xml:"dirTag,attr"`
	Lat              string   `xml:"lat,attr"`
	Lon              string   `xml:"lon,attr"`
	SecsSinceReport  string   `xml:"secsSinceReport,attr"`
	Predictable      string   `xml:"predictable,attr"`
	Heading          string   `xml:"heading,attr"`
	SpeedKmHr        string   `xml:"speedKmHr,attr"`
	LeadingVehicleID string   `xml:"leadingVehicleId,attr"`
}

type LocationLastTime struct {
	XMLName xml.Name `xml:"lastTime"`
	Time    string   `xml:"time,attr"`
}

type VehicleLocationParam func() string

func VehicleLocationRoute(routeTag string) VehicleLocationParam {
	return func() string {
		return "r=" + url.QueryEscape(routeTag)
	}
}

func VehicleLocationTime(t string) VehicleLocationParam {
	return func() string {
		return "t=" + url.QueryEscape(t)
	}
}

func (c *Client) GetVehicleLocations(agencyTag string, configParams ...VehicleLocationParam) (*LocationResponse, error) {
	params := []string{"command=vehicleLocations", "a=" + url.QueryEscape(agencyTag)}
	timeWasSet := false
	for _, cp := range configParams {
		paramText := cp()
		if strings.HasPrefix(paramText, "t=") {
			timeWasSet = true
		}
		params = append(params, paramText)
	}
	if !timeWasSet {
		params = append(params, VehicleLocationTime("0")())
	}
	resp, err := c.httpClient.Get("http://webservices.nextbus.com/service/publicXMLFeed?" + strings.Join(params, "&"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result LocationResponse
	if err = xml.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
