package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client



func main() {
	// Load environment variables
	mongoURI := os.Getenv("MONGODB_URI")
	mongoDatabase := os.Getenv("MONGODB_DATABASE")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	fmt.Println("Connected to MongoDB!")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/prompts", addPrompt(mongoDatabase))

	http.ListenAndServe(":3729", r)
}

func addPrompt(database string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

    type Prompt struct {
            ID         string    `bson:"_id" json:"id"`
            Content    string    `bson:"content" json:"content"`
            Status     string    `bson:"status" json:"status"`
            CreateDate time.Time `bson:"createDate" json:"createDate"`
            UpdateDate time.Time `bson:"updateDate" json:"updateDate"`
    }
        var prompt Prompt
		if err := json.NewDecoder(r.Body).Decode(&prompt); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		collection := client.Database(database).Collection("prompts")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

        res, err := collection.InsertOne(ctx, bson.M{"content": prompt.Content,
                                                    "status": "false",
                                                    "createDate": time.Now(),
		                                            "updateDate": time.Now(),
                                                    } )
		if err != nil {
			http.Error(w, "Failed to insert prompt", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(res)
	}
}

