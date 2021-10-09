package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
}

type Posts struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Caption         string             `json:"caption,omitempty" bson:"caption,omitempty"`
	ImageURL        string             `json:"imageUrl,omitempty" bson:"imageUrl,omitempty"`
	PostedTimestamp string             `json:"postedTimestamp,omitempty" bson:"postedTimestamp,omitempty"`
}

var client *mongo.Client

//connecting to mongoDb
func ConnectMongo() {
	var mongoURL = "mongodb://localhost:27017"
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURL))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		fmt.Println("could not ping to mongo db service: %v\n", err)
		return
	}
	fmt.Println("connected to database:", mongoURL)
}

//Create a user
func CreateUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var user User
	_ = json.NewDecoder(request.Body).Decode(&user)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil || hashedPassword != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	collection := client.Database("InstagramDB").Collection("user")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, user)
	json.NewEncoder(response).Encode(result)
}

// Get the user by id
func GetUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var user User
	collection := client.Database("InstagramDB").Collection("user")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, User{ID: id}).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(user)
}

//Create a Post
func CreatePost(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var post Posts
	_ = json.NewDecoder(request.Body).Decode(&post)
	collection := client.Database("InstagramDB").Collection("post")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, post)
	json.NewEncoder(response).Encode(result)
}

//=Get a post using id
func GetPost(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var post Posts
	collection := client.Database("InstagramDB").Collection("post")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, Posts{ID: id}).Decode(&post)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(post)
}

//List all posts of a user
func GetUserPost(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var post Posts
	collection := client.Database("InstagramDB").Collection("post")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, User{ID: id}).Decode(&post)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(post)
}

func main() {
	fmt.Println("Starting the application...")
	ConnectMongo()
	router := mux.NewRouter()
	router.HandleFunc("/user", CreateUser).Methods("POST")
	router.HandleFunc("/user/{id}", GetUser).Methods("GET")
	router.HandleFunc("/post", CreatePost).Methods("POST")
	router.HandleFunc("/post/{id}", GetPost).Methods("GET")
	router.HandleFunc("/post/user/{id}", GetUserPost).Methods("GET")
	http.ListenAndServe(":12345", router)
}
