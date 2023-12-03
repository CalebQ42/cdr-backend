package cdr

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/CalebQ42/stupid-backend/v2"
	"github.com/google/uuid"
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
	if req.Method != http.MethodPost {
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
	var toUpload = UploadedDie{
		Die:        make(map[string]any),
		ID:         uuid.New().String(),
		Expiration: time.Now().Add(12 * time.Hour).Round(time.Hour).Unix(),
	}
	err = json.Unmarshal(bod, &toUpload.Die)
	if err != nil {
		req.Resp.WriteHeader(http.StatusInternalServerError)
		log.Println("Error unmarshalling body:", err)
		return true
	}
	if toUpload.Die["uuid"] != nil {
		delete(toUpload.Die, "uuid")
	}
	_, err = b.db.Collection("dice").InsertOne(context.TODO(), toUpload)
	if err != nil {
		req.Resp.WriteHeader(http.StatusInternalServerError)
		log.Println("Error uploading die:", err)
		return true
	}
	out, err := json.Marshal(map[string]any{"id": toUpload.ID, "expiration": toUpload.Expiration})
	if err != nil {
		req.Resp.WriteHeader(http.StatusInternalServerError)
		log.Println("Error marshalling upload result:", err)
		return true
	}
	req.Resp.WriteHeader(http.StatusCreated)
	req.Resp.Write(out)
	return true
}

func (b Backend) GetDie(req *stupid.Request) bool {
	if len(req.Path) != 2 {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	if req.Method != http.MethodGet {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return true
	}
	res := b.db.Collection("dice").FindOne(context.TODO(), bson.M{"_id": req.Path[1]})
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
	req.Resp.Write(byts)
	return true
}
