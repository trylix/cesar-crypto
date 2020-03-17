package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func updateAnswer(answer map[string]interface{}) {
	jsonString, _ := json.Marshal(answer)
	ioutil.WriteFile("answer.json", jsonString, 0644)
}

func generateAnswer(token string) {
	resp, err := http.Get("https://api.codenation.dev/v1/challenge/dev-ps/generate-data?token=" + token)

	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	var answer map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&answer)

	updateAnswer(answer)
}

func decryptAnswer(cifrado string, rate byte) string {
	cifrado_slice := []byte(cifrado)

	decifrado := make([]byte, len(cifrado_slice), len(cifrado_slice))

	for i := 0; i < len(cifrado_slice); i++ {
		if cifrado_slice[i] == ' ' {
			decifrado[i] = ' '
			continue
		}

		match, _ := regexp.MatchString("[^a-z]", string(cifrado_slice[i]))
		if match {
			decifrado[i] = cifrado_slice[i]
			continue
		}

		if cifrado_slice[i] >= 97+rate {
			decifrado[i] = cifrado_slice[i] - rate
		} else {
			decifrado[i] = cifrado_slice[i] + 26 - rate
		}
	}

	return string(decifrado)
}

func generateResume(decifrado string) string {
	algorithm := sha1.New()
	algorithm.Write([]byte(decifrado))

	return hex.EncodeToString(algorithm.Sum(nil))
}

func sendFile(token string) {
	file, err := os.Open("answer.json")

	if err != nil {
		log.Fatalln(err)
	}

	defer file.Close()

	var requestBody bytes.Buffer

	multiPartWriter := multipart.NewWriter(&requestBody)

	fileWritter, err := multiPartWriter.CreateFormFile("answer", "answer.json")
	if err != nil {
		log.Fatalln(err)
	}

	io.Copy(fileWritter, file)

	multiPartWriter.Close()

	req, err := http.NewRequest("POST", "https://api.codenation.dev/v1/challenge/dev-ps/submit-solution?token="+token, &requestBody)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

	client := &http.Client{}

	response, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}

	json.NewDecoder(response.Body).Decode(&result)

	fmt.Println("Seu score foi:", result["score"].(float64))
}

func main() {
	token := "SEU_TOKEN"

	generateAnswer(token)

	file, _ := ioutil.ReadFile("answer.json")

	var answer map[string]interface{}

	err := json.Unmarshal(file, &answer)

	if err != nil {
		log.Fatalln(err)
	}

	cifrado := strings.ToLower(answer["cifrado"].(string))
	rate := byte(answer["numero_casas"].(float64))

	decifrado := decryptAnswer(cifrado, rate)

	answer["decifrado"] = string(decifrado)

	updateAnswer(answer)

	resume := generateResume(decifrado)

	answer["resumo_criptografico"] = resume

	updateAnswer(answer)
	sendFile(token)
}
