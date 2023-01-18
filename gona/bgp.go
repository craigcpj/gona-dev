package gona

import (
	"fmt"
	"net/url"
	"strconv"
)

type BGPSessions struct {
	Result  string      `json:"result"`
	Message interface{} `json:"message"`
	Data    struct {
		Sessions map[string]*BGPSession
	}
}

type BGPSession struct {
	ID             int         `json:"id"`
	CustomerIP     string      `json:"customer_peer_ip"`
	GroupID        int         `json:"group_id"`
	Locked         string      `json:"locked"`
	Description    string      `json:"description"`
	State          interface{} `json:"state"`
	RoutesReceived interface{} `json:"routes_received"`
	LastUpdate     interface{} `json:"last_update"`
	ConfigStatus   int         `json:"config_status"`
	Password       interface{} `json:"password"`
	Prefixes       []Prefix    `json:"prefixes"`
	ExportList     string      `json:"export_list"`
	Community      interface{} `json:"community"`
	ProviderPeerIP string      `json:"provider_peer_ip"`
	Location       string      `json:"location"`
	Latitude       string      `json:"latitude"`
	Longitude      string      `json:"longitude"`
	GroupName      string      `json:"group_name"`
	ProviderIPType string      `json:"provider_ip_type"`
	ProviderAsn    string      `json:"provider_asn"`
	CustomerAsn    string      `json:"customer_asn"`
}

type Prefix struct {
	ID          int         `json:"id"`
	MbID        int         `json:"mb_id"`
	Prefix      string      `json:"prefix"`
	Append      interface{} `json:"append"`
	RuleType    string      `json:"rule_type"`
	PrefixType  string      `json:"prefix_type"`
	Description string      `json:"description"`
	Date        string      `json:"date"`
	AllowedPps  int         `json:"allowed_pps"`
	BgpGroupID  int         `json:"bgp_group_id"`
	PrefixID    int         `json:"prefix_id"`
}

func (s *BGPSession) IsLocked() bool {
	return "1" == s.Locked
}

func (s *BGPSession) IsProviderIPTypeV4() bool {
	return string(IPv4) == s.ProviderIPType
}

// GetBGPSession external method on Client to get your BGP session
func (c *Client) GetBGPSession(id int) (*BGPSession, error) {
	var sessions struct {
		Data *BGPSession `json:"data"`
	}
	err := c.get("bgp/bgpsession/"+strconv.Itoa(id), &sessions)
	if err != nil {
		return nil, err
	}

	return sessions.Data, nil
}

// GetBGPSessions external method on Client to get BGP sessions
func (c *Client) GetBGPSessions(mbPkgID int) ([]*BGPSession, error) {

	var allSessions BGPSessions

	err := c.get("bgp/bgpsessions", &allSessions)
	if err != nil {
		return nil, err
	}
	if len(allSessions.Data.Sessions) == 0 {
		return nil, nil
	}

	ips, err := c.GetIPs(mbPkgID)
	if err != nil {
		return nil, err
	}
	if len(ips.IPv4) == 0 && len(ips.IPv6) == 0 {
		return nil, err
	}

	ipsMap := *ips.GetIPsMap()

	var sessions []*BGPSession

	for _, session := range allSessions.Data.Sessions {
		_, exists := ipsMap[session.CustomerIP]
		if exists {
			ss, err := c.GetBGPSession(session.ID)
			if err != nil {
				return nil, err
			}
			sessions = append(sessions, ss)
		}
	}

	return sessions, nil
}

type BGPCreateSessionsInput struct {
	MbPkgID   int `json:"mbpkgid"` // Contract BGP ID
	GroupID   int `json:"group_id"`
	Redundant int `json:"redundant"` //Force session redundancy
	IPV6      int `json:"ipv6"`      // IPv6 Session
}

func (c *Client) CreateBGPSessions(mbPkgID int, groupID int, isIPV6 bool, redundant bool) (*BGPSession, error) {
	values := make(url.Values)
	values.Set("mbpkgid", fmt.Sprint(mbPkgID))
	values.Set("group_id", fmt.Sprint(groupID))

	if isIPV6 {
		values.Set("ipv6", "1")
	}
	if redundant {
		values.Set("redundant", "1")
	}

	postData := []byte(values.Encode())
	var err error
	if err != nil {
		return nil, fmt.Errorf("converting data to json: %w", err)
	}

	var sessions struct {
		Data *BGPSession `json:"data"`
	}

	path := "bgp/bgpcreatesessions"
	if err := c.post(path, postData, &sessions); err != nil {
		return nil, fmt.Errorf("posting data: %w", err)
	}

	return sessions.Data, nil
}