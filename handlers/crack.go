package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/hashcatAPI/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type CrackHandler struct {
	handshakeRepo models.HandshakeRepository
	wpaCracker    models.Cracker
}

func NewUploadHandler(repository models.HandshakeRepository, wpaCracker models.Cracker) *CrackHandler {
	return &CrackHandler{repository, wpaCracker}
}

func (h *CrackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.bruteHandshake(w, r)
	return
}

func (h *CrackHandler) bruteHandshake(w http.ResponseWriter, r *http.Request) {
	file, err := receiveFile(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
	}
	defer file.Close()
	log.Println("File recieved: ", file.Name())
	log.Println("Run hashcat with file ", file.Name())
	handshakes, err := h.wpaCracker.CrackWPA(file)
	if err != nil {
		log.Println("crack tool error", err)
		return
	}
	if len(handshakes) == 0 {
		w.Write([]byte("No cracked handshakes"))
		log.Println("No cracked handshakes")
		return
	}
	result, err := json.MarshalIndent(handshakes, "", "  ")
	if err != nil {
		log.Println("Failed marshall response", err)
		return
	}
	fmt.Println(string(result))
	_, err = h.handshakeRepo.Save(handshakes)
	if err != nil {
		log.Println("Failed save handshake", err)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(result)
}

func receiveFile(r *http.Request) (*os.File, error) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		return nil, err
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	uploadedFile, err := ioutil.TempFile("./tempHandshakes", "shake-*.hccapx")
	if err != nil {
		return nil, err
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	uploadedFile.Write(fileBytes)
	return uploadedFile, nil
}
