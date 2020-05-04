package billing

import (
	"log"
	"math"
	"time"

	"gopkg.in/mgo.v2/bson"
)

func RoundUp(input float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * input
	round = math.Ceil(digit)
	newVal = round / pow
	return
}

func (r *Repo) CreatePrices(prices *PricesRequest) error {
	id := bson.NewObjectId()
	_, err := r.Coll.UpsertId(id, prices)
	if err != nil {
		return err
	}

	prices.Id = id

	return nil
}

func (r *Repo) Create(billRequest *BillRequest) error {
	id := bson.NewObjectId()
	_, err := r.Coll.UpsertId(id, billRequest)
	if err != nil {
		return err
	}

	billRequest.Id = id

	return nil
}

func (r *Repo) UpdateEmailedBill(id string) error {
	selector := bson.M{"id": bson.ObjectIdHex(id)}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"emailed": true}})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) FindNotEmailedBills() (BillsCollection, error) {
	result := BillsCollection{[]BillRequest{}}
	err := r.Coll.Find(bson.M{"emailed": false}).All(&result.Data)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (r *Repo) FindClientBill(ref string) (BillResource, error) {
	result := BillResource{}
	err := r.Coll.Find(bson.M{"ref": ref}).One(&result.Data)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (r *Repo) UpdatePrices(cartype string, measure string, value float64, minprice float64) error {
	selector := bson.M{"cartype": cartype, "measure": measure}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"price": value, "minprice": minprice}})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) FindPrices(cartype string, measure string) PriceResource {
	result := PriceResource{}
	//r.Coll.Find(bson.M{"cartype":cartype,"measure": measure}).One(&result.Data)
	err := r.Coll.Find(bson.M{"cartype": cartype, "measure": measure}).One(&result.Data)
	if err != nil {
		log.Println(err)
	}
	//log.Println(result)
	return result
}

func (r *Repo) FindDiscountPrice(coupon string) DiscountRequest {
	result := DiscountRequest{}
	r.Coll.Find(bson.M{"coupon": coupon}).One(&result)
	return result
}

type RevenueReport struct {
	//Id string `json:"id" bson:"_id,omitempty"` //You need to keep bson:_id otherwise you wouldn't get any result for driverid
	DriverId string  `json:"driverid"`
	Mobile   string  `json:"mobile"`
	Cartype  string  `json:"cartype"`
	Revenue  float64 `json:"revenue"`
}

type RevenuesReport struct {
	//Id string `json:"id" bson:"_id,omitempty"` //You need to keep bson:_id otherwise you wouldn't get any result for driverid
	DriverId string  `json:"driverid"`
	Mobile   string  `json:"mobile"`
	Cartype  string  `json:"cartype"`
	Revenue  float64 `json:"revenue"`
}

type RevenueReports struct {
	Data []RevenuesReport `json:"data"`
}

func (r *Repo) ShowRevenueReportPerDriver(driverId string, startDate time.Time, endDate time.Time) (RevenueReport, error) {
	o1 := bson.M{
		"$match": bson.M{
			"driverid": driverId,
			"date": bson.M{
				"$gte": startDate,
				"$lte": endDate,
			},
		},
	}

	o2 := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"driverid": "$driverid",
				"mobile":   "$mobile",
				"cartype":  "$cartype",
			},
			"revenue": bson.M{
				"$sum": "$totalprice",
			},
		},
	}

	o3 := bson.M{
		"$project": bson.M{
			"driverid": "$_id.driverid",
			"mobile":   "$_id.mobile",
			"cartype":  "$_id.cartype",
			"revenue":  "$revenue",
		},
	}

	operations := []bson.M{o1, o2, o3}
	// Prepare the query to run in the MongoDB aggregation pipeline
	pipe := r.Coll.Pipe(operations)
	// Run the queries and capture the results
	//results:=[]bson.M{}
	results := RevenueReport{}
	err := pipe.One(&results)
	log.Println(results)
	results.Revenue = RoundUp(results.Revenue, 2)

	return results, err
}

func (r *Repo) ShowAllRevenueReport(startDate time.Time, endDate time.Time) (RevenueReports, error) {
	o1 := bson.M{
		"$match": bson.M{
			"date": bson.M{
				"$gte": startDate,
				"$lte": endDate,
			},
		},
	}

	o2 := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"driverid": "$driverid",
				"mobile":   "$mobile",
				"cartype":  "$cartype",
			},
			"revenue": bson.M{
				"$sum": "$totalprice",
			},
		},
	}

	o3 := bson.M{
		"$project": bson.M{
			"driverid": "$_id.driverid",
			"mobile":   "$_id.mobile",
			"cartype":  "$_id.cartype",
			"revenue":  "$revenue",
		},
	}

	operations := []bson.M{o1, o2, o3}
	// Prepare the query to run in the MongoDB aggregation pipeline
	pipe := r.Coll.Pipe(operations)
	// Run the queries and capture the results
	allResults := RevenueReports{}
	err := pipe.All(&allResults.Data)
	return allResults, err
}
