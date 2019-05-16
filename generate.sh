# generate command for blog service
protoc blogpb/blog.proto --go_out=plugins=grpc:.