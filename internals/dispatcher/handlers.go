package dispatcher

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	common "github.com/wal99d/EasyCabBackend/internals/common"
	billingSvc "github.com/wal99d/EasyCabBackend/internals/services/billing"
	dispatchSvc "github.com/wal99d/EasyCabBackend/internals/services/dispatching"
	profileSvc "github.com/wal99d/EasyCabBackend/internals/services/profile"

	"github.com/gorilla/context"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Main Handlers

type appContext struct {
	db *mgo.Database
}

func GetAppCtx(session mgo.Session, appName string) appContext {
	return appContext{
		db: session.DB(appName),
	}
}

// Get the pickup location set by client
func (c *appContext) GetPickupLocationHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	result := dispatchSvc.PickupLocations{}

	ok, driver := repo.FindUser(body.Data.Properties.Mobile)
	if ok == false || driver.Data.Properties.Status == "" {
		//Show json message to tell the client that the driver is offline
		common.WriteError(w, common.ErrDriverOffline)
		return
	} else {
		found, currentClient := repo.FindClientByDriverId(driver.Data.Properties.DriverId)
		if !found {
			common.WriteError(w, common.ErrNoClient)
			return
		} else {
			//otherwise get his pickup location that
			//was set by his/her client
			//fmt.Println("driver.Data.PickupLocation: ",driver.Data.PickupLocation.Coordinates)
			result = *currentClient.Data.Properties.PickupLocation
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(&result)
		}
	}
}

// Get heatmap result as coords and show it to the drivers only
func (c *appContext) GetHeatmapsHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	repoUser := profileSvc.UserRepo{c.db.C("users")}
	result := dispatchSvc.HeatmapList{}
	//Store the request on DB
	//check if the client is already connected to dispatch
	//if so, then just update his coords, otherwise create
	//new document for him
	found, currentUser := repo.FindUser(body.Data.Properties.Mobile)
	if !found {
		//create new document for the user
		userFound, clientUser := repoUser.FindUser(body.Data.Properties.Mobile)
		if !userFound {
			common.WriteError(w, common.ErrNoClient)
			return
		}
		clientreq := dispatchSvc.DispatchRequest{}
		clientreq.Name = clientUser.Data.Name
		clientreq.Mobile = body.Data.Properties.Mobile
		clientreq.CarType = clientUser.Data.CarType
		clientreq.Country = clientUser.Data.Country
		clientreq.Usertype = clientUser.Data.Usertype
		clientreq.Status = "online"

		clientLoc := dispatchSvc.Locations{}
		clientLoc.Type = "Point"
		clientLoc.Coordinates = make([]float64, 2)

		clientPicLoc := dispatchSvc.PickupLocations{}
		clientPicLoc.Type = "Point"
		clientPicLoc.Coordinates = make([]float64, 2)
		clientreq.PickupLocation = &clientPicLoc
		clientreq.DriverId = clientUser.Data.DriverId
		geoReq := dispatchSvc.GeoRequest{}
		geoReq.Properties = &clientreq
		geoReq.Geometry = &clientLoc
		geoReq.Type = "Feature"
		err := repo.Create(&geoReq)
		if err != nil {
			log.Println("repo.Create:", err)
			common.WriteError(w, common.ErrInternalServer)
			return
		}
	} else {
		//otherwise update his locations coordinates
		err := repo.UpdateCoords(currentUser.Data.Properties.Mobile, body.Data.Geometry.Coordinates)
		if err != nil {
			log.Println("repo.UpdateCoords:", err)
			common.WriteError(w, common.ErrInternalServer)
			return
		}
		result, err = repo.ShowHeatmaps(body.Data.Properties.Mobile, body.Data.Properties.Country, body.Data.Geometry.Coordinates, currentUser.Data.Properties.Status)
		if err != nil {
			log.Println("repo.ShowHeatmaps:", err)
			common.WriteError(w, common.ErrInternalServer)
			return
		}
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(result)

}

// Get nearby drivers list of 5 drivers
func (c *appContext) GetDriverListHandler(w http.ResponseWriter, r *http.Request) {
	//Get myloc
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	//Store the request on DB
	//check if the client is already connected to dispatch
	//if so, then just update his coords, otherwise create
	//new document for him
	found, currentUser := repo.FindUser(body.Data.Properties.Mobile)
	if !found {
		//create new document for the user
		//body.Data.Id = bson.NewObjectId()
		err := repo.Create(&body.Data)
		if err != nil {
			common.WriteError(w, common.ErrInternalServer)
			return
		}

		//update his locations coordinates
		//err=repo.UpdateCoords(body.Data.Mobile , body.Data.Location.Coordinates)

		//then find nearby drivers to him/her
		drivers, err := repo.ShowNearbyDrivers(body.Data.Properties.Country, body.Data.Geometry.Coordinates, body.Data.Properties.CarType)
		if err != nil {
			common.WriteError(w, common.ErrNoDriverOnline)
			return
		}
		result := dispatchSvc.DriverList{}
		result.Data = drivers.Data
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(result)
	} else {
		//otherwise update his locations coordinates
		err := repo.UpdateCoords(currentUser.Data.Properties.Mobile, body.Data.Geometry.Coordinates)
		if err != nil {
			common.WriteError(w, common.ErrInternalServer)
			log.Println(err)
			return
		}
		//then find nearby drivers to him/her
		drivers, err := repo.ShowNearbyDrivers(currentUser.Data.Properties.Country, body.Data.Geometry.Coordinates, currentUser.Data.Properties.CarType)
		if err != nil {
			common.WriteError(w, common.ErrNoDriverOnline)
			return
		}
		result := dispatchSvc.DriverList{}
		result.Data = drivers.Data
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(result)
	}
}

// Here the client selected a driver, we should update the status of the driver to "hired" and
// assign the client to driver Id
func (c *appContext) RequestDriverHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	var result string
	//Find if the driver is online then update his status,
	//else show error message to show that the driver is offline
	found, driver := repo.FindDriverById(body.Data.Properties.AssignedToDriverId)
	if !found {
		//Show json message to tell the client that the driver is offline
		common.WriteError(w, common.ErrDriverOffline)
		return
	} else {

		//update client status to be "ontrip"
		err := repo.UpdateStatus(body.Data.Properties.Mobile, "ontrip")
		if err != nil {
			common.WriteError(w, common.ErrUpdatingStatus)
			return
		} else {
			//otherwise update the driver status
			err := repo.UpdateStatus(driver.Data.Properties.Mobile, "hired")
			if err != nil {
				common.WriteError(w, common.ErrUpdatingStatus)
				return
			}
			err = repo.UpdateClientAssignedToDriver(body.Data.Properties.Mobile, driver.Data.Properties.DriverId)
			if err != nil {
				common.WriteError(w, common.ErrUpdatingClientsDriverId)
				return
			} else {
				result = "Driver Status Updated Successfully"

				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(result)
			}
		}
	}
}

// this func will show the client his/her dirver's location
func (c *appContext) GetDriverLocationHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	result := dispatchSvc.GeoResource{}

	//Find if the driver is online then get his location,
	//else show error message to show that the driver is offline
	found, driver := repo.FindUser(body.Data.Properties.Mobile)
	if !found {
		//Show json message to tell the client that the driver is offline
		common.WriteError(w, common.ErrNoDriverOnline)
		return
	} else {
		//otherwise get his location
		result.Data = driver.Data
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(&result)
}

// this func will set the pickup location by the client for driver later on
func (c *appContext) SetPickupLocationHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	var result string
	ok, currentUser := repo.FindUser(body.Data.Properties.Mobile)
	if !ok {
		common.WriteError(w, common.ErrDriverOffline)
		return
	} else {
		if body.Data.Properties.PickupLocation == nil {
			common.WriteError(w, common.ErrNoPickupCoords)
			return
		} else {
			err := repo.UpdatePickupLocation(currentUser.Data.Properties.Mobile, body.Data.Properties.PickupLocation.Coordinates)
			if err != nil {
				common.WriteError(w, common.ErrInternalServer)
				return
			}
			result = "Pickup Location Set Successfully"
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(result)
		}
	}
}

func (c *appContext) CancelPickupHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	var result string

	err := repo.CancelPickupLocationRequest(body.Data.Properties.Mobile)
	if err != nil {
		panic(err)
	}
	result = "Pickup Location Canceled Successfully"
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(result)
}

func (c *appContext) StartJourneyHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	var result string

	//_ , driver :=repo.FindUser(body.Data.Mobile)
	t := time.Now()
	startTime := t.Format(time.RFC3339)
	log.Println(startTime)
	err := repo.UpdateStartJourney(body.Data.Properties.Mobile, startTime, body.Data.Properties.Milage)
	if err != nil {
		common.WriteError(w, common.ErrInternalServer)
		return
	}
	result = "Journey Start Time Recorded Successfully"
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(result)

}

//this func will be consumed by driver in order to stop the journy and record the updated time
//along with "billed" to true in order to info the client through the app to call the "getBillHandler"
//to get his/her bill
func (c *appContext) StopJourneyHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*billingSvc.BillResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	discountRepo := billingSvc.Repo{c.db.C("discount")}
	pricesRepo := billingSvc.Repo{c.db.C("prices")}
	profileSvcepo := profileSvc.UserRepo{c.db.C("users")}
	billRepo := billingSvc.Repo{c.db.C("bills")}

	var result string

	//_ , driver :=repo.FindUser(body.Data.Mobile)
	t := time.Now()
	finishTime := t.Format(time.RFC3339)
	err := repo.UpdateFinishJourneyAndBilled(body.Data.Mobile, finishTime)
	if err != nil {
		common.WriteError(w, common.ErrInternalServer)
		return
	}
	//We need to calculate the bills for the currentClient
	//first we need to find current client by assgined
	//we need to find driverId
	found, driver := repo.FindUser(body.Data.Mobile)
	if !found {
		common.WriteError(w, common.ErrDriverOffline)
		return
	}
	driverId := driver.Data.Properties.DriverId
	//then we need to find client by "AssignedToDriverId" field
	found, currentClient := repo.FindClientByDriverId(driverId)
	if !found {
		common.WriteError(w, common.ErrNoClient)
		return
	}
	//Now we will search for discount coupon by searching on his/her profiles or disptach collections
	found, user := profileSvcepo.FindUser(currentClient.Data.Properties.Mobile)
	if !found {
		common.WriteError(w, common.ErrNoClient)
		return
	}
	bill := billingSvc.BillRequest{}
	bill.Mobile = currentClient.Data.Properties.Mobile
	bill.Date = t
	bill.Cartype = driver.Data.Properties.CarType
	bill.DriverId = driverId
	bill.Distance = body.Data.Distance
	bill.Measure = driver.Data.Properties.Milage
	if user.Data.Coupon != "" {
		//Get his/here discount from his/her profile during registration
		//Get discount from discount coll based on coupon
		discountData := discountRepo.FindDiscountPrice(user.Data.Coupon)
		bill.Discount = discountData.Discount
	} else {
		//otherwise, get it once he/she select type of car
		////Get discount from discount coll based on coupon
		discountData := discountRepo.FindDiscountPrice(currentClient.Data.Properties.Coupon)
		bill.Discount = discountData.Discount
	}
	measure := pricesRepo.FindPrices(driver.Data.Properties.CarType, driver.Data.Properties.Milage)
	measureValue := measure.Data.Price
	measureMinValue := measure.Data.Minprice
	log.Println("measure= ", measure)
	kmPrice := measureValue * body.Data.Distance
	sTime, _ := time.Parse(time.RFC3339, driver.Data.Properties.JourneyStartTime)
	fTime, _ := time.Parse(time.RFC3339, driver.Data.Properties.JourneyFinishTime)
	eTime := fTime.Sub(sTime)
	elaspsedTime := fmt.Sprintf("%.1f", eTime.Minutes())
	oTime, _ := strconv.ParseFloat(elaspsedTime, 64)
	offeredTime := billingSvc.RoundUp(oTime, 2)
	bill.Minprice = measureMinValue
	//mintuesPrice := measureValue * (offeredTime - 5) here we subtract 5 mintues as per easycab request
	mintuesPrice := measureValue * (offeredTime)

	switch {
	case measure.Data.Measure == "km" && kmPrice >= measureMinValue:
		log.Println("kmPrice>= measureMinValue")
		//  we need to claculate by distance travelled in KM
		totalPrice := kmPrice - (kmPrice * bill.Discount)
		bill.Price = kmPrice
		bill.TotalPrice = billingSvc.RoundUp(totalPrice, 2)
		log.Println("bill.TotalPrice= ", totalPrice)

	case measure.Data.Measure == "km" && kmPrice < measureMinValue:
		log.Println("kmPrice< measureMinValue")
		//calculate for minimum charge per KM as per prices collection
		kmPrice = measureMinValue
		totalPrice := kmPrice - (kmPrice * bill.Discount)
		bill.Price = kmPrice
		bill.TotalPrice = billingSvc.RoundUp(totalPrice, 2)

	case measure.Data.Measure == "min" && mintuesPrice >= measureMinValue:
		//get current distance that he travelled
		currentDistance := body.Data.Distance
		//get his avrage speed
		avgSpeed := currentDistance / (offeredTime * 60)
		log.Println("currentSpeed= ", avgSpeed)
		if avgSpeed >= 0.5 {
			//calculate per min
			totalPrice := kmPrice - (kmPrice * bill.Discount)
			bill.Price = kmPrice
			bill.TotalPrice = billingSvc.RoundUp(totalPrice, 2)
		} else {
			//calculate per KM
			if kmPrice >= measureMinValue {
				totalPrice := kmPrice - (kmPrice * bill.Discount)
				bill.Price = kmPrice
				bill.TotalPrice = billingSvc.RoundUp(totalPrice, 2)
			} else {
				kmPrice = measureMinValue
				totalPrice := kmPrice - (kmPrice * bill.Discount)
				bill.Price = kmPrice
				bill.TotalPrice = billingSvc.RoundUp(totalPrice, 2)
			}
		}
	case measure.Data.Measure == "min" && mintuesPrice < measureMinValue:
		log.Println("mintuesPrice< measureMinValue")
		//calculate for minimum charge per Min as per prices collection
		if mintuesPrice <= 0 {
			mintuesPrice := measureMinValue
			totalPrice := mintuesPrice - (mintuesPrice * bill.Discount)
			bill.Price = mintuesPrice
			bill.TotalPrice = billingSvc.RoundUp(totalPrice, 2)
		} else {
			totalPrice := mintuesPrice - (mintuesPrice * bill.Discount)
			bill.Price = mintuesPrice
			bill.TotalPrice = billingSvc.RoundUp(totalPrice, 2)
		}
	}
	bill.Ref = bson.NewObjectId().String()
	err = repo.UpdateLastbill(currentClient.Data.Properties.Mobile, bill.Ref)
	if err != nil {
		common.WriteError(w, common.ErrInternalServer)
		return
	}
	err = billRepo.Create(&bill)
	if err != nil {
		common.WriteError(w, common.ErrInternalServer)
		log.Println(err)
		return
	}

	result = "Journey Finish Time Recorded Successfully"
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(result)
}

//this function will be consumed by the owner to show him list of connected clients
func (c *appContext) GetConnectedClients(w http.ResponseWriter, r *http.Request) {
	//body:= context.Get(r, "body").(*dispatchSvc.DispatchResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	_, count, err := repo.FindAllClientsOnline()
	if err != nil {
		common.WriteError(w, common.ErrNoClient)
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(count)
}

//this function will be consumed by client to select cartype and update coupon if needed
func (c *appContext) SelectCarHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repoUser := profileSvc.UserRepo{c.db.C("users")}
	repoDispatch := dispatchSvc.DispatchRepo{c.db.C("requests")}

	//find user's request first
	found, client := repoDispatch.FindUser(body.Data.Properties.Mobile)
	if !found {
		userFound, clientUser := repoUser.FindUser(body.Data.Properties.Mobile)
		if !userFound {
			common.WriteError(w, common.ErrNoClient)
			return
		}
		//Create new request for him in mongodb
		geoReq := dispatchSvc.GeoRequest{}
		clientreq := dispatchSvc.DispatchRequest{}

		clientreq.Name = clientUser.Data.Name
		clientreq.Mobile = body.Data.Properties.Mobile
		clientreq.CarType = body.Data.Properties.CarType
		clientreq.Country = clientUser.Data.Country
		clientreq.Coupon = body.Data.Properties.Coupon
		clientreq.Usertype = clientUser.Data.Usertype
		clientreq.Status = "online"

		clientLoc := dispatchSvc.Locations{}
		clientLoc.Type = "Point"
		clientLoc.Coordinates = make([]float64, 2)

		clientPicLoc := dispatchSvc.PickupLocations{}
		clientPicLoc.Type = "Point"
		clientPicLoc.Coordinates = make([]float64, 2)
		clientreq.PickupLocation = &clientPicLoc

		geoReq.Geometry = &clientLoc
		geoReq.Properties = &clientreq
		geoReq.Type = "Feature"
		err := repoDispatch.Create(&geoReq)
		if err != nil {
			common.WriteError(w, common.ErrInternalServer)
			return
		}
		err = repoUser.UpdateUserCoupon(body.Data.Properties.Mobile, body.Data.Properties.Coupon)
		if err != nil {
			common.WriteError(w, common.ErrUserCouponNotUpdated)
			return
		}

		err = repoDispatch.SetCartypeAndClientStatus(body.Data.Properties.Mobile, body.Data.Properties.CarType)
		if err != nil {
			common.WriteError(w, common.ErrSettingCartype)
			return
		}

		result := "CarType has been Set Successfully"
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(result)
	} else {
		//client's request exsits in mongodb
		err := repoUser.UpdateUserCoupon(client.Data.Properties.Mobile, body.Data.Properties.Coupon)
		if err != nil {
			common.WriteError(w, common.ErrUserCouponNotUpdated)
			return
		}

		err = repoDispatch.SetCartypeAndClientStatus(client.Data.Properties.Mobile, body.Data.Properties.CarType)
		if err != nil {
			common.WriteError(w, common.ErrSettingCartype)
			return
		}

		result := "CarType has been Set Successfully"
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(result)
	}
}

//this function will show owner list online drivers
func (c *appContext) GetConnectedDrivers(w http.ResponseWriter, r *http.Request) {
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	_, count, err := repo.FindAllDriversOnline()
	if err != nil {
		common.WriteError(w, common.ErrNoDriverOnline)
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(count)
}

//this function will be consumed by driver to get his client's details
func (c *appContext) GetClientInfoHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	repoUser := profileSvc.UserRepo{c.db.C("users")}
	//we need to find driverId
	found, driver := repo.FindUser(body.Data.Properties.Mobile)
	if !found {
		common.WriteError(w, common.ErrDriverOffline)
	}
	driverId := driver.Data.Properties.DriverId
	//then we need to find client request by "AssignedToDriverId" field
	found, clientreq := repo.FindClientByDriverId(driverId)
	//finally, we need to find client user
	found, client, _ := repoUser.FindUserByMobile(clientreq.Data.Properties.Mobile)

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(client)
}

//this function will be consumed by driver to cancel client's request to him
func (c *appContext) CancelClientInfoHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	repoUser := profileSvc.UserRepo{c.db.C("users")}
	//first we need to update driver's status from "hired" to "online" and we clear start and finish time
	found, driver := repo.FindUser(body.Data.Properties.Mobile)
	if !found {
		common.WriteError(w, common.ErrDriverOffline)
	}
	err := repo.ClearStartAndFinishTime(body.Data.Properties.Mobile)
	if err != nil {
		common.WriteError(w, common.ErrDriverOffline)
		return
	}
	//then we need to find client request by "AssignedToDriverId" field
	found, clientreq := repo.FindClientByDriverId(driver.Data.Properties.DriverId)
	//finally, we need to find client user
	found, client, _ := repoUser.FindUserByMobile(clientreq.Data.Properties.Mobile)
	//now we need to update his/her details
	err = repo.ClearStartAndFinishTime(client.Data.Mobile)
	if err != nil {
		common.WriteError(w, common.ErrDriverOffline)
		return
	}
	//now we need to clear "AssignedToDriverId"
	err = repo.ClearAssignedToDriverIdAndStatusAndBilled(client.Data.Mobile)
	if err != nil {
		common.WriteError(w, common.ErrDriverOffline)
	}
	result := "Cleint Request has been Canceled Successfully"
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(result)
}

//this function will be consumed by client to give his current bill status
func (c *appContext) GetStatusHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	result := dispatchSvc.GeoResource{}

	ok, client := repo.FindUser(body.Data.Properties.Mobile)
	if ok == false {
		//Show json message to tell the client that the driver is offline
		common.WriteError(w, common.ErrNoClient)
		return
	} else {
		//otherwise get his location
		result.Data = client.Data
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(result)
	}
}

func (c *appContext) SetDriverRatingHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}

	ok, driver := repo.FindUser(body.Data.Properties.Mobile)
	if !ok {
		common.WriteError(w, common.ErrNoClient)
		return
	} else {
		driver.Data.Properties.Rating = body.Data.Properties.Rating
		result := "Rating updated successfully"
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(result)
	}
}

func (c *appContext) GetClientLocationHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	//we need to find driverId
	found, driver := repo.FindUser(body.Data.Properties.Mobile)
	if !found {
		//create new document for the user
		err := repo.Create(&body.Data)
		if err != nil {
			common.WriteError(w, common.ErrDriverOffline)
			return
		}
	} else {
		//otherwise update his locations coordinates
		err := repo.UpdateCoords(driver.Data.Properties.Mobile, body.Data.Geometry.Coordinates)
		if err != nil {
			common.WriteError(w, common.ErrInternalServer)
			return
		}
	}

	driverId := driver.Data.Properties.DriverId
	//then we need to find client request by "AssignedToDriverId" field
	found, clientreq := repo.FindClientByDriverId(driverId)
	if !found {
		common.WriteError(w, common.ErrNoClient)
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(clientreq)
}

func (c *appContext) SetClientRatingHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*dispatchSvc.GeoResource)
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	//we need to find driverId
	found, driver := repo.FindUser(body.Data.Properties.Mobile)
	if !found {
		common.WriteError(w, common.ErrDriverOffline)
		return
	}
	driverId := driver.Data.Properties.DriverId
	//then we need to find client request by "AssignedToDriverId" field
	found, clientreq := repo.FindClientByDriverId(driverId)
	if !found {
		common.WriteError(w, common.ErrNoClient)
		return
	} else {
		//update the client rating as per driver request
		clientreq.Data.Properties.Rating = body.Data.Properties.Rating
		result := "Client Rating Updated Successfully"
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(result)
	}

}

func (c *appContext) GetDriversMapforGoogleCallback(w http.ResponseWriter, r *http.Request) {
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	//we need to show all drivers
	result, _, _ := repo.FindAllDriversOnline()

	w.Header().Set("Content-Type", "application/vnd.api+json")
	b, _ := json.Marshal(&result)
	fmt.Fprintf(w, "%s(%s)", "data_callback", string(b))

}

func (c *appContext) GetDriversMapforGoogle(w http.ResponseWriter, r *http.Request) {
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	//we need to show all drivers
	result, _, _ := repo.FindAllDriversOnline()

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(result)

}

//These functions are for Control Panel
func (c *appContext) GetConnectedDriversCountHandler(w http.ResponseWriter, r *http.Request) {
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	_, count, err := repo.FindAllDriversOnline()
	if err != nil {
		common.WriteError(w, common.ErrNoDriverOnline)
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "authorization , content-type, accept , user-agent")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(count)
}

func (c *appContext) GetConnectedClientsCountHandler(w http.ResponseWriter, r *http.Request) {
	repo := dispatchSvc.DispatchRepo{c.db.C("requests")}
	_, count, err := repo.FindAllClientsOnline()
	if err != nil {
		common.WriteError(w, common.ErrNoClient)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "authorization , content-type, accept , user-agent")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(count)
}
