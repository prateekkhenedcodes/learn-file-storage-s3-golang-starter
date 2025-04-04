package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	const maxMemory = 10 << 20
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	// TODO: implement the upload here
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not parse the max memory", err)
		return
	}
	file, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not retrieve the thumbnail", err)
		return
	}
	defer file.Close()
	mediaType := fileHeader.Header.Get("Content-Type")

	imageData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not read the file ", err)
		return
	}
	dbMetaData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get the meta data from the database", err)
		return
	}
	if dbMetaData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "not the owner of the video", fmt.Errorf("not the owner of the video"))
		return
	}

	videoThumbnails[dbMetaData.ID] = thumbnail{
		imageData,
		mediaType,
	}
	thumbnaillink := fmt.Sprintf("http://localhost:%v/api/thumbnails/%v", cfg.port, dbMetaData.ID)

	dbMetaData.UpdatedAt = time.Now()
	dbMetaData.ThumbnailURL = &thumbnaillink
	err = cfg.db.UpdateVideo(dbMetaData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not update the video", err)
		return
	}

	updatedVideoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get the video metadata", err)
		return
	}


	respondWithJSON(w, http.StatusOK, updatedVideoData)
}
