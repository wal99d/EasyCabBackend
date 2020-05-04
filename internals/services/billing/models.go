package billing

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type BillRequest struct {
	Id          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Ref         string        `json:"ref"`
	Mobile      string        `json:"mobile"`
	PaymentType string        `json:"paymenttype"`
	Date        time.Time     `json:"date"`
	DriverId    string        `json:"driverid"`
	Distance    float64       `json:"distance"`
	Measure     string        `json:"measure"`
	Discount    float64       `json:"discount"`
	Emailed     bool          `json:"emailed"`
	Price       float64       `json:"price"`
	Cartype     string        `json:"cartype"`
	Minprice    float64       `json:"min"`
	TotalPrice  float64       `json:"totalprice"`
}

type BillsCollection struct {
	Data []BillRequest `json:"data"`
}

type DiscountRequest struct {
	Id       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Coupon   string        `json:"coupon"`
	Discount float64       `json:"discount"`
}

type Repo struct {
	Coll *mgo.Collection
}

type BillResource struct {
	Data BillRequest `json:"data"`
}

type PricesRequest struct {
	Id       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Cartype  string        `json:"cartype"`
	Measure  string        `json:"measure"`
	Minprice float64       `json:"min"`
	Price    float64       `json:"price"`
}

type PriceResource struct {
	Data PricesRequest `json:"data"`
}

type ReportRequest struct {
	Selection string `json:"selection"`
	DriverId  string `json:"driverid"`
	StartDate string `json:"startdate"`
	EndDate   string `json:"enddate"`
}

type ReportResource struct {
	Data ReportRequest `json:"data"`
}
