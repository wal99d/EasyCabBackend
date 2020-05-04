package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/wal99d/EasyCabBackend/internals/biller"
	"gopkg.in/mgo.v2"
)

func main() {
	//Pasing port and monogdb host address
	var port, mdb string
	flag.StringVar(&port, "port", "30006", "biller -port PORT_NUMBER")
	flag.StringVar(&mdb, "mongodb", "localhost", "biller -mongodb MONGODB_HOST")
	flag.Parse()
	log.Println("Biller_App_V1.0_STARTED_On_Port: ", port)
	log.Println("Mongodb_STARTED_On_Host: ", mdb)
	//Creating Mongodb session
	session, err := mgo.Dial("mongodb://" + mdb)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	appC := biller.GetAppCtx(*session, "easycab")
	router := biller.PrepareRoutes(appC)
	http.ListenAndServe(":"+port, router)
}
