package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/arienmalec/alexa-go"
	"github.com/aws/aws-lambda-go/lambda"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Stores a handle to the collection being used by the Lambda function
type Connection struct {
	collection *mongo.Collection
}

// A data structure representation of the collection schema
type Recipe struct {
	ID          primitive.ObjectID `bson:"_id"`
	Name        string             `bson:"name"`
	Ingredients []string           `bson:"ingredients"`
}

func (connection Connection) IntentDispatcher(ctx context.Context, request alexa.Request) (alexa.Response, error) {
	var response alexa.Response
	switch request.Body.Intent.Name {
	case "GetIngredientsForRecipeIntent":
		var recipe Recipe
		recipeName := request.Body.Intent.Slots["recipe"].Value
		if recipeName == "" {
			return alexa.Response{}, errors.New("Recipe name is not present in the request")
		}
		if err := connection.collection.FindOne(ctx, bson.M{"name": recipeName}).Decode(&recipe); err != nil {
			return alexa.Response{}, err
		}
		response = alexa.NewSimpleResponse("Ingredients", strings.Join(recipe.Ingredients, ", "))
	case "GetRecipeFromIngredientsIntent":
		var recipes []Recipe
		ingredient1 := request.Body.Intent.Slots["ingredientone"].Value
		ingredient2 := request.Body.Intent.Slots["ingredienttwo"].Value
		cursor, err := connection.collection.Find(ctx, bson.M{
			"ingredients": bson.D{
				{"$all", bson.A{ingredient1, ingredient2}},
			},
		})
		if err != nil {
			return alexa.Response{}, err
		}
		if err = cursor.All(ctx, &recipes); err != nil {
			return alexa.Response{}, err
		}
		var recipeList []string
		for _, recipe := range recipes {
			recipeList = append(recipeList, recipe.Name)
		}
		response = alexa.NewSimpleResponse("Recipes", strings.Join(recipeList, ", "))
	case "AboutIntent":
		response = alexa.NewSimpleResponse("About", "Created by Nic Raboy in Tracy, CA")
	default:
		response = alexa.NewSimpleResponse("Unknown Request", "The intent was unrecognized")
	}
	return response, nil
}

func main() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("ATLAS_URI")))
	if err != nil {
		panic(err)
	}

	defer client.Disconnect(ctx)

	connection := Connection{
		collection: client.Database("alexa").Collection("recipes"),
	}

	lambda.Start(connection.IntentDispatcher)
}
