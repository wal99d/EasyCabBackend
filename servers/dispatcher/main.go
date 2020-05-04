package main

import (
	//3rd Party Library imports

	"gopkg.in/mgo.v2"

	"flag"
	"log"
	"net/http"

	"github.com/wal99d/EasyCabBackend/internals/dispatcher"
)

func main() {
	//Pasing port and monogdb host address
	var port, mdb string
	flag.StringVar(&port, "port", "30005", "dispatcher -port PORT_NUMBER")
	flag.StringVar(&mdb, "mongodb", "localhost", "dispatcher -mongodb MONGODB_HOST")
	flag.Parse()
	log.Println("Dispatcher_App_V1.0_STARTED_On_Port: ", port)
	log.Println("Mongodb_STARTED_On_Host: ", mdb)
	//Creating Mongodb session
	session, err := mgo.Dial("mongodb://" + mdb)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	appC := dispatcher.GetAppCtx(*session, "easycab")

	router := dispatcher.PrepareRoutes(appC)
	http.ListenAndServe(":"+port, router)
}
