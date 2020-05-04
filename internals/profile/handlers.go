package profile

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	ojson "encoding/json"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/rpc/json"
	profileSvc "github.com/wal99d/EasyCabBackend/internals/services/profile"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	common "github.com/wal99d/EasyCabBackend/internals/common"
)

//Main Handlers
var secretkey = os.Getenv("SECRET")

type appContext struct {
	db *mgo.Database
}

type AppSession struct {
	s *mgo.Session
}

func GetAppCtx(session mgo.Session, appName string) appContext {
	return appContext{
		db: session.DB(appName),
	}
}

func (c *appContext) UpdateUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*profileSvc.UserResource)
	repo := profileSvc.UserRepo{c.db.C("users")}
	result := profileSvc.Result{}
	//Here we need to find if user exsist in our DB.coll
	found, _ := repo.FindUser(body.Data.Mobile)
	if found {
		//then update it
		err := repo.UpdateUser(body.Data.Mobile, body.Data)
		if err != nil {
			panic(err)
		}
		result.Message = "User Profile Updated Successfully"
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		ojson.NewEncoder(w).Encode(result)
	} else {
		results := profileSvc.Message{}
		results.MessageCode = "error02"
		results.Content = "User doesn't exsist!!"
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		ojson.NewEncoder(w).Encode(results)
	}
}

func (c *appContext) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*profileSvc.UserResource)
	repo := profileSvc.UserRepo{c.db.C("users")}
	result := profileSvc.Result{}
	//Here we need to find if user exsist in our DB.coll
	found, _ := repo.FindUser(body.Data.Mobile)
	if !found {
		if body.Data.DriverId != "" {
			//this is a creation of driver coming from owner
			accountInfo := profileSvc.UserResource{}
			accountInfo.Data.Mobile = body.Data.Mobile
			accountInfo.Data.Name = body.Data.Name
			accountInfo.Data.Nationality = body.Data.Nationality
			accountInfo.Data.DriverId = body.Data.DriverId
			accountInfo.Data.Email = body.Data.Email
			accountInfo.Data.Country = "SA"
			accountInfo.Data.Usertype = "driver"
			accountInfo.Data.CarType = body.Data.CarType
			err := repo.Create(&accountInfo.Data)
			if err != nil {
				common.WriteError(w, common.ErrInternalServer)
				return
			}
			//Create Token for the user just created and add it to the json output
			// Declare the expiration time of the token
			// here, we have kept it as 5 minutes
			expirationTime := time.Now().Add(5 * time.Minute)
			// Create the JWT claims, which includes the username and expiry time
			claims := &common.Claims{
				Mobile: body.Data.Mobile,
				StandardClaims: jwt.StandardClaims{
					// In JWT, the expiry time is expressed as unix milliseconds
					ExpiresAt: expirationTime.Unix(),
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
			tokenstring, _ := token.SignedString([]byte(secretkey))

			result.Message = "User Prfile Created Successfully"
			result.Token = tokenstring
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(201)
			ojson.NewEncoder(w).Encode(result)
		} else {
			//This is a client request to signup
			err := repo.Create(&body.Data)
			if err != nil {
				common.WriteError(w, common.ErrInternalServer)
				return
			}
			//Create Token for the user just created and add it to the json output
			// Declare the expiration time of the token
			// here, we have kept it as 5 minutes
			expirationTime := time.Now().Add(5 * time.Minute)
			// Create the JWT claims, which includes the username and expiry time
			claims := &common.Claims{
				Mobile: body.Data.Mobile,
				StandardClaims: jwt.StandardClaims{
					// In JWT, the expiry time is expressed as unix milliseconds
					ExpiresAt: expirationTime.Unix(),
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
			tokenstring, _ := token.SignedString([]byte(secretkey))

			result.Message = "User Prfile Created Successfully"
			result.Token = tokenstring
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(201)
			ojson.NewEncoder(w).Encode(result)
		}
	} else {
		common.WriteError(w, common.ErrClientExists)
		return
	}
}

func (c *appContext) CreateDeviceHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*profileSvc.DeviceResource)
	repo := profileSvc.UserRepo{c.db.C("devices")}
	deviceInfo := profileSvc.DeviceResource{}
	deviceInfo.Data.DeviceToken = body.Data.DeviceToken
	deviceInfo.Data.RegId = body.Data.RegId
	deviceInfo.Data.Type = body.Data.Type

	err := repo.CreateDevice(&deviceInfo.Data)
	if err != nil {
		common.WriteError(w, common.ErrInternalServer)
		return
	} else {
		result := "Device Info Recorded Successfully"
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		ojson.NewEncoder(w).Encode(result)
	}
}

type Args struct {
	Data   string
	ApiKey string
}

func (c *appContext) SendPushMessageToClientsHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*profileSvc.DeviceResource)
	deviceInfo := profileSvc.DeviceResource{}
	deviceInfo.Data.ApiKey = body.Data.ApiKey
	deviceInfo.Data.Data = body.Data.Data

	// provide URL for Control Panel for example http://easycabsa.com/rpc"
	url := os.Getenv("CP_RPC")
	args := &Args{
		Data:   deviceInfo.Data.Data,
		ApiKey: deviceInfo.Data.ApiKey,
	}
	message, err := json.EncodeClientRequest("Send.SendTo", args)
	if err != nil {
		log.Println("%s", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
	if err != nil {
		log.Println("%s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error in sending request to %s. %s", url, err)
	}
	defer resp.Body.Close()

	var result int
	err = json.DecodeClientResponse(resp.Body, &result)
	if err != nil {
		common.WriteError(w, common.ErrInternalServer)
		return
	} else {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		ojson.NewEncoder(w).Encode(result)
	}
}

// Get Token based on Mobile from user
func (c *appContext) CreateUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*profileSvc.UserResource)

	//JWT
	//Create Token for the user just created and add it to the json output
	// Declare the expiration time of the token
	// here, we have kept it as 5 minutes
	expirationTime := time.Now().Add(5 * time.Minute)
	// Create the JWT claims, which includes the username and expiry time
	claims := &common.Claims{
		Mobile: body.Data.Mobile,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	tokenstring, _ := token.SignedString([]byte(secretkey))
	log.Println(tokenstring)
}

func (c *appContext) AddPicHandler(w http.ResponseWriter, r *http.Request) {
	formFile, formHead, err := r.FormFile("pic")
	if err != nil {
		log.Println("fromFile erro:", err)
		common.WriteError(w, common.ErrBadRequest)
		return
	}
	defer formFile.Close()

	//remove any directory names in the filename
	//START: work around IE sending full filepath and manually get filename
	itemHead := formHead.Header["Content-Disposition"][0]
	lookfor := "filename=\""
	fileIndex := strings.Index(itemHead, lookfor)

	if fileIndex < 0 {
		log.Println("fileIndex<0 err:", err)
		common.WriteError(w, common.ErrBadRequest)
		return
	}

	filename := itemHead[fileIndex+len(lookfor):]
	filename = filename[:strings.Index(filename, "\"")]

	slashIndex := strings.LastIndex(filename, "\\")
	if slashIndex > 0 {
		filename = filename[slashIndex+1:]
	}

	slashIndex = strings.LastIndex(filename, "/")
	if slashIndex > 0 {
		filename = filename[slashIndex+1:]
	}
	//END: work around IE sending full filepath

	// GridFs actions
	file, err := c.db.GridFS("pics").Create(filename)
	if err != nil {
		log.Println("Gridfs err:", err)
		common.WriteError(w, common.ErrBadRequest)
		return
	}
	defer file.Close()

	io.Copy(file, formFile)
	if err != nil {
		log.Println("io.Copy err:", err)
		common.WriteError(w, common.ErrBadRequest)
		return
	}

	b := make([]byte, 512)
	formFile.Seek(0, 0)
	formFile.Read(b)

	file.SetContentType(http.DetectContentType(b))
	file.SetMeta(r.Form)
	err = file.Close()

	// json response
	field := "/pics/" + filename

	bytes, _ := ojson.Marshal(map[string]interface{}{
		"error": err,
		"pic":   field,
	})
	w.Write(bytes)
}

func (c *appContext) GetImageHandler(w http.ResponseWriter, r *http.Request) {
	/id := r.FormValue("id")
	query := c.db.C("pics.files").Find(bson.M{"metadata.id": id})
	meta := bson.M{}
	err := query.One(&meta)

	// found file or not
	if err != nil {
		if err.Error() == "not found" {
			common.WriteError(w, common.ErrImgNotExists)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}
		return
	}
	_id := meta["_id"].(bson.ObjectId)

	
	uploadDate := meta["uploadDate"].(time.Time)
	contentType := meta["contentType"].(string)
	fileName := meta["filename"].(string)

	r.ParseForm()
	head := w.Header()
	head.Add("Accept-Ranges", "bytes")
	head.Add("ETag", string(_id)+"+"+r.URL.RawQuery)
	head.Add("Date", uploadDate.Format(profileSvc.FORMAT))
	head.Add("Last-Modified", uploadDate.Format(profileSvc.FORMAT))
	// Expires after ten years :)
	head.Add("Expires", uploadDate.Add(87600*time.Hour).Format(profileSvc.FORMAT))
	head.Add("Cache-Control", "public, max-age=31536000")
	head.Add("Content-Type", contentType)
	if _, dl := r.Form["dl"]; (contentType == "application/octet-stream") || dl {
		head.Add("Content-Disposition", "attachment; filename='"+fileName+"'")
	}

	// already served
	if h := r.Header.Get("If-None-Match"); h == string(_id)+"+"+r.URL.RawQuery {
		w.WriteHeader(http.StatusNotModified)
		w.Write([]byte("304 Not Modified"))
		return
	}

	// get file
	file, err := c.db.GridFS("pics").OpenId(_id)
	defer file.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	io.Copy(w, file)

}

//This function will be consumed by the owner only to show him how many client registered on EasyCab
func (c *appContext) GetClinetListHandler(w http.ResponseWriter, r *http.Request) {
	repo := profileSvc.UserRepo{c.db.C("users")}
	_, count, err := repo.FindAllClients()
	if err != nil {
		common.WriteError(w, common.ErrNoClient)
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	ojson.NewEncoder(w).Encode(count)
}

//This function will check if user is authorize, if so then it return his type "client, driver, or owner"
func (c *appContext) AuthUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*profileSvc.UserResource)
	repo := profileSvc.UserRepo{c.db.C("users")}
	result := profileSvc.AuthorizedUser{}
	//Here we need to find if user exsist in our DB.coll
	found, user := repo.FindUser(body.Data.Mobile)
	log.Println("User: ", user.Data.Mobile)
	if !found {
		common.WriteError(w, common.ErrUnauth)
	} else {
		//Create Token for the user just created and add it to the json output
		// Declare the expiration time of the token
		// here, we have kept it as 5 minutes
		expirationTime := time.Now().Add(5 * time.Minute)
		// Create the JWT claims, which includes the username and expiry time
		claims := &common.Claims{
			Mobile: body.Data.Mobile,
			StandardClaims: jwt.StandardClaims{
				// In JWT, the expiry time is expressed as unix milliseconds
				ExpiresAt: expirationTime.Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		tokenstring, _ := token.SignedString([]byte(secretkey))

		result.UserType = user.Data.Usertype
		result.Token = tokenstring
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		ojson.NewEncoder(w).Encode(result)
	}
}

//This function will show all user profile based on his/her mobile
func (c *appContext) ShowUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*profileSvc.UserResource)
	repo := profileSvc.UserRepo{c.db.C("users")}
	userProfile := profileSvc.UserResource{}

	found, userProfile, _ := repo.FindUserByMobile(body.Data.Mobile)
	if !found {
		common.WriteError(w, common.ErrNoClient)
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	ojson.NewEncoder(w).Encode(userProfile)
}

func (c *appContext) GetAllUserHandler(w http.ResponseWriter, r *http.Request) {
	repo := profileSvc.UserRepo{c.db.C("users")}
	userProfile := profileSvc.UsersCollection{}

	userProfile, err := repo.All()
	if err != nil {
		common.WriteError(w, common.ErrNoClient)
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	ojson.NewEncoder(w).Encode(userProfile)
}

//This function will handle the login to Control Panel
func (c *appContext) ControlPanelHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*profileSvc.User)
	repo := profileSvc.UserRepo{c.db.C("users")}
	//Here we need to find if user exsist in our DB.coll
	log.Println(body.Username)
	log.Println(body.Password)
	found, user := repo.FindUsername(body.Username)
	if !found {
		common.WriteError(w, common.ErrUnauth)
	} else {
		if user.Data.Password != body.Password {
			common.WriteError(w, common.ErrUnauth)
		} else {
			result := profileSvc.AuthorizedUser{}
			//Create Token for the user just created and add it to the json output
			// Declare the expiration time of the token
			// here, we have kept it as 5 minutes
			expirationTime := time.Now().Add(5 * time.Minute)
			// Create the JWT claims, which includes the username and expiry time
			claims := &common.Claims{
				Mobile: body.Mobile,
				StandardClaims: jwt.StandardClaims{
					// In JWT, the expiry time is expressed as unix milliseconds
					ExpiresAt: expirationTime.Unix(),
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
			tokenstring, _ := token.SignedString([]byte(secretkey))

			result.UserType = user.Data.Usertype
			result.Token = tokenstring
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(201)
			ojson.NewEncoder(w).Encode(result)
		}
	}
}

//This function will be consumed by the owner only to show him how many client registered on EasyCab
func (c *appContext) GetRegsiteredCientsHandler(w http.ResponseWriter, r *http.Request) {
	repo := profileSvc.UserRepo{c.db.C("users")}
	_, count, err := repo.FindAllClients()
	if err != nil {
		common.WriteError(w, common.ErrNoClient)
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	ojson.NewEncoder(w).Encode(count)
}

//This function will be consumed by the owner only to show him registered clients profiel's on EasyCab
func (c *appContext) GetRegsiteredCientsProfileHandler(w http.ResponseWriter, r *http.Request) {
	repo := profileSvc.UserRepo{c.db.C("users")}
	profiles, _, err := repo.FindAllClients()
	if err != nil {
		common.WriteError(w, common.ErrNoClient)
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	ojson.NewEncoder(w).Encode(profiles)
}

func (c *appContext) MainControlPanel(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("cp/login.html")
	if err != nil {
		panic(err)
		return
	}
	var cpBody bytes.Buffer
	t.Execute(&cpBody, nil)
	fmt.Fprint(w, cpBody.String())
}

func (c *appContext) RemoveUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*profileSvc.User)
	repo := profileSvc.UserRepo{c.db.C("users")}
	err := repo.RemoveUser(string(body.Id))
	if err != nil {
		common.WriteError(w, common.ErrNoClient)
	} else {
		userProfile := profileSvc.UsersCollection{}

		userProfile, _ = repo.All()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		ojson.NewEncoder(w).Encode(userProfile)
	}
}

func (c *appContext) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*profileSvc.User)
	repo := profileSvc.UserRepo{c.db.C("users")}
	err := repo.UpdateCpUser(string(body.Id), body.Mobile, body.Username, body.Usertype)
	if err != nil {
		common.WriteError(w, common.ErrNoClient)
	} else {
		userProfile := profileSvc.UsersCollection{}

		userProfile, _ = repo.All()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		ojson.NewEncoder(w).Encode(userProfile)
	}
}

func (c *appContext) GetCpUsersHandler(w http.ResponseWriter, r *http.Request) {
	repo := profileSvc.UserRepo{c.db.C("users")}
	userProfile := profileSvc.UsersCollection{}

	userProfile, err := repo.FindAllCpClients()
	if err != nil {
		common.WriteError(w, common.ErrNoClient)
	} else {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(201)
		ojson.NewEncoder(w).Encode(userProfile)
	}
}
