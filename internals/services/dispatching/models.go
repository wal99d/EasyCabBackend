package dispatching

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type DispatchRequest struct {
	Id                 bson.ObjectId    `json:"id" bson:"_id,omitempty"`
	Name               string           `json:"name"`
	Mobile             string           `json:"mobile"`
	Country            string           `json:"country"`
	Picture            string           `json:"picture"`
	Usertype           string           `json:"usertype"`
	Status             string           `json:"status"`
	CarType            string           `json:"cartype"`
	Rating             int              `json:"rating"`
	DriverId           string           `json:"driverid"`
	PickupLocation     *PickupLocations `json:"pickuplocation"`
	AssignedToDriverId string           `json:"assignedtodriverid"`
	Coupon             string           `json:"coupon"`
	JourneyStartTime   string           `json:"JourneyStartTime"`
	JourneyFinishTime  string           `json:"journeyfinishtime"`
	Billed             bool             `json:"billed"`
	Milage             string           `json:"milage"`
	LastBill           string           `json:"lastbill"`
}

type GeoRequest struct {
	Id         bson.ObjectId    `json:"id" bson:"_id,omitempty"`
	Type       string           `json:"type"`
	Geometry   *Locations       `json:"geometry"`
	Properties *DispatchRequest `json:"properties"`
}

type Locations struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type PickupLocations struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type DispatchResource struct {
	Data DispatchRequest `json:"data"`
}

type DispatchCollection struct {
	Data []DispatchRequest `json:"data"`
}

type DispatchGeoCollection struct {
	Data []GeoRequest `json:"features"`
	Type string       `json:"type"`
}

type GeoResource struct {
	Data GeoRequest `json:"data"`
}

type DispatchRepo struct {
	Coll *mgo.Collection
}

type Driver struct {
	Name     string    `json:"name"`
	Mobile   string    `json:"mobile"`
	DriverId string    `json:"driverid"`
	Country  string    `json:"country"`
	CarType  string    `json:"cartype"`
	Coords   []float64 `json:"coords"`
	Rating   int       `json:"rating"`
}

type Heatmap struct {
	Coords []float64 `json:"coords"`
}

type HeatmapList struct {
	Data   []Heatmap `json:"data"`
	Color  string    `json:"color"`
	Status string    `json:"status"`
}

type DriverList struct {
	Data []Driver `json:"data"`
}
