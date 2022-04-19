package main

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"

	"storj.io/common/ranger/httpranger"
	"storj.io/linksharing/objectranger"
	"storj.io/uplink"
)

type Server struct {
	project *uplink.Project
	bucket  string
}

func init() {

	log.SetFormatter(&log.JSONFormatter{})

	// set log file
	file, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// get log level
	logLevel, err := log.ParseLevel(os.Getenv("INTERN_LOG_LEVEL"))
	if err != nil {
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)
	log.SetOutput(file)
	log.SetReportCaller(true)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.URL.Path[0] != '/' {
		// TODO: log that we got an unexpected path - warning
		fmt.Println("- warning  unexpected path ")
		log.Warningln("unexpected path. expected '/' but got ", r.URL.Path[0])
		http.NotFound(w, r)
		return
	}
	objectKey := r.URL.Path[1:]
	if objectKey == "" {
		objectKey = "index.html"
	}
	o, err := s.project.StatObject(ctx, s.bucket, objectKey)
	if err != nil {
		if errors.Is(err, uplink.ErrObjectNotFound) {
			// TODO: expected not found error - add debug logging
			fmt.Println("debug expected not found error  ")
			log.Debugln("expected not found error")
			http.NotFound(w, r)
			return
		}
		// TODO: no idea what this error is, add error logging
		fmt.Println("not expected ")
		log.Errorln("unexpected error")
		http.Error(w, err.Error(), 500)
		return
	}
	// TODO: add debug logging that we're serving a request
	fmt.Println("serving ")
	log.Debugln("serving requesting")
	ranger := objectranger.New(s.project, o, s.bucket)
	httpranger.ServeContent(ctx, w, r, objectKey, o.System.Created, ranger)
}

func main() {

	log.Infoln("Application starting")

	const (
		access = `15D2da2YnRyWsNuJ4MBqDMh6MpE3EYB1CpvKKz74zUyStHwKqkDWM3eo7aRsUYm3KxwoUZPN6xAcrhifCmW9QHw1XvK5Jb4rHYTBsT2wAzhyitDUHNbvmuuTBvJcFHGGqxVjbdi8P6mAfZiDm5wNHqCUfQDNVRBRTvHcNRqnkwMUQ318GgF7jNgTWaoUrHCBatfd7mBXDtToCfHXs9ftJiwyoqNzowedbtcYLsXQRFvUm2yPsUCeDc1ZoQGxy5b3sUKYu6ETTuhH73ofGD1ttgsK2Sd98Z4ex9PRPqWL1DZQHcCtbSGTr8WTB8X4jwSfmpyooQ3UEswQyokUrdGJfLZ3z`
		bucket = `intern-infra-web`
	)

	webserverPort, err := strconv.Atoi(os.Getenv("INTERN_WEBSERVER_PORT"))
	if err != nil {
		log.Panicln("unable to retrieve webserver port from environment variable INTERN_WEBSERVER_PORT:")
		panic(fmt.Sprintf("unable to retrieve webserver port from environment variable INTERN_WEBSERVER_PORT: %v", err))
	}

	ctx := context.Background()

	ag, err := uplink.ParseAccess(access)
	if err != nil {
		log.Panicln("uplink.ParseAccess access error")
		panic(err)
	}

	project, err := uplink.OpenProject(ctx, ag)
	if err != nil {
		log.Panicln("uplink.OpenProject error")
		panic(err)
	}

	s := &Server{
		project: project,
		bucket:  bucket,
	}

	panic(http.ListenAndServe(fmt.Sprintf(":%d", webserverPort), s))

}
