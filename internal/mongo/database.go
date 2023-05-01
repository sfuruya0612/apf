package mongo

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const dbName = "aws_price_list"

func Connect(mongoUri string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoUri)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return &mongo.Client{}, err
	}

	return client, nil
}

func Disconnect(client *mongo.Client) error {
	return client.Disconnect(context.TODO())
}

func Collection(client *mongo.Client, collName string) *mongo.Collection {
	return client.Database(dbName).Collection(collName)
}

func DropCollection(coll *mongo.Collection, ctx context.Context) error {
	if err := coll.Drop(ctx); err != nil {
		return err
	}

	log.Printf("Dropped %s collection\n", coll.Name())

	return nil
}

func Find(coll *mongo.Collection, filter interface{}, opt *options.FindOptions) ([]primitive.M, error) {
	cursor, err := coll.Find(context.TODO(), filter, opt)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []bson.M
	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}
