package biller

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	common "github.com/wal99d/EasyCabBackend/internals/common"
	billingSvc "github.com/wal99d/EasyCabBackend/internals/services/billing"
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
	loggingHandlerOnly := alice.New(common.AllowOptions, common.LoggingHandler)
	router := &router{httprouter.New()}
	/* CLINET */
	//the below REST API will be consumed by client to show him/her current trip bill
	router.Post("/api/v/1/payment/client/bill", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(billingSvc.BillResource{})).ThenFunc(appC.GetBillHandler))
	router.Post("/api/v/1/payment/client/createprice", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(billingSvc.PricesRequest{})).ThenFunc(appC.CreatePriceHandler))

	/* OWNER */
	//the below REST API will be consumed by the owner to show revenue report per single driver or all drivers
	router.Post("/api/v/1/payment/owner/revenues", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(billingSvc.ReportResource{})).ThenFunc(appC.GetRevenuesHandler))
	//the below REST API will be consumed by the owner to allow him to modify the prices according to KM or MIN
	router.Post("/api/v/1/payment/owner/setPrices", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(billingSvc.PriceResource{})).ThenFunc(appC.SetPricesHandler))
	router.Options("/api/v/1/payment/owner/revenues", loggingHandlerOnly.ThenFunc(common.PassOptionsHandler))

	return router
}
