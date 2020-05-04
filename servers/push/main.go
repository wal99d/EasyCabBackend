package main

import (
	//3rd Party Library imports
	"github.com/alexjlockwood/gcm"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	//GO Stndard library imports
	ojson "encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	//local project imports
	profileSvc "github.com/wal99d/EasyCabBackend/internals/services/profile"
)

type Result int
type Send int
type Args struct {
	Data   string
	ApiKey string
}

var mdb string

func (s *Send) SendTo(r *http.Request, args *Args, res *Result) error {

	// Request a socket connection from the session to process our query.
	// Close the session when the goroutine exits and put the connection back
	// into the pool.
	//sessionCopy := AppSession.MongoSession.Copy()
	session, err := mgo.Dial("mongodb://" + mdb)
	if err != nil {
		log.Fatal(err)
	}

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	devicesRepo := session.DB("easycab").C("devices")
	result := profileSvc.DevicesCollection{[]profileSvc.Device{}}
	err = devicesRepo.Find(bson.M{"type": "client"}).All(&result.Data)
	if err != nil {
		log.Println("Couldn't get any devices from DB!!")
	}

	strArgs := `{"Data":"` + args.Data + `"}`
	byt := []byte(strArgs)
	var data map[string]interface{}
	if err := ojson.Unmarshal(byt, &data); err != nil {
		log.Println(err)
	}

	for k, _ := range result.Data {
		regId := []string{result.Data[k].RegId}
		msg := gcm.NewMessage(data,
			regId...)
		sender := &gcm.Sender{ApiKey: args.ApiKey}
		response, err := sender.Send(msg, 2)
		if err != nil {
			fmt.Println("Failed to send message:", err)
			*res = Result(0)
		} else {
			*res = Result(200)
		}
		fmt.Println(response)
	}

	return nil
}

func main() {

	var port string
	flag.StringVar(&port, "port", "30008", "messanger -port PORT_NUMBER")
	flag.StringVar(&mdb, "mongodb", "localhost", "emailer -mongodb MONGODB_HOST")
	flag.Parse()
	log.Println("PushMessages_App_V1.0_STARTED_On_Port: ", port)

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	send := new(Send)
	s.RegisterService(send, "")
	r := mux.NewRouter()
	r.Handle("/rpc", s)
	http.ListenAndServe(":"+port, r)
}
