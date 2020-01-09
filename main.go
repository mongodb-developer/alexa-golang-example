package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/arienmalec/alexa-go"
	"github.com/aws/aws-lambda-go/lambda"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connection struct {
	database *mongo.Database
}

type Recipe struct {
	Id          primitive.ObjectID `bson:"_id"`
	Name        string             `bson:"name"`
	Ingredients []string           `bson:"ingredients"`
}

func (connection Connection) IntentDispatcher(ctx context.Context, request alexa.Request) (alexa.Response, error) {
	var response alexa.Response
	switch request.Body.Intent.Name {
	case "GetIngredientsForRecipeIntent":
		var recipe Recipe
		recipesCollection := connection.database.Collection("recipes")
		recipeName := request.Body.Intent.Slots["recipe"].Value
		if recipeName == "" {
			return alexa.Response{}, errors.New("Recipe name is not present in the request")
		}
		if err := recipesCollection.FindOne(ctx, bson.M{"name": recipeName}).Decode(&recipe); err != nil {
			return alexa.Response{}, err
		}
		response = alexa.NewSimpleResponse("Ingredients", strings.Join(recipe.Ingredients, ", "))
	case "GetRecipeFromIngredientsIntent":
		var recipes []Recipe
		recipesCollection := connection.database.Collection("recipes")
		ingredient1 := request.Body.Intent.Slots["ingredientone"].Value
		ingredient2 := request.Body.Intent.Slots["ingredienttwo"].Value
		cursor, err := recipesCollection.Find(ctx, bson.M{"ingredients": bson.D{{"$all", bson.A{ingredient1, ingredient2}}}})
		if err != nil {
			return alexa.Response{}, err
		}
		if err = cursor.All(ctx, &recipes); err != nil {
			return alexa.Response{}, err
		}
		recipeList := ""
		for _, recipe := range recipes {
			recipeList += recipe.Name
		}
		response = alexa.NewSimpleResponse("Recipes", recipeList)
	case "AboutIntent":
		response = alexa.NewSimpleResponse("About", "Created by Nic Raboy in Tracy, CA")
	default:
		response = alexa.NewSimpleResponse("Unknown Request", "The intent was unrecognized")
	}
	return response, nil
}

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("ATLAS_URI")))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	database := client.Database("alexa")
	connection := Connection{
		database: database,
	}
	lambda.Start(connection.IntentDispatcher)
}
