package cdr

import (
	"context"
	"log"
	"time"

	"github.com/CalebQ42/stupid-backend/v2"
	"github.com/CalebQ42/stupid-backend/v2/crash"
	"github.com/CalebQ42/stupid-backend/v2/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Backend struct {
	db *mongo.Database
}

func NewBackend(client *mongo.Client) *Backend {
	go func() {
		for range time.Tick(time.Hour) {
			log.Println("CDR: Deleting expired dice")
			res, err := client.Database("cdr").Collection("profiles").DeleteMany(context.TODO(), bson.M{"expiration": bson.M{"$lt": time.Now().Unix()}})
			if err == mongo.ErrNoDocuments {
				continue
			}
			log.Println("CDR: Deleted", res.DeletedCount, "dice")
		}
	}()
	return &Backend{
		db: client.Database("cdr"),
	}
}

func (b Backend) Logs() db.LogTable {
	return db.NewMongoTable(b.db.Collection("logs"))
}

func (b Backend) Crashes() db.CrashTable {
	return db.NewMongoTable(b.db.Collection("crashes"))
}

func (s Backend) AcceptCrash(cr crash.Individual) bool {
	res := s.db.Collection("versions").FindOne(context.TODO(), bson.M{"version": cr.Version})
	return res.Err() != mongo.ErrNoDocuments
	//TODO: Lookup a list of known "bad" errors that get automatically ignored.
}

func (b Backend) Extension(req *stupid.Request) bool {
	if len(req.Path) < 1 {
		return false
	}
	switch req.Path[0] {
	case "upload":
		return b.UploadDie(req)
	case "die":
		return b.GetDie(req)
	}
	return false
}
