package profile

import (
	"gopkg.in/mgo.v2/bson"
)

const FORMAT = "Mon, 2 Jan 2006 15:04:05 GMT"

type User struct {
	Id           bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Mobile       string        `json:"mobile"`
	Country      string        `json:"country"`
	PaymentType  string        `json:"paymenttype"`
	CCType       string        `json:"cctype"`
	CCNumber     string        `json:"ccnumber"`
	CCExpireDate string        `json:"ccexpiredate"`
	CCSecCode    int           `json:"ccseccode"`
	Name         string        `json:"name"`
	Picture      string        `json:"picture"`
	Usertype     string        `json:"usertype"`
	Email        string        `json:"email"`
	DriverId     string        `json:"driverid"`
	Coupon       string        `json:"coupon"`
	Nationality  string        `json:"nationality"`
	CarType      string        `json:"cartype"`
	Password     string        `json:"password"`
	Username     string        `json:"username"`
}

type UsersCollection struct {
	Data []User `json:"data"`
}

type UserResource struct {
	Data User `json:"data"`
}

type Device struct {
	Id          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	DeviceToken string        `json:"deviceToken"`
	RegId       string        `json:"regId"`
	Type        string        `json:"type"`
	ApiKey      string        `json:"apikey"`
	Data        string        `json:"data"`
}

type DevicesCollection struct {
	Data []Device `json:"data"`
}

type DeviceResource struct {
	Data Device `json:"data"`
}

type Message struct {
	MessageCode string
	Content     string
}

type Result struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

type AuthorizedUser struct {
	UserType string `json:"usertype"`
	Token    string `json:"token"`
}

type ClinetList struct {
	Result []User `json:"result"`
	Length int    `json:"length"`
}
