package common

import (
	"encoding/json"
	"net/http"
)

// Errors

type Errors struct {
	Errors []*Error `json:"errors"`
}

type Error struct {
	Id     string `json:"id"`
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func WriteError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(Errors{[]*Error{err}})
}

var (
	ErrBadRequest              = &Error{"bad_request", 400, "Bad request", "Request body is not well-formed. It must be JSON."}
	ErrNotAcceptable           = &Error{"not_acceptable", 406, "Not Acceptable", "Accept header must be set to 'application/vnd.api+json'."}
	ErrUnsupportedMediaType    = &Error{"unsupported_media_type", 415, "Unsupported Media Type", "Content-Type header must be set to: 'application/vnd.api+json'."}
	ErrInternalServer          = &Error{"internal_server_error", 500, "Internal Server Error", "Something went wrong."}
	ErrDriverOffline           = &Error{"Error showing driver", 900, "", "Sorry the driver went offline"}
	ErrNoPickupCoords          = &Error{"Error Pickup Location", 901, "", "Please add pickup location"}
	ErrUnauth                  = &Error{"Error in Authentication", 902, "", "You're not authorized to issue a request"}
	ErrNoClient                = &Error{"Error in Getting Clients", 903, "", "There is no client yet exsits on the system"}
	ErrPricesCouldNotUpdate    = &Error{"Error in updating prices", 904, "", "There is an issue updating the price please contact admin"}
	ErrParseDate               = &Error{"Error in getting Dates", 905, "", "Error Parsing Dates"}
	ErrUserCouponNotUpdated    = &Error{"Error in Updating User's Coupon", 906, "", "There is an error to update user's coupon please contact admin"}
	ErrSettingCartype          = &Error{"Error in Setting Cartype for User", 907, "", "There is an error in setting cartype for user please contact admin"}
	ErrNoDriverOnline          = &Error{"Error in Getting Drivers", 908, "", "There is no driver yet exsits on the system"}
	ErrUpdatingStatus          = &Error{"Error in Updating Status", 909, "", "Sorry there is error in updating the status right now please check the mobile number or consulte the admin"}
	ErrUpdatingClientsDriverId = &Error{"Error in Updating Cleint's Driver Id", 910, "", "Sorry there is an error while updating the client's driver Id"}
	ErrNoResultFound           = &Error{"Error showing revenue reports", 911, "", "Sorry there is no revenue reports yet"}
	ErrNoUserFound             = &Error{"Error no user in the system", 912, "", "Sorry there is no user found"}
	ErrImgNotExists            = &Error{"No_Driver_Img", 913, "There no image for this driver", "Something went wrong."}
	ErrClientExists            = &Error{"Error in Creating Clients", 914, "", "Client already exsits on the system"}
)
