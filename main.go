package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
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

	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	logger.SetOutput(os.Stdout)
	file, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		logger.Fatal(err)
	}
	defer file.Close()
	logger.SetOutput(file)

}

// func createOrLoadLogFile(){
//     f, err := os.OpenFile("main_err.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
//
//     if err != nil {
//
//         log.Fatalf("Encountered error opening log file: %v", err)
//
//     }
//
//     defer f.Close()
//
//     log.SetOutput(file)
// }

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.URL.Path[0] != '/' {
		// TODO: log that we got an unexpected path - warning
		fmt.Println("log that we got an unexpected path - warning")
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
			fmt.Println("expected not found error - add debug logging")
			http.NotFound(w, r)
			return
		}
		// TODO: no idea what this error is, add error logging
		fmt.Println("not expected ")
		http.Error(w, err.Error(), 500)
		return
	}
	// TODO: add debug logging that we're serving a request
	fmt.Println("serving")
	ranger := objectranger.New(s.project, o, s.bucket)
	httpranger.ServeContent(ctx, w, r, objectKey, o.System.Created, ranger)
}

func main() {
	fmt.Println("expected not found error - add debug logging")
	const (
		access = `15D2da2YnRyWsNuJ4MBqDMh6MpE3EYB1CpvKKz74zUyStHwKqkDWM3eo7aRsUYm3KxwoUZPN6xAcrhifCmW9QHw1XvK5Jb4rHYTBsT2wAzhyitDUHNbvmuuTBvJcFHGGqxVjbdi8P6mAfZiDm5wNHqCUfQDNVRBRTvHcNRqnkwMUQ318GgF7jNgTWaoUrHCBatfd7mBXDtToCfHXs9ftJiwyoqNzowedbtcYLsXQRFvUm2yPsUCeDc1ZoQGxy5b3sUKYu6ETTuhH73ofGD1ttgsK2Sd98Z4ex9PRPqWL1DZQHcCtbSGTr8WTB8X4jwSfmpyooQ3UEswQyokUrdGJfLZ3z`
		bucket = `intern-infra-web`
	)

	os.Setenv("INTERN_WEBSERVER_PORT", "8080")
	// 	os.Setenv("INTERN_PROMETHEUS_PORT", "8080")
	// 	os.Setenv("INTERN_LOG_LEVEL", "INFO")

	webserverPort, err := strconv.Atoi(os.Getenv("INTERN_WEBSERVER_PORT"))
	if err != nil {
		panic(fmt.Sprintf("unable to retrieve webserver port from environment variable INTERN_WEBSERVER_PORT: %v", err))
	}

	ctx := context.Background()

	ag, err := uplink.ParseAccess(access)
	if err != nil {
		panic(err)
	}

	project, err := uplink.OpenProject(ctx, ag)
	if err != nil {
		panic(err)
	}

	s := &Server{
		project: project,
		bucket:  bucket,
	}

	panic(http.ListenAndServe(fmt.Sprintf(":%d", webserverPort), s))
}
