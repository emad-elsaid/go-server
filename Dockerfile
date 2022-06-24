#Build stage
FROM golang:1.17.7-bullseye as build
ENV LANG=C.UTF-8
RUN apt-get update && apt-get install -qq -y postgresql-client
ENV app /app
RUN mkdir -p $app
WORKDIR $app
ADD . $app
RUN go build -o main
# Runtime stage 
FROM golang:1.17.7-bullseye
ENV app /app
RUN mkdir -p $app
WORKDIR $app
COPY --from=build  $app/main ./
CMD ./main
