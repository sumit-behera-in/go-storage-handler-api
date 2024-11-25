package controllers

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sumit-behera-in/go-storage-handler/db"
)

type DBController struct {
	dbClient  *db.Clients
	fileLocks sync.Map // Map to store file-specific locks
}

func New(c *db.Clients) *DBController {
	return &DBController{
		dbClient: c,
	}
}

func (dc *DBController) RegisterUserRoutes(rg *gin.RouterGroup) {
	userRoute := rg.Group("/goStorage")
	userRoute.GET("/:fileName", dc.getFile)
	userRoute.POST("/:fileName", dc.postFile)
	userRoute.PATCH("/:fileName", dc.updateFile)
	userRoute.DELETE("/:fileName", dc.deleteFile)
}

func (dc *DBController) getFile(ctx *gin.Context) {
	// Get the file name or identifier from the URL parameter
	fileName := ctx.Param("fileName")
	if fileName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "File name is required"})
		return
	}

	// Acquire a lock for the requested file
	lock := dc.getFileLock(fileName)
	lock.Lock() // Wait until this file is available
	defer lock.Unlock()

	// Fetch the file as []byte from the service
	fileData := dc.dbClient.Download(fileName)
	if fileData.IsEmpty() {
		// Return an error if the file is not found or another issue occurs
		ctx.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("file with fileName : %s not found", fileName)})
		return
	}

	contentType := http.DetectContentType(fileData.File)

	// Set the appropriate Content-Type and return the file data
	ctx.Data(http.StatusOK, contentType, fileData.File)
}

func (dc *DBController) postFile(ctx *gin.Context) {
	// Get the file name from the URL parameter
	fileName := ctx.Param("fileName")
	if fileName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "File name is required"})
		return
	}

	// Acquire a lock for the requested file
	lock := dc.getFileLock(fileName)
	lock.Lock() // Wait until this file is available
	defer lock.Unlock()

	// Parse the file from the request body
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to retrieve the file"})
		return
	}

	// Open the uploaded file
	fileData, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to open the uploaded file"})
		return
	}
	defer fileData.Close()

	// Read the file content into a byte slice
	content, err := io.ReadAll(fileData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to read the file content"})
		return
	}

	data := db.Data{
		FileName: fileName,
		FileType: getFileType(fileName),
		File:     content,
	}

	// Store the file in the database
	err = dc.dbClient.Upload(data, getFileSize(content))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Respond with success
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("File %s uploaded successfully", fileName)})
}

func (dc *DBController) updateFile(ctx *gin.Context) {
	// Get the file name from the URL parameter
	fileName := ctx.Param("fileName")
	if fileName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "File name is required"})
		return
	}

	// Acquire a lock for the requested file
	lock := dc.getFileLock(fileName)
	lock.Lock() // Wait until this file is available
	defer lock.Unlock()

	// Parse the file from the request body
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Failed to retrieve the file"})
		return
	}

	// Open the uploaded file
	fileData, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to open the uploaded file"})
		return
	}
	defer fileData.Close()

	// Read the file content into a byte slice
	content, err := io.ReadAll(fileData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to read the file content"})
		return
	}

	data := db.Data{
		FileName: fileName,
		FileType: getFileType(fileName),
		File:     content,
	}

	// Store the file in the database
	err = dc.dbClient.Update(data, getFileSize(content))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update the file with error : " + err.Error()})
		return
	}

	// Respond with success
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("File %s updated successfully", fileName)})
}

func (dc *DBController) deleteFile(ctx *gin.Context) {
	// Get the file name from the URL parameter
	fileName := ctx.Param("fileName")
	if fileName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "File name is required"})
		return
	}

	err := dc.dbClient.Delete(fileName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete the file with error : " + err.Error()})
		return
	}

	// Respond with success
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("File %s delete successfully", fileName)})
}

// getFileLock ensures per-file locking using sync.Map
func (dc *DBController) getFileLock(fileName string) *sync.Mutex {
	lock, _ := dc.fileLocks.LoadOrStore(fileName, &sync.Mutex{})
	return lock.(*sync.Mutex)
}
