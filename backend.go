package cdr

import (
	"context"
	"log"
	"time"

	"github.com/CalebQ42/stupid-backend"
	"github.com/CalebQ42/stupid-backend/pkg/db"
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

func (b Backend) IgnoreOldVersionCrashes() bool {
	return true
}

func (b Backend) CurrentVersions() (out []string) {
	verCol := b.db.Collection("versions")
	cur, err := verCol.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Println("Error getting current version:", err)
		return
	}
	var vers []struct {
		ID  string `bson:"_id"`
		Ver string `bson:"version"`
	}
	err = cur.All(context.TODO(), &vers)
	if err != nil {
		log.Println("Error marshalling current versions:", err)
		return
	}
	for _, v := range vers {
		out = append(out, v.Ver)
	}
	return
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
