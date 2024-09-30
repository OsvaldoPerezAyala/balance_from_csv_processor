package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type TransactionRow struct {
	Id          string             `bson:"Id"`
	Date        primitive.DateTime `bson:"Date"`
	Transaction float64            `bson:"Transaction"`
	FileName    string             `bson:"FileName"`
}
