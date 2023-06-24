## Dockerfile tells docker-compose how to build the image. The Dockerfile is used to build images, while docker-compose helps you run them as containers.


# # base go image (builder is the name of this image)
# FROM golang:1.18-alpine as builder 

# # run a command on the docker image we're building
# RUN mkdir /app

# # copy everything from the current folder (.) into the app folder we created above in our docker image
# COPY . /app

# # Set the working directory
# WORKDIR /app

# # BUild our go code , CGO_ENABLED is a environment variable
# RUN CGO_ENABLED=0 go build -o brokerApp ./cmd/api

# # chmod +x on a file (your script) only means, that you'll make it executable
# RUN chmod +x /app/brokerApp



#### We removed the above lines as we're already building the brokerApp binary from the makefile command build_broker. The above code also does the same thing ,so we avoided repeating.



## Below is a new Docker image ,seperate to the above image
# Build a tiny docker image
FROM alpine:latest

# Run command on new docker image
RUN mkdir /app


# COPY --from=builder /app/brokerApp /app 

# brokerApp binary will be build intially by makefile target build_broker and then this dockerfile will be RUN
COPY brokerApp /app

# When we run this command ,it should first build all of our code on one docker image and then create a much smaller Docker image and copy over the executable (brokerApp) to this new image
CMD ["/app/brokerApp"]