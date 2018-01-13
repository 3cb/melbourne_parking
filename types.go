package main

// Spot contains parking bay information
type Spot struct {
	StMarkerID string `json:"st_marker_id"`
	BayID      string `json:"bay_id"`
	Location   struct {
		Latitude      string `json:"latitude"`
		HumanAddress  string `json:"human_address"`
		NeedsRecoding bool   `json:"needs_recoding"`
		Longitude     string `json:"longitude"`
	} `json:"location"`
	Lon    string `json:"lon"`
	Lat    string `json:"lat"`
	Status string `json:"status"`
}
