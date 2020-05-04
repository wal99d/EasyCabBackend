package dispatcher

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"

	"github.com/wal99d/EasyCabBackend/internals/common"
	billingSvc "github.com/wal99d/EasyCabBackend/internals/services/billing"
	dispatchSvc "github.com/wal99d/EasyCabBackend/internals/services/dispatching"
)

// Router

type router struct {
	*httprouter.Router
}

func (r *router) Get(path string, handler http.Handler) {
	r.GET(path, common.WrapHandler(handler))
}

func (r *router) Post(path string, handler http.Handler) {
	r.POST(path, common.WrapHandler(handler))
}

func (r *router) Put(path string, handler http.Handler) {
	r.PUT(path, common.WrapHandler(handler))
}

func (r *router) Delete(path string, handler http.Handler) {
	r.DELETE(path, common.WrapHandler(handler))
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *router) Options(path string, handler http.Handler) {
	r.OPTIONS(path, common.WrapHandler(handler))
}

func PrepareRoutes(appC appContext) *router {
	commonHandlers := alice.New(common.AllowOptions, context.ClearHandler, common.LoggingHandler, common.RecoverHandler, common.AcceptHandler)
	loggingHandlerOnly := alice.New(common.LoggingHandler)
	router := &router{httprouter.New()}
	/* CLINET */
	//the below REST API will be consumed bt client to select cartype and update coupon
	router.Post("/api/v/1/dispatch/client/selectcar", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.SelectCarHandler))
	//the below REST API will be consumed by client to show him/her list of nearby drivers
	router.Post("/api/v/1/dispatch/client/driverlist", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.GetDriverListHandler))
	//the below REST API will be consumed by client once he/she selected the driver from above API list to update only the driver status
	router.Post("/api/v/1/dispatch/client/requestDriver", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.RequestDriverHandler))
	//the below REST API will be consumed by client to show his/her current driver location
	router.Post("/api/v/1/dispatch/client/driverLocation", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.GetDriverLocationHandler))
	//the below REST API will be consumed by client to set pickup location
	router.Post("/api/v/1/dispatch/client/setPickupLocation", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.SetPickupLocationHandler))
	//the below REST API will be consumed by client to cancel the pickup location request
	router.Post("/api/v/1/dispatch/client/cancelPickup", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.CancelPickupHandler))
	//the below REST API will be consumed by client to give him his current billed status
	router.Post("/api/v/1/dispatch/client/status", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.GetStatusHandler))
	//the below REST API will be consumed by client to set his driver's rating
	router.Post("/api/v/1/dispatch/client/ratedriver", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.SetDriverRatingHandler))
	//this for google map to show drivers
	router.Get("/api/v/1/dispatch/client/driversmap", loggingHandlerOnly.ThenFunc(appC.GetDriversMapforGoogle))
	router.Options("/api/v/1/dispatch/client/driversmap", loggingHandlerOnly.ThenFunc(common.PassOptionsHandler))
	//this for google map with callback
	router.Get("/api/v/1/dispatch/client/driversmapcallback", loggingHandlerOnly.ThenFunc(appC.GetDriversMapforGoogleCallback))

	/* DRIVER */
	//the below REST API will be consumed by driver to show him heatmaps
	router.Post("/api/v/1/dispatch/driver/heatmaps", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.GetHeatmapsHandler))
	//the below REST API will be consumed by driver to show him pickup location
	router.Post("/api/v/1/dispatch/driver/getPickupLocation", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.GetPickupLocationHandler))
	//the below REST API will be consumed by driver to record the journey start time
	router.Post("/api/v/1/dispatch/driver/startJourney", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.StartJourneyHandler))
	//the below REST API will be consumed by driver to record the journey finish time
	router.Post("/api/v/1/dispatch/driver/finishJourney", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(billingSvc.BillResource{})).ThenFunc(appC.StopJourneyHandler))
	//the below REST API will be consumed by driver to get his client's details
	router.Post("/api/v/1/dispatch/driver/clientinfo", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.GetClientInfoHandler))
	//the below REST API will be consumed by driver to cancel client's request
	router.Post("/api/v/1/dispatch/driver/cancel", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.CancelClientInfoHandler))
	//the below REST API will be consumed by driver to get his client's location
	router.Post("/api/v/1/dispatch/driver/clientlocation", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.GetClientLocationHandler))
	//the below REST API will be consumed by driver to get his client's location
	router.Post("/api/v/1/dispatch/driver/rateclient", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.GeoResource{})).ThenFunc(appC.SetClientRatingHandler))

	/* OWNER */
	router.Post("/api/v/1/dispatch/clients/list", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.DispatchResource{})).ThenFunc(appC.GetConnectedClients))
	router.Post("/api/v/1/dispatch/drivers/list", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(dispatchSvc.DispatchResource{})).ThenFunc(appC.GetConnectedDrivers))
	router.Get("/api/v/1/dispatch/drivers/count", commonHandlers.Append(common.AuthHandler).ThenFunc(appC.GetConnectedDriversCountHandler))
	router.Options("/api/v/1/dispatch/drivers/count", loggingHandlerOnly.ThenFunc(common.PassOptionsHandler))
	router.Get("/api/v/1/dispatch/clients/count", commonHandlers.Append(common.AuthHandler).ThenFunc(appC.GetConnectedClientsCountHandler))
	router.Options("/api/v/1/dispatch/clients/count", loggingHandlerOnly.ThenFunc(common.PassOptionsHandler))

	return router
}
