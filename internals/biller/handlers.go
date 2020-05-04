package biller

import (
	"github.com/gorilla/context"
	"gopkg.in/mgo.v2"

	common "github.com/wal99d/EasyCabBackend/internals/common"
	billingSvc "github.com/wal99d/EasyCabBackend/internals/services/billing"
	dispatchSvc "github.com/wal99d/EasyCabBackend/internals/services/dispatching"

	"encoding/json"
	"log"
	"net/http"
	"time"
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

func (c *appContext) CreatePriceHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*billingSvc.PricesRequest)
	repo := billingSvc.Repo{c.db.C("prices")}

	err := repo.CreatePrices(body)
	if err != nil {
		log.Println("repo.Create ", err)
		return
	}

	result := "Prices Created Successfully"
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(result)

}

func (c *appContext) GetBillHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*billingSvc.BillResource)
	repo := billingSvc.Repo{c.db.C("bills")}
	dispathRepo := dispatchSvc.DispatchRepo{c.db.C("requests")}

	//Now we need to find his/her driverId by fetching requests collection
	found, client := dispathRepo.FindUser(body.Data.Mobile)
	if !found {
		common.WriteError(w, common.ErrNoUserFound)
	}
	found, driver := dispathRepo.FindDriverById(client.Data.Properties.AssignedToDriverId)
	if !found {
		common.WriteError(w, common.ErrNoUserFound)
	}
	//we need to find the bill fot this client
	currentBill, err := repo.FindClientBill(client.Data.Properties.LastBill)
	if err != nil {
		common.WriteError(w, common.ErrInternalServer)
		return
	}

	//we need to clear the client's assignedtodriverid and clear his/her driver start and finish time also "billed" to false
	err = dispathRepo.ClearStartAndFinishTime(driver.Data.Properties.Mobile)
	if err != nil {
		log.Println("dispathRepo.ClearStartAndFinishTime ", err)
	}

	err = dispathRepo.ClearAssignedToDriverIdAndStatusAndBilled(body.Data.Mobile)
	if err != nil {
		log.Println("dispathRepo.ClearAssignedToDriverId ", err)
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(currentBill)
}

//Here we will show the revenue report per driver or all drivers filtered by month
func (c *appContext) GetRevenuesHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*billingSvc.ReportResource)
	repo := billingSvc.Repo{c.db.C("bills")}
	p := log.Println
	//we need to get the selection "singleDriver or All" , driverId , startDate and endDate
	selection := body.Data.Selection
	driverId := body.Data.DriverId
	sDate := body.Data.StartDate
	eDate := body.Data.EndDate
	startDate, err := time.Parse("01/02/2006", sDate)
	endDate, err := time.Parse("01/02/2006", eDate)
	p(startDate)
	p(endDate)
	if err != nil {
		common.WriteError(w, common.ErrParseDate)
	} else {

		if selection == "single" {
			//if we select single driver then we need to get driverId, startDate and endDate
			reports, err := repo.ShowRevenueReportPerDriver(driverId, startDate, endDate)
			if err != nil {
				common.WriteError(w, common.ErrNoResultFound)
			} else {
				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(reports)
			}

		} else if selection == "all" {
			//else if we select All then we need to get startDate and endDate
			reports, err := repo.ShowAllRevenueReport(startDate, endDate)
			if err != nil {
				common.WriteError(w, common.ErrNoResultFound)
			} else {
				//loop through all reports and roundup the revenues and store them
				for k, _ := range reports.Data {
					reports.Data[k].Revenue = billingSvc.RoundUp(reports.Data[k].Revenue, 2)
				}
				log.Println(r.Method)
				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(201)
				json.NewEncoder(w).Encode(reports)
			}
		}
	}
}

func (c *appContext) SetPricesHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*billingSvc.PriceResource)
	repo := billingSvc.Repo{c.db.C("prices")}

	err := repo.UpdatePrices(body.Data.Cartype, body.Data.Measure, body.Data.Price, body.Data.Minprice)
	if err != nil {
		common.WriteError(w, common.ErrPricesCouldNotUpdate)
		return
	}
	result := "Priceses Updated Successfully"
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(result)

}
