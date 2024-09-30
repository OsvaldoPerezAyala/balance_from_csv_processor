package repository

import (
	"balance_from_csv_processor/models"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"sync"
	"time"
)

var clientOptions = GenerateOptsByURL()
var mongoClient = connect(clientOptions)

func GenerateOptsByURL() *options.ClientOptions {
	mongoHost := os.Getenv("MONGO_HOST")
	mongoUser := os.Getenv("MONGO_USER")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:27017", mongoUser, mongoPassword, mongoHost)
	return options.Client().ApplyURI(mongoURI)
}

func connect(clientOptions *options.ClientOptions) *mongo.Client {
	var connectOnce sync.Once
	var session *mongo.Client

	connectOnce.Do(func() {
		session = connectToMongo(clientOptions)
	})

	return session
}

func connectToMongo(clientOptions *options.ClientOptions) *mongo.Client {

	var err error
	session, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Println("Error al abrir conexion", err)
	}

	err = session.Ping(context.TODO(), nil)
	if err != nil {
		fmt.Printf("Failed to connect to MongoDB: %v\n", err)
	}

	return session
}

func ReloadData() {

	elapsed := time.Since(time.Now())
	fmt.Printf("Tiempo de %s : -> [%s]\n", "Reload Data", elapsed)

	fraudConnError := mongoClient.Ping(context.TODO(), nil)

	if fraudConnError != nil {
		log.Println("--SE REQUIERE RECONEXION--")
		mongoClient.Disconnect(context.TODO())
		var err error
		mongoClient, err = mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func SaveData(database string, collection string, doc models.TransactionRow) *mongo.InsertOneResult {

	col := mongoClient.Database(database).Collection(collection)

	result, err := col.InsertOne(context.TODO(), doc)
	if err != nil {
		log.Fatal("Error in DB")
		log.Fatal(err)
		return nil
	}
	if result == nil {
		return nil
	}
	return result
}
