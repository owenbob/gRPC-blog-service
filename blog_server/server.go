package main


import (
	blogpb "gRPC-blog-service/blogpb"
	bson "go.mongodb.org/mongo-driver/bson"
	codes "google.golang.org/grpc/codes"
	context "context"
	fmt "fmt"
	grpc "google.golang.org/grpc"
	log "log"
	mongo "go.mongodb.org/mongo-driver/mongo"
	net "net"
	objectid "go.mongodb.org/mongo-driver/bson/primitive"
	options "go.mongodb.org/mongo-driver/mongo/options"
	os "os"
	signal "os/signal"
	status "google.golang.org/grpc/status"
	time "time"

)
var collection *mongo.Collection

type server struct{}

// Data model
type blogItem struct{
	ID 		 objectid.ObjectID  `bson:"_id,omitempty"`
	AuthorID string 			`bson:"author_id"`
	Content  string 			`bson:"content"`
	Title    string 			`bson:"title"`
}

// CreateBlog Function
func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error){
	now := time.Now()
	log.Println("Create Blog request received at: ", now)
	blog := req.GetBlog()

	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title: blog.GetTitle(),
		Content: blog.GetContent(),
	}

	// Insert  blog in db 
	res, err := collection.InsertOne(context.Background(), data)
	if err != nil{
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}
	oid, ok := res.InsertedID.(objectid.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Can not convert to OID"),
		)
	}
	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id: oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title: blog.GetTitle(),
			Content: blog.GetContent(),
		},
	}, nil
}

//ReadBlog Function
func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest)(*blogpb.ReadBlogResponse, error){
	now := time.Now()
	log.Println("Read Blog request received at: ", now)
	blogID := req.GetBlogId()
	oid, err := objectid.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	// create an empty struct
	data := &blogItem{}
	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil{
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %v", err),
		)
	}
	return &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id: data.ID.Hex(),
			AuthorId: data.AuthorID,
			Title: data.Title,
			Content: data.Content,
		}, 
	}, nil

}


func main(){
	// Debugging Tool
	// display file name and line  number
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	

	//Start MongoDb connection
	fmt.Println("Starting MongoDb connection ...")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil { log.Fatalf("Failed to create mongo client: %v", err) }
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil { log.Fatalf("Failed to connect to mongo: %v", err) }

	// Start collection
	collection = client.Database("blogdb").Collection("blog")

	fmt.Println("Blog Service Started")
	//Start listener
	fmt.Println("Starting Listener ...")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil{
		log.Fatalf("Failed to listen: %v", err)
	}
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(s, &server{})

	go func(){
		fmt.Println("Starting Server ...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for Control C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt) 

	// Block until a signal is received
	<-ch
	fmt.Println("Stopping the server ...")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("Closing MongoDB Connection")
	client.Disconnect(ctx)
	fmt.Println("...Blog service succefully shutdown...")
	
}