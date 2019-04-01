package main

import (
	"github.com/labstack/echo"

	"github.com/fuku2014/nifcloud-storage-multipart-presigned-upload/backend"
)

func main() {
	e := echo.New()
	e.Static("/", "./frontend/out/")

	api := e.Group("/api")
	{
		api.GET("/create-multipart-upload", backend.CreateMultipartUpload)
		api.GET("/get-upload-url", backend.GetUploadURL)
		api.POST("/complete-multipart-upload", backend.CompleteMultipartUpload)
	}

	e.Logger.Fatal(e.Start(":8080"))
}
