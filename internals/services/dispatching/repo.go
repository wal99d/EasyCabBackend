package dispatching

import (
	"log"

	"gopkg.in/mgo.v2/bson"
)

func (r *DispatchRepo) Create(geoRequest *GeoRequest) error {
	id := bson.NewObjectId()
	_, err := r.Coll.UpsertId(id, geoRequest)
	if err != nil {
		return err
	}

	return nil
}

func (r *DispatchRepo) UpdateLastbill(mobile string, billref string) error {
	selector := bson.M{"properties.mobile": mobile}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"properties.lastbill": billref}})
	if err != nil {
		return err
	}

	return nil
}

func (r *DispatchRepo) UpdateCoords(mobile string, mycoords []float64) error {
	selector := bson.M{"properties.mobile": mobile}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"geometry.coordinates": mycoords, "geometry.type": "Point"}})
	if err != nil {
		return err
	}

	return nil
}

func (r *DispatchRepo) UpdatePickupLocation(mobile string, pickuploc []float64) error {
	selector := bson.M{"properties.mobile": mobile}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"properties.pickuplocation.coordinates": pickuploc, "properties.pickuplocation.type": "Point"}})
	if err != nil {
		return err
	}

	return nil
}

func (r *DispatchRepo) CancelPickupLocationRequest(mobile string) error {
	selector := bson.M{"properties.mobile": mobile}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"properties.pickuplocation.coordinates": []float64{}, "properties.pickuplocation.type": "", "properties.assignedtodriverid": "", "properties.status": "online"}})
	if err != nil {
		return err
	}

	return nil
}

func (r *DispatchRepo) UpdateStatus(mobile string, status string) error {
	selector := bson.M{"properties.mobile": mobile}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"properties.status": status}})
	if err != nil {
		return err
	}

	return nil
}

func (r *DispatchRepo) SetCartypeAndClientStatus(mobile string, cartype string) error {
	selector := bson.M{"properties.mobile": mobile}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"properties.cartype": cartype, "properties.status": "online"}})
	if err != nil {
		return err
	}

	return nil
}

func (r *DispatchRepo) UpdateClientAssignedToDriver(mobile string, driverId string) error {
	selector := bson.M{"properties.mobile": mobile}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"properties.assignedtodriverid": driverId}})
	if err != nil {
		return err
	}

	return nil
}

func (r *DispatchRepo) UpdateStartJourney(mobile string, start string, milage string) error {
	selector := bson.M{"properties.mobile": mobile}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"properties.journeystarttime": start, "properties.milage": milage}})
	if err != nil {
		return err
	}

	return nil
}

func (r *DispatchRepo) ClearStartAndFinishTime(mobile string) error {
	selector := bson.M{"properties.mobile": mobile}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"properties.journeystarttime": "", "properties.journeyfinishtime": "", "properties.status": "online"}})
	if err != nil {
		return err
	}

	return nil
}

func (r *DispatchRepo) ClearAssignedToDriverIdAndStatusAndBilled(mobile string) error {
	selector := bson.M{"properties.mobile": mobile}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"properties.assignedtodriverid": "", "properties.status": "online", "properties.billed": false, "properties.lastbill": ""}})
	if err != nil {
		return err
	}

	return nil
}

func (r *DispatchRepo) UpdateFinishJourneyAndBilled(mobile string, finish string) error {
	selector := bson.M{"properties.mobile": mobile}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"properties.journeyfinishtime": finish}})
	if err != nil {
		return err
	}
	//we need to get the client by assignedtodriverid and update his "billed" to true
	//first get driver based on his mobile
	found, driver := r.FindUser(mobile)
	if !found {

	}
	driverId := driver.Data.Properties.DriverId
	//then we get client based on assignedtodriverid
	client := DispatchResource{}
	r.Coll.Find(bson.M{"properties.assignedtodriverid": driverId}).One(&client.Data)

	//then we need to update his/her "billed" to true in order for his/her app to call "getBillHandler" to issue bill
	clientSelector := bson.M{"properties.mobile": client.Data.Mobile}
	r.Coll.Update(clientSelector, bson.M{"$set": bson.M{"properties.billed": true}})

	return nil
}

func (r *DispatchRepo) All() (DispatchCollection, error) {
	result := DispatchCollection{[]DispatchRequest{}}
	err := r.Coll.Find(nil).All(&result.Data)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *DispatchRepo) AllDriversForGoogle() (DispatchGeoCollection, error) {
	result := DispatchGeoCollection{[]GeoRequest{}, "FeatureCollection"}
	err := r.Coll.Find(nil).All(&result.Data)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *DispatchRepo) ShowHeatmaps(mobile string, mycountry string, mycoords []float64, status string) (HeatmapList, error) {
	//We need to update the status of driver to online
	err := r.UpdateStatus(mobile, "online")
	if err != nil {
		log.Println("driver not exists")
	}
	//Here we aggregate over our dispatch requests collection to populate an array of 5 drivers
	DISTANCE_MULTIPLIER := 3963.192          // This is the raduis of Earth in miles
	MaxDistance := (1 / DISTANCE_MULTIPLIER) // We set maxDistance to be almost 200meters to client
	log.Println("MaxDistance: ", MaxDistance)
	o1 := bson.M{
		"$geoNear": bson.M{
			"near": bson.M{
				"type":        "Point",
				"coordinates": []float64{mycoords[0], mycoords[1]},
			},
			"query": bson.M{
				"properties.usertype": "client", //bson.M{"$ne": "client"},
				"properties.status":   "online",
			},
			"limit":              5, //We could use "num" to show only 5 drivers
			"distanceField":      "distance",
			"spherical":          true,
			"distanceMultiplier": DISTANCE_MULTIPLIER,
		},
	}
	o2 := bson.M{
		"$project": bson.M{
			"coords": "$geometry.coordinates",
			"_id":    0,
		},
	}
	operations := []bson.M{o1, o2}
	// Prepare the query to run in the MongoDB aggregation pipeline
	pipe := r.Coll.Pipe(operations)
	// Run the queries and capture the results
	results := HeatmapList{}
	err = pipe.All(&results.Data)
	if err != nil {
		log.Println("pipe.All: ", err)
	}

	numbOfClients := len(results.Data)

	switch {
	case numbOfClients <= 3:
		results.Color = "light"
	case numbOfClients > 3:
		results.Color = "dark"
	}
	results.Status = status

	return results, nil
}

func (r *DispatchRepo) ShowNearbyDrivers(mycountry string, mycoords []float64, carType string) (DriverList, error) {
	//Here we aggregate over our dispatch requests collection to populate an array of 5 drivers
	DISTANCE_MULTIPLIER := 3963.192            // This is the raduis of Earth in miles
	MaxDistance := (0.2 / DISTANCE_MULTIPLIER) // We set maxDistance to be almost 200meters to client
	log.Println("MaxDistance: ", MaxDistance)
	o1 := bson.M{
		"$geoNear": bson.M{
			"near": bson.M{
				"type":        "Point",
				"coordinates": []float64{mycoords[0], mycoords[1]},
			},
			"query": bson.M{
				"properties.usertype": "driver", //bson.M{"$ne": "client"},
				"properties.cartype":  carType,
				"properties.status":   "online",
			},
			"limit":              5, //We could use "num" to show only 5 drivers
			"distanceField":      "distance",
			"spherical":          true,
			"distanceMultiplier": DISTANCE_MULTIPLIER,
		},
	}
	o2 := bson.M{
		"$project": bson.M{
			"name":     "$properties.name",
			"mobile":   "$properties.mobile",
			"driverid": "$properties.driverid",
			"country":  "$properties.country",
			"coords":   "$geometry.coordinates",
			"cartype":  "$properties.cartype",
			"rating":   "$properties.rating",
			"_id":      0,
		},
	}
	operations := []bson.M{o1, o2}
	// Prepare the query to run in the MongoDB aggregation pipeline
	pipe := r.Coll.Pipe(operations)
	// Run the queries and capture the results
	results := DriverList{}
	err := pipe.All(&results.Data)
	if err != nil {
		return DriverList{}, nil
	}

	return results, nil
}

func (r *DispatchRepo) FindAllClientsOnline() (DispatchGeoCollection, int, error) {
	result := DispatchGeoCollection{[]GeoRequest{}, "FeatureCollection"}
	err := r.Coll.Find(bson.M{
		"properties.usertype": "client", "$and": []interface{}{
			bson.M{"properties.status": "online"},
		},
	}).All(&result.Data)
	count := len(result.Data)
	if err != nil {
		return result, count, err
	}

	return result, count, nil
}

func (r *DispatchRepo) FindAllDriversOnline() (DispatchGeoCollection, int, error) {
	result := DispatchGeoCollection{[]GeoRequest{}, "FeatureCollection"}
	err := r.Coll.Find(bson.M{
		"properties.usertype": "driver", "$and": []interface{}{
			bson.M{"properties.status": "online"},
		},
	}).All(&result.Data)
	count := len(result.Data)
	if err != nil {
		return result, count, err
	}

	return result, count, nil
}

func (r *DispatchRepo) Find(id string) (DispatchResource, error) {
	result := DispatchResource{}
	err := r.Coll.FindId(bson.ObjectIdHex(id)).One(&result.Data)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *DispatchRepo) FindUser(m string) (bool, GeoResource) {
	result := GeoResource{}
	count, _ := r.Coll.Find(bson.M{"properties.mobile": m}).Count()
	r.Coll.Find(bson.M{"properties.mobile": m}).One(&result.Data)
	if count > 0 {
		return true, result
	} else {
		return false, result
	}
}

func (r *DispatchRepo) FindDriverById(id string) (bool, GeoResource) {
	result := GeoResource{}
	count, _ := r.Coll.Find(bson.M{"properties.driverid": id}).Count()
	r.Coll.Find(bson.M{"properties.driverid": id}).One(&result.Data)
	if count > 0 {
		return true, result
	} else {
		return false, result
	}
}

func (r *DispatchRepo) FindClientByDriverId(id string) (bool, GeoResource) {
	result := GeoResource{}
	count, _ := r.Coll.Find(bson.M{"properties.assignedtodriverid": id}).Count()
	r.Coll.Find(bson.M{"properties.assignedtodriverid": id}).One(&result.Data)
	if count > 0 {
		return true, result
	} else {
		return false, result
	}
}

func (r *DispatchRepo) Delete(id string) error {
	err := r.Coll.RemoveId(bson.ObjectIdHex(id))
	if err != nil {
		return err
	}

	return nil
}
