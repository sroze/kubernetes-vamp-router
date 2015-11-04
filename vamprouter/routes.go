package vamprouter

type Filter struct {
	Name string `json:"name"`
	Condition string `json:"condition"`
	Destination string `json:"destination"`
}

type Quota struct {
	SampleWindow string `json:"sampleWindow"`
	Rate int `json:"rate"`
	ExpiryTime string `json:"expiryTime"`
}

type Server struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int `json:"port"`
}

type Service struct {
	Name string `json:"name"`
	Weight int `json:"weight"`
	Servers []Server `json:"servers,omitempty"`
}


const (
	ProtocolHttp string = "http"
	ProtocolTcp string = "tcp"
)

type Route struct {
	Name string `json:"name"`
	Port int `json:"port"`
	Protocol string `json:"protocol"`
	Filters []Filter `json:"filters"`
	HttpQuota Quota `json:"httpQuota,omitempty"`
	TcpQuota Quota `json:"tcpQuota,omitempty"`
	Services []Service `json:"services"`
}

// Get a route by its name
//
//
func (c *Client) GetRoute(name string) (*Route, error) {
	var route Route
	return &route, c.Get(&route, "/v1/routes/"+name)
}

func (c *Client) UpdateRoute(route *Route) (*Route, error) {
	var resp errorResp
	return route, c.Put(&resp, "/v1/routes/"+route.Name, route)
}

func (c *Client) CreateRoute(route *Route) (*Route, error) {
	var resp errorResp
	return route, c.Post(&resp, "/v1/routes", route)
}
