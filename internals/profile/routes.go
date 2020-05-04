package profile

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/wal99d/EasyCabBackend/internals/common"
	profileSvc "github.com/wal99d/EasyCabBackend/internals/services/profile"
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

func PrepareRoutes(appC appContext) *router {

	commonHandlers := alice.New(context.ClearHandler, common.LoggingHandler, common.RecoverHandler, common.AcceptHandler)
	loggingHandlerOnly := alice.New(common.LoggingHandler)
	router := &router{httprouter.New()}

	router.Post("/api/v/1/profiles/login/", commonHandlers.Append(common.ContentTypeHandler, common.BodyHandler(profileSvc.UserResource{})).ThenFunc(appC.AuthUserHandler))
	router.Post("/api/v/1/profiles/authUser/", commonHandlers.Append(common.ContentTypeHandler, common.BodyHandler(profileSvc.UserResource{})).ThenFunc(appC.CreateUserTokenHandler))
	router.Post("/api/v/1/profiles/createUser/", commonHandlers.Append(common.ContentTypeHandler, common.BodyHandler(profileSvc.UserResource{})).ThenFunc(appC.CreateUserHandler))
	router.Post("/api/v/1/profiles/updateUser/", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(profileSvc.UserResource{})).ThenFunc(appC.UpdateUserProfileHandler))
	router.Post("/api/v/1/profiles/user/profile/", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(profileSvc.UserResource{})).ThenFunc(appC.ShowUserProfileHandler))
	router.Post("/api/v/1/profiles/addpic/", loggingHandlerOnly.ThenFunc(appC.AddPicHandler))
	router.Get("/api/v/1/profiles/pic/", loggingHandlerOnly.ThenFunc(appC.GetImageHandler))
	router.Get("/api/v/1/profiles/users/", loggingHandlerOnly.ThenFunc(appC.GetAllUserHandler))
	//Control Panel REST
	router.NotFound = http.FileServer(http.Dir("cp"))
	router.Get("/", loggingHandlerOnly.ThenFunc(appC.MainControlPanel))
	router.Post("/api/v/1/profiles/cp/login", commonHandlers.Append(common.ContentTypeHandler, common.BodyHandler(profileSvc.User{})).ThenFunc(appC.ControlPanelHandler))
	router.Delete("/api/v/1/profiles/cp/user/remove/", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(profileSvc.User{})).ThenFunc(appC.RemoveUserHandler))
	router.Put("/api/v/1/profiles/cp/user/update/", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(profileSvc.User{})).ThenFunc(appC.UpdateUserHandler))
	router.Get("/api/v/1/profiles/cp/users", commonHandlers.Append(common.AuthHandler).ThenFunc(appC.GetCpUsersHandler))
	router.Get("/api/v/1/profiles/cp/upsers/profiles", commonHandlers.Append(common.AuthHandler).ThenFunc(appC.GetRegsiteredCientsProfileHandler))
	//Owner Routes
	router.Post("/api/v/1/profiles/clients/list/", commonHandlers.Append(common.ContentTypeHandler, common.AuthHandler, common.BodyHandler(profileSvc.UserResource{})).ThenFunc(appC.GetClinetListHandler))
	router.Get("/api/v/1/profiles/clients/count/", commonHandlers.ThenFunc(appC.GetRegsiteredCientsHandler))
	//Used to keep record of connected devices
	router.Post("/api/v/1/profiles/device/", commonHandlers.Append(common.ContentTypeHandler, common.BodyHandler(profileSvc.DeviceResource{})).ThenFunc(appC.CreateDeviceHandler))
	router.Post("/api/v/1/profiles/device/send2client", commonHandlers.Append(common.ContentTypeHandler, common.BodyHandler(profileSvc.DeviceResource{})).ThenFunc(appC.SendPushMessageToClientsHandler))
	return router
}
