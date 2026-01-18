package handler

import (
	"net/http"

	"github.com/DoDuy2004/slack-clone/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService service.FileService
}

func NewFileHandler(fileService service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

func (h *FileHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open file"})
		return
	}
	defer f.Close()

	attachment, err := h.fileService.UploadFile(file.Filename, file.Size, file.Header.Get("Content-Type"), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"attachment_id": attachment.ID,
		"url":           attachment.FileURL,
	})
}
