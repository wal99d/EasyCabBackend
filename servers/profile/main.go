package main

import (
	//3rd Party Library imports

	"gopkg.in/mgo.v2"
	// local project imports

	"github.com/wal99d/EasyCabBackend/internals/profile"
	//GO Stndard library imports

	"flag"

	"log"
	"net/http"
)

func main() {
	//Pasing port and monogdb host address
	var port, mdb string
	flag.StringVar(&port, "port", "30004", "profile -port PORT_NUMBER")
	flag.StringVar(&mdb, "mongodb", "localhost", "profile -mongodb MONGODB_HOST")
	flag.Parse()
	log.Println("Profile_App_V1.0_STARTED_On_Port: ", port)
	log.Println("Mongodb_STARTED_On_Host: ", mdb)
	//Creating Mongodb session
	session, err := mgo.Dial("mongodb://" + mdb)
	if err != nil {
		log.Fatal(err)
	}

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	appC := profile.GetAppCtx(*session, "easycab")
	router := profile.PrepareRoutes(appC)
	http.ListenAndServe(":"+port, router)
}
