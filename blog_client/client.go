package main


import (
	blogpb "gRPC-blog-service/blogpb"
	context "context"
	fmt "fmt"
	grpc "google.golang.org/grpc"
	log "log"
)

func main(){
	// Debugging Tool
	// display file name and line  number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Starting client ....")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil{
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()
	c := blogpb.NewBlogServiceClient(cc)
	
	// create Blog
	fmt.Println("creating the Blog")
	req := &blogpb.Blog{
		AuthorId: "Stephane",
		Title: "My First Blog",
		Content: "Content of the first blog",
	}
	res, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: req})
	if err != nil{
		log.Fatalf("Unexpected error: %v", err)
	}
	log.Printf("Response from Blog Service: %v", res)
	blogID := res.GetBlog().GetId()

	readRes, readResErr := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: blogID})
	if readResErr != nil{
		log.Fatalf("Unexpected error: %v\n", readResErr)
	}
	fmt.Printf("Blog was read: %v", readRes)
} 