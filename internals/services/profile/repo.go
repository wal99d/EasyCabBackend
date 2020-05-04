package profile

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (r *UserRepo) CreateDevice(device *Device) error {
	id := bson.NewObjectId()
	_, err := r.Coll.UpsertId(id, device)
	if err != nil {
		return err
	}

	device.Id = id

	return nil
}

type UserRepo struct {
	Coll *mgo.Collection
}

func (r *UserRepo) All() (UsersCollection, error) {
	result := UsersCollection{[]User{}}
	err := r.Coll.Find(nil).All(&result.Data)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *UserRepo) FindAllClinetDevices() (DevicesCollection, error) {
	result := DevicesCollection{[]Device{}}
	err := r.Coll.Find(bson.M{"type": "client"}).All(&result.Data)
	return result, err
}

func (r *UserRepo) RemoveUser(id string) error {
	userId := bson.ObjectId(id)
	//fmt.Println(userId)
	err := r.Coll.Remove(bson.M{"_id": userId})
	return err
}

func (r *UserRepo) UpdateCpUser(id string, mobile, username, usertype string) error {
	userId := bson.ObjectId(id)
	//fmt.Println(userId)
	err := r.Coll.Update(bson.M{"_id": userId}, bson.M{"$set": bson.M{"mobile": mobile, "username": username, "usertype": usertype}})
	return err
}

func (r *UserRepo) FindAllClients() (UsersCollection, int, error) {
	result := UsersCollection{[]User{}}
	err := r.Coll.Find(bson.M{"usertype": "client"}).All(&result.Data)
	count := len(result.Data)
	if err != nil {
		return result, count, err
	}

	return result, count, nil
}

func (r *UserRepo) FindAllCpClients() (UsersCollection, error) {
	result := UsersCollection{[]User{}}
	err := r.Coll.Find(bson.M{"usertype": bson.M{"$ne": "client"}}).All(&result.Data)
	return result, err
}

func (r *UserRepo) Find(id string) (UserResource, error) {
	result := UserResource{}
	err := r.Coll.FindId(bson.ObjectIdHex(id)).One(&result.Data)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *UserRepo) FindUser(m string) (bool, UserResource) {
	result := UserResource{}
	count, _ := r.Coll.Find(bson.M{"mobile": m}).Count()
	r.Coll.Find(bson.M{"mobile": m}).One(&result.Data)
	if count > 0 {
		return true, result
	} else {
		return false, result
	}
}

func (r *UserRepo) FindUsername(username string) (bool, UserResource) {
	result := UserResource{}
	count, _ := r.Coll.Find(bson.M{"username": username}).Count()
	r.Coll.Find(bson.M{"username": username}).One(&result.Data)
	if count > 0 {
		return true, result
	} else {
		return false, result
	}
}

func (r *UserRepo) FindUserByMobile(m string) (bool, UserResource, error) {
	result := UserResource{}
	count, _ := r.Coll.Find(bson.M{"mobile": m}).Count()
	err := r.Coll.Find(bson.M{"mobile": m}).One(&result.Data)
	if count > 0 {
		return true, result, nil
	} else {
		return false, result, err
	}
}

func (r *UserRepo) FindDriverUserById(id string) (bool, UserResource) {
	result := UserResource{}
	count, _ := r.Coll.Find(bson.M{"driverid": id}).Count()
	r.Coll.Find(bson.M{"driverid": id}).One(&result.Data)
	if count > 0 {
		return true, result
	} else {
		return false, result
	}
}

func (r *UserRepo) Create(user *User) error {
	id := bson.NewObjectId()
	_, err := r.Coll.UpsertId(id, user)
	if err != nil {
		return err
	}

	user.Id = id

	return nil
}

func (r *UserRepo) UpdateUser(m string, user User) error {
	selector := bson.M{"mobile": m}
	err := r.Coll.Update(selector, &user)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) UpdateUserCoupon(mobile string, coupon string) error {
	selector := bson.M{"mobile": mobile}
	err := r.Coll.Update(selector, bson.M{"$set": bson.M{"coupon": coupon}})
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepo) Update(user *User) error {
	err := r.Coll.UpdateId(user.Id, user)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepo) Delete(id string) error {
	err := r.Coll.RemoveId(bson.ObjectIdHex(id))
	if err != nil {
		return err
	}

	return nil
}

func UserToken(r *http.Request) (string, error) {
	var secretkey = "innovativetech"
	token, err := jwt.Parse(r.PostForm.Get("token"), func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretkey), nil
	})
	if err == nil && token.Valid {
		return token.Raw, nil
	} else {
		return "", err
	}
}
