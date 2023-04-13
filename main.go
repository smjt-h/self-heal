package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/smjt-h/self-heal/model"
)

func main() {
	// Define the file path and the two sets of keywords to search for
	var logfile string
	flag.StringVar(&logfile, "logfile", "", "path of log file that will be used to parse error and fix")

	// Parse command-line flags
	flag.Parse()
	fixType := os.Getenv("FIX_TYPE")
	// filePath := os.Getenv("FILE_PATH")

	errMsg, fileName := "", ""
	if fixType == "checkstyle" {
		errMsg, fileName = parseCheckstyle(logfile)
	} else if fixType == "compile" {
		errMsg, fileName = parseCompile(logfile)
	} else {
		errMsg, fileName = parseUT(logfile)
	}

	// fileName = "/Users/soumyajitdas/go/src/github.com/smjt-h/selfhealingpipelinetest/src/test/java/sampleUnitTes.java"
	// errMsg = "org.opentest4j.AssertionFailedError: expected: <true> but was: <false>\n\tat sampleUnitTes.t6(sampleUnitTes.java:16)"
	fmt.Println(fileName, errMsg)
	// Read the file at the first extracted string path
	fileBytes2, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Print the contents of the second file
	// fmt.Println("codecontent===", string(fileBytes2), "\nerr==", string(errMsg))
	heal(errMsg, string(fileBytes2), fileName)
}

func heal(errorMsg, content, filename string) {
	verbose := true

	if errorMsg == "" || content == "" {
		log.Fatal("errorMsg or content cannot be empty")
	}

	if verbose {
		fmt.Println("Additional information...")
	}

	code, err := fixCode(content, errorMsg)
	if err != nil {
		log.Fatal(err)
	}

	code = removenewLine(code)
	err = writeToFile(filename, code)
	if err != nil {
		log.Fatal(err)
	}
	// Exit
	os.Exit(0)
}

func fixCode(content, errorMsg string) (string, error) {
	query := "fix the error " + strconv.Quote(errorMsg) + "\n\tand give me the updated code from" + strconv.Quote(content) + "\n### Fixed code"

	fmt.Println("-----------\n\n\n")
	fmt.Println(query, os.Getenv("OPENAI_TOKEN"))

	code, err := openAIRequest(query)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return code.Choices[0].Text, nil
}

func openAIRequest(query string) (*model.CompletionResponse, error) {
	token := os.Getenv("OPENAI_TOKEN")

	url := "https://api.openai.com/v1/completions"
	method := "POST"

	// fmt.Println("query==", query)
	payload := strings.NewReader(`{
	"model": "text-davinci-003",
	"prompt": ` + strconv.Quote(query) + `,
	"temperature": 0,
	"max_tokens": 2000,
	"top_p": 1.0,
	"stop": ["###"],
	"frequency_penalty": 0.0,
	"presence_penalty": 0.0
  }`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var completionResp model.CompletionResponse
	err = json.Unmarshal(body, &completionResp)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	return &completionResp, nil
}

func writeToFile(filename, code string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%s\n", code))
	if err != nil {
		return err
	}

	fmt.Printf("Code successfully written to file '%s'.\n", filename)
	return nil
}

func removenewLine(str string) string {
	//Will remove the first occurance of newline from the input string
	index := strings.Index(str, "\n")
	if index != -1 {
		str = str[index+1:]
	}

	return str
}

func parseCheckstyle(filePath string) (string, string) {
	startKeyword1 := "file name=\""
	endKeyword1 := "\">"
	startKeyword2 := "message="
	endKeyword2 := "/>"

	// Read the file at the specified path
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Convert the file contents to a string
	fileString := string(fileBytes)

	// Find the index of the first set of start and end keywords
	startIndex1 := strings.Index(fileString, startKeyword1)
	endIndex1 := strings.Index(fileString, endKeyword1)

	// Extract the first string between the start and end keywords
	fileName := fileString[startIndex1+len(startKeyword1) : endIndex1]

	// Find the index of the second set of start and end keywords
	startIndex2 := strings.Index(fileString, startKeyword2)
	endIndex2 := strings.Index(fileString, endKeyword2)

	// Extract the second string between the start and end keywords
	errMsg := fileString[startIndex2+len(startKeyword2) : endIndex2]
	return errMsg, fileName
}

func parseUT(filePath string) (string, string) {
	//logic here
	return "errMsg", "fileName"
}

func parseCompile(filePath string) (string, string) {
	//logic here
	errMsg := "fix the compilation error in the file"
	return errMsg, "fileName"
}
