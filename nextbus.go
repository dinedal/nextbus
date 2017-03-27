package nextbus

import (
	"encoding/xml"
	"fmt"
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

// NewClient creates a new nextbus client.
func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient}
}

// AgencyResponse represents a list of transit agencies.
type AgencyResponse struct {
	XMLName    xml.Name `xml:"body"`
	AgencyList []Agency `xml:"agency"`
}

// Agency represents a single transit agency.
type Agency struct {
	XMLName     xml.Name `xml:"agency"`
	Tag         string   `xml:"tag,attr"`
	Title       string   `xml:"title,attr"`
	RegionTitle string   `xml:"regionTitle,attr"`
}

// GetAgencyList fetches the list of supported transit agencies by nextbus.
func (c *Client) GetAgencyList() ([]Agency, error) {
	resp, httpErr := c.httpClient.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=agencyList")
	if httpErr != nil {
		return nil, fmt.Errorf("could not fetch agencies from nextbus: %v", httpErr)
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("could not parse agencies response body: %v", readErr)
	}

	var a AgencyResponse
	if xmlErr := xml.Unmarshal(body, &a); xmlErr != nil {
		return nil, fmt.Errorf("could not parse agencies XML: %v", xmlErr)
	}
	return a.AgencyList, nil
}

// RouteResponse is a set of transit routes.
type RouteResponse struct {
	XMLName   xml.Name `xml:"body"`
	RouteList []Route  `xml:"route"`
}

// Route is an individual transit route.
type Route struct {
	XMLName xml.Name `xml:"route"`
	Tag     string   `xml:"tag,attr"`
	Title   string   `xml:"title,attr"`
}

// GetRouteList fetches the list of routes within the specified agency.
func (c *Client) GetRouteList(agencyTag string) ([]Route, error) {
	resp, httpErr := c.httpClient.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=routeList&a=" + agencyTag)
	if httpErr != nil {
		return nil, fmt.Errorf("could not fetch routes from nextbus: %v", httpErr)
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("could not parse routes response body: %v", readErr)
	}

	var a RouteResponse
	xmlErr := xml.Unmarshal(body, &a)
	if xmlErr != nil {
		return nil, fmt.Errorf("could not parse routes XML: %v", xmlErr)
	}
	return a.RouteList, nil
}

// RouteConfigResponse is a collection of RouteConfigs.
type RouteConfigResponse struct {
	XMLName   xml.Name      `xml:"body"`
	RouteList []RouteConfig `xml:"route"`
}

// RouteConfig is the metadata for a particular transit route.
type RouteConfig struct {
	XMLName       xml.Name    `xml:"route"`
	StopList      []Stop      `xml:"stop"`
	Tag           string      `xml:"tag,attr"`
	Title         string      `xml:"title,attr"`
	Color         string      `xml:"color,attr"`
	OppositeColor string      `xml:"oppositeColor,attr"`
	LatMin        string      `xml:"latMin,attr"`
	LatMax        string      `xml:"latMax,attr"`
	LonMin        string      `xml:"lonMin,attr"`
	LonMax        string      `xml:"lonMax,attr"`
	DirList       []Direction `xml:"direction"`
	PathList      []Path      `xml:"path"`
}

// Stop is the metadata for a particular stop.
type Stop struct {
	XMLName xml.Name `xml:"stop"`
	Tag     string   `xml:"tag,attr"`
	Title   string   `xml:"title,attr"`
	Lat     string   `xml:"lat,attr"`
	Lon     string   `xml:"lon,attr"`
	StopID  string   `xml:"stopId,attr"`
}

// Direction is the metadata for one individual route direction. A transit route
// usually has at least two "directions": "inbound" and "outbound", for example.
type Direction struct {
	XMLName        xml.Name     `xml:"direction"`
	Tag            string       `xml:"tag,attr"`
	Title          string       `xml:"title,attr"`
	Name           string       `xml:"name,attr"`
	UseForUI       string       `xml:"useForUI,attr"`
	StopMarkerList []StopMarker `xml:"stop"`
}

// StopMarker identifies a particular stop for a direction of a route.
type StopMarker struct {
	XMLName xml.Name `xml:"stop"`
	Tag     string   `xml:"tag,attr"`
}

// Path contains a set of points that define the geographical path of a route.
type Path struct {
	XMLName   xml.Name `xml:"path"`
	PointList []Point  `xml:"point"`
}

// Point contains a latitude and longitude representing a geographical location.
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

// GetRouteConfig fetches the metadata for routes in a particular transit
// agency. Use the configParams to filter the requested data.
func (c *Client) GetRouteConfig(agencyTag string, configParams ...RouteConfigParam) ([]RouteConfig, error) {
	params := []string{"command=routeConfig", "a=" + url.QueryEscape(agencyTag)}
	for _, cp := range configParams {
		params = append(params, cp())
	}
	resp, httpErr := c.httpClient.Get("http://webservices.nextbus.com/service/publicXMLFeed?" + strings.Join(params, "&"))
	if httpErr != nil {
		return nil, fmt.Errorf("could not fetch route config from nextbus: %v", httpErr)
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("could not parse route config response body: %v", readErr)
	}

	var a RouteConfigResponse
	if xmlErr := xml.Unmarshal(body, &a); xmlErr != nil {
		return nil, fmt.Errorf("could not parse route config XML: %v", xmlErr)
	}
	return a.RouteList, nil
}

// PredictionResponse contains a set of predictions.
type PredictionResponse struct {
	XMLName            xml.Name         `xml:"body"`
	PredictionDataList []PredictionData `xml:"predictions"`
}

// PredictionData represents a prediction for a particular route and stop. It
// contains a set of predictions arranged by direction.
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

// PredictionDirection contains a list of arrival predictions for a particular
// route and stop traveling in a specific direction.
type PredictionDirection struct {
	XMLName        xml.Name     `xml:"direction"`
	PredictionList []Prediction `xml:"prediction"`
	Title          string       `xml:"title,attr"`
}

// Prediction is an individual arrival prediction for a particular route, stop,
// and direction.
type Prediction struct {
	XMLName           xml.Name `xml:"prediction"`
	EpochTime         string   `xml:"epochTime,attr"`
	Seconds           string   `xml:"seconds,attr"`
	Minutes           string   `xml:"minutes,attr"`
	IsDeparture       string   `xml:"isDeparture,attr"`
	AffectedByLayover string   `xml:"affectedByLayover,attr"`
	DirTag            string   `xml:"dirTag,attr"`
	Vehicle           string   `xml:"vehicle,attr"`
	VehiclesInConsist string   `xml:"vehiclesInConsist,attr"`
	Block             string   `xml:"block,attr"`
	TripTag           string   `xml:"tripTag,attr"`
}

// Message is an informational message provided by the transit agency.
type Message struct {
	XMLName  xml.Name `xml:"message"`
	Text     string   `xml:"text,attr"`
	Priority string   `xml:"priority,attr"`
}

// GetStopPredictions fetches a set of predictions for a transit agency at the
// provided stop. Note that this requires the 'stopID' which is the unique
// identifier for a stop indepenedent of a route.
func (c *Client) GetStopPredictions(agencyTag string, stopID string) ([]PredictionData, error) {
	resp, httpErr := c.httpClient.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=predictions&a=" + agencyTag + "&stopId=" + stopID)
	if httpErr != nil {
		return nil, fmt.Errorf("could not fetch stop predictions from nextbus: %v", httpErr)
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("could not parse stop predictions response body: %v", readErr)
	}

	var a PredictionResponse
	if xmlErr := xml.Unmarshal(body, &a); xmlErr != nil {
		return nil, fmt.Errorf("could not parse stop predictions XML: %v", xmlErr)
	}
	return a.PredictionDataList, nil
}

// GetPredictions fetches a set of predictions for a transit agency at the
// provided route and stop.
func (c *Client) GetPredictions(agencyTag string, routeTag string, stopTag string) ([]PredictionData, error) {
	resp, httpErr := c.httpClient.Get("http://webservices.nextbus.com/service/publicXMLFeed?command=predictions&a=" + agencyTag + "&r=" + routeTag + "&s=" + stopTag)
	if httpErr != nil {
		return nil, fmt.Errorf("could not fetch predictions from nextbus: %v", httpErr)
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("could not parse predictions response body: %v", readErr)
	}

	var a PredictionResponse
	if xmlErr := xml.Unmarshal(body, &a); xmlErr != nil {
		return nil, fmt.Errorf("could not parse predictions XML: %v", xmlErr)
	}
	return a.PredictionDataList, nil
}

// PredReqParam knows how to configure a request for a multi stop prediction.
type PredReqParam func() string

// PredReqStop specifies a route and stop which we want predictions for.
func PredReqStop(routeTag, stopTag string) PredReqParam {
	return func() string {
		return "stops=" + url.QueryEscape(fmt.Sprintf("%s|%s", routeTag, stopTag))
	}
}

// PredReqShortTitles specifies that we want short titles in our
// predictions response.
func PredReqShortTitles() PredReqParam {
	return func() string {
		return "useShortTitles=true"
	}
}

// GetPredictionsForMultiStops Issues a request to get predictions for multiple stops.
func (c *Client) GetPredictionsForMultiStops(agencyTag string, params ...PredReqParam) ([]PredictionData, error) {
	queryParams := []string{
		"command=predictionsForMultiStops",
		"a=" + url.QueryEscape(agencyTag),
	}
	for _, p := range params {
		queryParams = append(queryParams, p())
	}

	resp, httpErr := c.httpClient.Get("http://webservices.nextbus.com/service/publicXMLFeed?" + strings.Join(queryParams, "&"))
	if httpErr != nil {
		return nil, fmt.Errorf("could not fetch predictions for multiple stops from nextbus: %v", httpErr)
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("could not parse predictions for multiple stops response body: %v", readErr)
	}

	var a PredictionResponse
	if xmlErr := xml.Unmarshal(body, &a); xmlErr != nil {
		return nil, fmt.Errorf("could not parse predictions for multiple stops XML: %v", xmlErr)
	}
	return a.PredictionDataList, nil
}

// LocationResponse is a list of vehicle locations.
type LocationResponse struct {
	XMLName     xml.Name          `xml:"body"`
	VehicleList []VehicleLocation `xml:"vehicle"`
	LastTime    LocationLastTime  `xml:"lastTime"`
}

// VehicleLocation represents the location of an individual vehicle traveling
// on a route.
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

// LocationLastTime represents the last time that a location was reported.
type LocationLastTime struct {
	XMLName xml.Name `xml:"lastTime"`
	Time    string   `xml:"time,attr"`
}

// VehicleLocationParam is used to specify options when fetching vehicle
// locations.
type VehicleLocationParam func() string

// VehicleLocationRoute returns a VehicleLocationParam that indicates the
// desired route to filter vehicle locations by.
func VehicleLocationRoute(routeTag string) VehicleLocationParam {
	return func() string {
		return "r=" + url.QueryEscape(routeTag)
	}
}

// VehicleLocationTime returns a VehicleLocationParam that indicates the
// desired time after which to fetch vehicle locations.
func VehicleLocationTime(t string) VehicleLocationParam {
	return func() string {
		return "t=" + url.QueryEscape(t)
	}
}

// GetVehicleLocations fetches the set of vehicle locations for a transit
// agency. Use the configParams to filter the requested data.
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
	resp, httpErr := c.httpClient.Get("http://webservices.nextbus.com/service/publicXMLFeed?" + strings.Join(params, "&"))
	if httpErr != nil {
		return nil, fmt.Errorf("could not fetch vehicle locations from nextbus: %v", httpErr)
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr == nil {
		return nil, fmt.Errorf("could not parse vehicle locations response body: %v", readErr)
	}

	var result LocationResponse
	if xmlErr := xml.Unmarshal(body, &result); xmlErr != nil {
		return nil, fmt.Errorf("could not parse vehicle locations XML: %v", xmlErr)
	}
	return &result, nil
}
