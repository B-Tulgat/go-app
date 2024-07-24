package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/go-redis/redis/v8"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "github.com/joho/godotenv"
)

var (
    redisClient *redis.Client
    mongoClient *mongo.Client
    ports         = []int{3101, 3102, 3103, 3104, 3105}
    retrylimit    = 5
)

type Prompt struct {
    ID      string `bson:"_id" json:"id"`
    Content string `bson:"content" json:"content"`
    Status  string `bson:"status" json:"status"`
    CreateDate time.Time `bson:"createDate" json:"createDate"`
    UpdateDate time.Time `bson:"updateDate" json:"updateDate"`
}

func init() {
    godotenv.Load()
}

func main() {
    redisClient = redis.NewClient(&redis.Options{
        Addr: "localhost:6379",})
    //Connecting to MongoDB
    mongoURI      := os.Getenv("MONGODB_URI")
    mongoDatabase := os.Getenv("MONGODB_DATABASE")
    mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer mongoCancel()
    var err error
    mongoClient, err = mongo.Connect(mongoCtx, options.Client().ApplyURI(mongoURI))
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)}
    err = mongoClient.Ping(mongoCtx, nil)
    if err != nil {
        log.Fatalf("Failed to ping MongoDB: %v", err)}
    fmt.Println("Connected to MongoDB!")
    defer mongoClient.Disconnect(mongoCtx)
    collection := mongoClient.Database(mongoDatabase).Collection("prompts")

    // Concurrent Redis to Ports call
    go processPrompts(collection)

    for retrylimit >= 0{
        run(collection)
        retrylimit--}

    //after the each retrylimit is used
    fmt.Println("All retries used. Exitting.")
}

func run(collection *mongo.Collection) {
        // Mongo-client context will be refreshed every for loop inside the main where it
        for {
            mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 10*time.Second)
            defer mongoCancel()
            var prompt Prompt
            err := collection.FindOneAndUpdate(mongoCtx, bson.M{"status": "false"}, bson.M{"$set": bson.M{"status": "pending", "updateDate": time.Now()}} ).Decode(&prompt)
            if err == mongo.ErrNoDocuments {
                time.Sleep(10 * time.Second)
                continue
            } else if err != nil {
                fmt.Printf("Error finding and updating document: %v\n", err)
                time.Sleep(10 * time.Second)
                continue
            } else {
                fmt.Println("Changed status to pending for document ID:", prompt.ID)}

            err = redisClient.LPush(context.Background(), "pending_prompts", prompt.ID).Err()
            if err != nil {
                fmt.Printf("Error pushing to Redis: %v\n", err)
            } else {
                fmt.Println("Pushed document ID to Redis:", prompt.ID)}
        }
    }

func processPrompts(collection *mongo.Collection) {
    portIdx := 0
    for {
        redisCtx := context.Background()
        promptID, err := redisClient.BRPop(redisCtx, 0, "pending_prompts").Result()

        if err != nil {
            fmt.Printf("Failed to pop prompt from Redis: %v", err)
            time.Sleep( 2 * time.Second)
        } else{
            mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 10*time.Second)
            defer mongoCancel()
            var prompt Prompt
            objID, errForID := primitive.ObjectIDFromHex(promptID[1])
            if errForID != nil {
                fmt.Printf("invalid ObjectID: %v", err)}

            filter := bson.M{"_id": objID}
            err := collection.FindOneAndUpdate(mongoCtx, filter, bson.M{"$set": bson.M{"status": "done"}}).Decode(&prompt)
            if err != nil {
                fmt.Printf("Failed to find prompt in MongoDB: %v", err)}

            port := ports[portIdx%len(ports)]
            portIdx++
            //sending json of prompt to selected ports
            sendPromptToPort(prompt, port, collection, mongoCtx)}

        time.Sleep(2 * time.Second)
        }
    }

func sendPromptToPort(prompt Prompt, port int, collection *mongo.Collection, ctx context.Context) {
    url := fmt.Sprintf("http://localhost:%d", port)
    payload, _ := json.Marshal(map[string]string{"prompt": prompt.Content, "id":prompt.ID})

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
    if err != nil {
        log.Printf("Failed to create request: %v", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Failed to send request to %d: %v", port, err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK {
        collection.UpdateOne(ctx, bson.M{"_id": prompt.ID}, bson.M{"$set": bson.M{"status": "processed"}})
    }
}
