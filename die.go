package cdr

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/CalebQ42/stupid-backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UploadedDie struct {
	Die        map[string]any `json:"die" bson:"die"`
	ID         string         `json:"id" bson:"_id"`
	Expiration int64          `json:"expiration" bson:"expiration"`
}

func (b Backend) UploadDie(req *stupid.Request) bool {
	if len(req.Path) != 1 {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	if req.Body == nil {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	bod, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		req.Resp.WriteHeader(http.StatusInternalServerError)
		log.Println("Error reading body:", err)
		return true
	}
	if len(bod) > 1048576 { //1MB
		req.Resp.WriteHeader(http.StatusRequestEntityTooLarge)
		return true
	}
	//TODO
	return true
}

func (b Backend) GetDie(req *stupid.Request) bool {
	if len(req.Path) != 2 {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	res := b.db.Collection("Dice").FindOne(context.TODO(), bson.M{"_id": req.Path[1]})
	if res.Err() == mongo.ErrNoDocuments {
		req.Resp.WriteHeader(http.StatusNotFound)
		return true
	} else if res.Err() != nil {
		req.Resp.WriteHeader(http.StatusInternalServerError)
		log.Println("Error getting die:", res.Err())
		return true
	}
	var dieGet UploadedDie
	err := res.Decode(&dieGet)
	if err != nil {
		req.Resp.WriteHeader(http.StatusInternalServerError)
		log.Println("Error decoding die:", err)
		return true
	}
	byts, err := json.Marshal(dieGet.Die)
	if err != nil {
		req.Resp.WriteHeader(http.StatusInternalServerError)
		log.Println("Error marshalling die:", err)
		return true
	}
	req.Resp.Header().Set("Content-Type", "application/json")
	req.Resp.Write(byts)
	return true
}
